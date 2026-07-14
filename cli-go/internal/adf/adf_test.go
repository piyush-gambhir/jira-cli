package adf

import (
	"encoding/json"
	"strings"
	"testing"
)

func toJSON(t *testing.T, d Doc) string {
	t.Helper()
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

func TestFromPlainText_SingleParagraph(t *testing.T) {
	js := toJSON(t, FromPlainText("hello world"))
	if !strings.Contains(js, `"type":"doc"`) || !strings.Contains(js, `"version":1`) {
		t.Fatalf("missing doc root: %s", js)
	}
	if !strings.Contains(js, `"type":"paragraph"`) || !strings.Contains(js, `"text":"hello world"`) {
		t.Fatalf("missing paragraph/text: %s", js)
	}
}

func TestFromPlainText_ParagraphsAndHardBreak(t *testing.T) {
	d := FromPlainText("line1\nline2\n\npara2")
	if len(d.Content) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(d.Content))
	}
	js := toJSON(t, d)
	if !strings.Contains(js, `"type":"hardBreak"`) {
		t.Fatalf("expected hardBreak between line1/line2: %s", js)
	}
}

func TestFromPlainText_EmptyYieldsEmptyParagraph(t *testing.T) {
	d := FromPlainText("")
	if len(d.Content) != 1 || d.Content[0].Type != "paragraph" {
		t.Fatalf("empty input should yield one empty paragraph, got %+v", d.Content)
	}
}

func TestFromMarkdown_Marks(t *testing.T) {
	js := toJSON(t, FromMarkdown("a **bold** and *italic* and `code` and [x](http://e.com)"))
	for _, want := range []string{`"type":"strong"`, `"type":"em"`, `"type":"code"`, `"type":"link"`, `"href":"http://e.com"`} {
		if !strings.Contains(js, want) {
			t.Fatalf("missing %s in %s", want, js)
		}
	}
}

func TestFromMarkdown_Blocks(t *testing.T) {
	md := "# Title\n\n- one\n- two\n\n```go\nx := 1\n```\n\n> quote"
	js := toJSON(t, FromMarkdown(md))
	for _, want := range []string{`"type":"heading"`, `"level":1`, `"type":"bulletList"`, `"type":"listItem"`, `"type":"codeBlock"`, `"language":"go"`, `"type":"blockquote"`} {
		if !strings.Contains(js, want) {
			t.Fatalf("missing %s in %s", want, js)
		}
	}
}

func TestExtractText(t *testing.T) {
	// Build a doc, marshal to generic JSON, extract back to text.
	d := FromMarkdown("# Heading\n\nHello **world**\n\n- a\n- b")
	b, _ := json.Marshal(d)
	var generic any
	_ = json.Unmarshal(b, &generic)
	got := ExtractText(generic)
	for _, want := range []string{"Heading", "Hello", "world", "a", "b"} {
		if !strings.Contains(got, want) {
			t.Fatalf("extracted text %q missing %q", got, want)
		}
	}
}
