package adf

import (
	"strings"
)

// ExtractText walks an ADF document (as decoded generic JSON — map[string]any /
// []any) and returns a readable plain-text rendering. Used to display issue
// descriptions and comments in table/text output. It is lossy by design.
func ExtractText(v any) string {
	var sb strings.Builder
	walk(v, &sb)
	return strings.TrimSpace(collapseBlankLines(sb.String()))
}

func walk(v any, sb *strings.Builder) {
	switch n := v.(type) {
	case map[string]any:
		typ, _ := n["type"].(string)
		switch typ {
		case "text":
			if s, ok := n["text"].(string); ok {
				sb.WriteString(s)
			}
		case "hardBreak":
			sb.WriteString("\n")
		case "paragraph", "heading", "blockquote":
			walkContent(n["content"], sb)
			sb.WriteString("\n\n")
		case "listItem":
			sb.WriteString("- ")
			walkContent(n["content"], sb)
		case "codeBlock":
			walkContent(n["content"], sb)
			sb.WriteString("\n\n")
		case "rule":
			sb.WriteString("\n---\n")
		default:
			// doc, bulletList, orderedList, and anything else: recurse into content
			walkContent(n["content"], sb)
		}
	case []any:
		for _, item := range n {
			walk(item, sb)
		}
	}
}

func walkContent(content any, sb *strings.Builder) {
	if arr, ok := content.([]any); ok {
		for _, item := range arr {
			walk(item, sb)
		}
	}
}

func collapseBlankLines(s string) string {
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return s
}
