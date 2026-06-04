package skill

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		s       Skill
		wantErr bool
	}{
		{"ok", Skill{Name: "pdf-tools", Description: "Work with PDFs"}, false},
		{"missing name", Skill{Description: "x"}, true},
		{"missing desc", Skill{Name: "pdf-tools"}, true},
		{"bad chars", Skill{Name: "PDF_Tools", Description: "x"}, true},
		{"too long name", Skill{Name: strings.Repeat("a", 65), Description: "x"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestParse(t *testing.T) {
	s, err := Parse("testdata/good")
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "good" || s.Description != "A good skill" {
		t.Fatalf("frontmatter wrong: %+v", s)
	}
	if s.License != "MIT" || s.Compatibility != "Designed for Claude Code" {
		t.Fatalf("optional fields wrong: %+v", s)
	}
	if len(s.AllowedTools) != 2 || s.AllowedTools[0] != "Read" {
		t.Fatalf("allowed-tools wrong: %v", s.AllowedTools)
	}
	if !strings.Contains(s.Body, "Body heading") {
		t.Fatalf("body not captured: %q", s.Body)
	}
	if len(s.Assets) != 1 || s.Assets[0].RelPath != "scripts/run.sh" {
		t.Fatalf("assets wrong: %+v", s.Assets)
	}
}
