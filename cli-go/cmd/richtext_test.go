package cmd

import (
	"strings"
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/adf"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/auth"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/config"
)

func useClient(t *testing.T, apiVersion string) {
	t.Helper()
	a, err := auth.New(config.Profile{Site: "https://jira.example.com", Token: "pat", AuthType: config.AuthPAT}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	old := jiraClient
	jiraClient = client.NewClient(a, apiVersion, false, false)
	t.Cleanup(func() { jiraClient = old })
}

func TestRichTextIsAStringOnAPIv2(t *testing.T) {
	useClient(t, "2")
	got, err := richText("*bold* text", false)
	if err != nil || got != "*bold* text" {
		t.Fatalf("richText on v2 = %#v, %v; want the plain string", got, err)
	}
	if _, err := richText("**bold**", true); err == nil || !strings.Contains(err.Error(), "--markdown") {
		t.Fatalf("richText --markdown on v2: err = %v; want a --markdown error", err)
	}
}

func TestRichTextIsADFOnAPIv3(t *testing.T) {
	useClient(t, "3")
	for _, markdown := range []bool{false, true} {
		got, err := richText("text", markdown)
		if err != nil {
			t.Fatal(err)
		}
		doc, ok := got.(adf.Doc)
		if !ok || doc.Type != "doc" {
			t.Fatalf("richText(markdown=%v) on v3 = %#v; want an ADF doc", markdown, got)
		}
	}
}
