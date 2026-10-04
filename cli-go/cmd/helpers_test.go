package cmd

import (
	"os"
	"testing"
)

func TestPromptsShareBufferedPipedStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString("https://acme.atlassian.net\nme@acme.com\nsecret\n"); err != nil {
		t.Fatal(err)
	}
	w.Close()
	oldStdin, oldNoInput := os.Stdin, noInputFlag
	os.Stdin, noInputFlag = r, false
	t.Cleanup(func() { os.Stdin, noInputFlag = oldStdin, oldNoInput; r.Close() })

	for _, want := range []string{"https://acme.atlassian.net", "me@acme.com"} {
		if got, err := prompt(""); err != nil || got != want {
			t.Fatalf("prompt() = %q, %v; want %q", got, err, want)
		}
	}
	if got, err := promptSecret(""); err != nil || got != "secret" {
		t.Fatalf("promptSecret() = %q, %v; want %q", got, err, "secret")
	}
}
