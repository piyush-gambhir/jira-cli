package cmd

import "testing"

func TestCheckEnableUndoRefusesServerDC(t *testing.T) {
	if err := checkEnableUndo("2", true); err == nil {
		t.Fatal("--enable-undo on REST v2 must be refused: Server/DC deletes permanently")
	}
	for _, c := range []struct {
		version string
		undo    bool
	}{{"3", true}, {"3", false}, {"2", false}} {
		if err := checkEnableUndo(c.version, c.undo); err != nil {
			t.Fatalf("checkEnableUndo(%q, %v): %v", c.version, c.undo, err)
		}
	}
}

func TestAttachmentFileNameStaysInCurrentDirectory(t *testing.T) {
	cases := map[string]string{
		"report.pdf":           "report.pdf",
		"../../.bashrc":        ".bashrc",
		"/etc/passwd":          "passwd",
		`..\..\evil.exe`:       "evil.exe",
		"dir/../../secret.txt": "secret.txt",
		"..":                   "attachment-42",
		"":                     "attachment-42",
		"/":                    "attachment-42",
	}
	for in, want := range cases {
		if got := attachmentFileName(in, "42"); got != want {
			t.Errorf("attachmentFileName(%q) = %q, want %q", in, got, want)
		}
	}
}
