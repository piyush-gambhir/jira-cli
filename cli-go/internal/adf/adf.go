// Package adf builds Atlassian Document Format (ADF) documents. In the Jira
// Cloud v3 REST API, rich-text fields (issue description, comment body, worklog
// comment, textarea custom fields) must be ADF JSON documents, not plain strings.
//
// See jira/references/agile-adf-conventions.md §3 for the spec this implements.
package adf

import (
	"regexp"
	"strings"
)

// Mark is inline styling applied to a text node (strong, em, code, link, …).
type Mark struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

// Node is any ADF node: a block (paragraph, heading, list, …) or inline (text, hardBreak).
type Node struct {
	Type    string         `json:"type"`
	Attrs   map[string]any `json:"attrs,omitempty"`
	Content []Node         `json:"content,omitempty"`
	Marks   []Mark         `json:"marks,omitempty"`
	Text    string         `json:"text,omitempty"`
}

// Doc is the root ADF document. Marshals directly to the JSON the API expects.
type Doc struct {
	Version int    `json:"version"`
	Type    string `json:"type"`
	Content []Node `json:"content"`
}

// newDoc wraps top-level block nodes in a doc root, guaranteeing at least one block.
func newDoc(content []Node) Doc {
	if len(content) == 0 {
		content = []Node{paragraph(nil)}
	}
	return Doc{Version: 1, Type: "doc", Content: content}
}

// Text returns a plain text inline node (empty string yields no node — callers filter).
func text(s string, marks ...Mark) Node {
	n := Node{Type: "text", Text: s}
	if len(marks) > 0 {
		n.Marks = marks
	}
	return n
}

func paragraph(children []Node) Node {
	return Node{Type: "paragraph", Content: children}
}

func heading(level int, children []Node) Node {
	return Node{Type: "heading", Attrs: map[string]any{"level": level}, Content: children}
}

func listItem(children []Node) Node {
	return Node{Type: "listItem", Content: []Node{paragraph(children)}}
}

func codeBlock(code, lang string) Node {
	attrs := map[string]any{}
	if lang != "" {
		attrs["language"] = lang
	}
	n := Node{Type: "codeBlock", Content: []Node{{Type: "text", Text: code}}}
	if len(attrs) > 0 {
		n.Attrs = attrs
	}
	return n
}

// FromPlainText converts a plain-text string to ADF: blank lines separate
// paragraphs; single newlines become hardBreaks within a paragraph.
func FromPlainText(s string) Doc {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	blocks := splitBlankLines(s)
	var content []Node
	for _, b := range blocks {
		content = append(content, paragraphFromLines(strings.Split(b, "\n")))
	}
	return newDoc(content)
}

// FromMarkdown converts a lightweight-markdown string to ADF. Supports headings
// (#..######), bullet/ordered lists, fenced code blocks, blockquotes, and the
// inline marks bold/italic/code/link. Anything unrecognized is treated as text.
func FromMarkdown(s string) Doc {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	var content []Node
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		switch {
		case trimmed == "":
			i++
		case strings.HasPrefix(trimmed, "```"):
			lang := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			var code []string
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				code = append(code, lines[i])
				i++
			}
			if i < len(lines) {
				i++ // consume closing fence
			}
			content = append(content, codeBlock(strings.Join(code, "\n"), lang))
		case headingLevel(trimmed) > 0:
			lvl := headingLevel(trimmed)
			content = append(content, heading(lvl, parseInline(strings.TrimSpace(trimmed[lvl:]))))
			i++
		case isBullet(trimmed):
			var items []Node
			for i < len(lines) && isBullet(strings.TrimSpace(lines[i])) {
				items = append(items, listItem(parseInline(bulletText(strings.TrimSpace(lines[i])))))
				i++
			}
			content = append(content, Node{Type: "bulletList", Content: items})
		case isOrdered(trimmed):
			var items []Node
			for i < len(lines) && isOrdered(strings.TrimSpace(lines[i])) {
				items = append(items, listItem(parseInline(orderedText(strings.TrimSpace(lines[i])))))
				i++
			}
			content = append(content, Node{Type: "orderedList", Content: items})
		case strings.HasPrefix(trimmed, ">"):
			var quote []string
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), ">") {
				q := strings.TrimSpace(lines[i])
				q = strings.TrimPrefix(strings.TrimPrefix(q, ">"), " ")
				quote = append(quote, q)
				i++
			}
			content = append(content, Node{Type: "blockquote", Content: []Node{paragraph(parseInline(strings.Join(quote, " ")))}})
		default:
			var para []string
			for i < len(lines) {
				t := strings.TrimSpace(lines[i])
				if t == "" || strings.HasPrefix(t, "```") || headingLevel(t) > 0 || isBullet(t) || isOrdered(t) || strings.HasPrefix(t, ">") {
					break
				}
				para = append(para, lines[i])
				i++
			}
			content = append(content, paragraphFromLines(para))
		}
	}
	return newDoc(content)
}

// paragraphFromLines builds one paragraph from lines, joining them with hardBreaks
// and parsing inline marks in each line.
func paragraphFromLines(lines []string) Node {
	var children []Node
	for idx, ln := range lines {
		if idx > 0 {
			children = append(children, Node{Type: "hardBreak"})
		}
		children = append(children, parseInline(ln)...)
	}
	return paragraph(children)
}

func splitBlankLines(s string) []string {
	re := regexp.MustCompile(`\n[ \t]*\n+`)
	parts := re.Split(strings.Trim(s, "\n"), -1)
	var out []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return out
}

func headingLevel(line string) int {
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	if n >= 1 && n <= 6 && n < len(line) && line[n] == ' ' {
		return n
	}
	return 0
}

func isBullet(line string) bool {
	return strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ")
}
func bulletText(line string) string { return strings.TrimSpace(line[2:]) }

var orderedRe = regexp.MustCompile(`^\d+[.)]\s+`)

func isOrdered(line string) bool { return orderedRe.MatchString(line) }
func orderedText(line string) string {
	return strings.TrimSpace(orderedRe.ReplaceAllString(line, ""))
}

var linkRe = regexp.MustCompile(`^\[([^\]]*)\]\(([^)]+)\)`)

// parseInline parses a single line into inline text nodes with marks. Non-nested,
// deterministic: handles `code`, [text](url), **bold**/__bold__, *italic*/_italic_.
func parseInline(s string) []Node {
	var nodes []Node
	var buf strings.Builder
	flush := func() {
		if buf.Len() > 0 {
			nodes = append(nodes, text(buf.String()))
			buf.Reset()
		}
	}
	i := 0
	for i < len(s) {
		switch {
		case s[i] == '`':
			if end := strings.IndexByte(s[i+1:], '`'); end >= 0 {
				flush()
				if inner := s[i+1 : i+1+end]; inner != "" {
					nodes = append(nodes, text(inner, Mark{Type: "code"}))
				}
				i += end + 2
				continue
			}
		case s[i] == '[':
			if m := linkRe.FindStringSubmatch(s[i:]); m != nil {
				flush()
				label := m[1]
				if label == "" {
					label = m[2]
				}
				nodes = append(nodes, text(label, Mark{Type: "link", Attrs: map[string]any{"href": m[2]}}))
				i += len(m[0])
				continue
			}
		case strings.HasPrefix(s[i:], "**") || strings.HasPrefix(s[i:], "__"):
			delim := s[i : i+2]
			if end := strings.Index(s[i+2:], delim); end >= 0 {
				flush()
				if inner := s[i+2 : i+2+end]; inner != "" {
					nodes = append(nodes, text(inner, Mark{Type: "strong"}))
				}
				i += end + 4
				continue
			}
		case s[i] == '*' || s[i] == '_':
			delim := s[i]
			if end := strings.IndexByte(s[i+1:], delim); end >= 0 {
				flush()
				if inner := s[i+1 : i+1+end]; inner != "" {
					nodes = append(nodes, text(inner, Mark{Type: "em"}))
				}
				i += end + 2
				continue
			}
		}
		buf.WriteByte(s[i])
		i++
	}
	flush()
	if nodes == nil {
		nodes = []Node{}
	}
	return nodes
}
