package skill

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type frontmatter struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	License       string   `yaml:"license"`
	Compatibility string   `yaml:"compatibility"`
	AllowedTools  []string `yaml:"allowed-tools"`
}

// Parse reads dir/SKILL.md, parses YAML frontmatter and markdown body,
// loads assets from scripts/ references/ assets/, and validates.
func Parse(dir string) (*Skill, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return nil, fmt.Errorf("skill: reading SKILL.md: %w", err)
	}
	fmText, body, err := splitFrontmatter(raw)
	if err != nil {
		return nil, err
	}
	var fm frontmatter
	if err := yaml.Unmarshal(fmText, &fm); err != nil {
		return nil, fmt.Errorf("skill: parsing frontmatter: %w", err)
	}
	s := &Skill{
		Name:          fm.Name,
		Description:   fm.Description,
		License:       fm.License,
		Compatibility: fm.Compatibility,
		AllowedTools:  fm.AllowedTools,
		Body:          body,
		Dir:           dir,
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	// name must match parent dir per spec.
	if base := filepath.Base(dir); base != s.Name {
		return nil, fmt.Errorf("skill: name %q must match directory %q", s.Name, base)
	}
	assets, err := loadAssets(dir)
	if err != nil {
		return nil, err
	}
	s.Assets = assets
	return s, nil
}

func splitFrontmatter(raw []byte) (fm []byte, body string, err error) {
	r := bytes.TrimLeft(raw, "\xef\xbb\xbf \t\r\n")
	if !bytes.HasPrefix(r, []byte("---")) {
		return nil, "", fmt.Errorf("skill: SKILL.md must start with YAML frontmatter delimited by ---")
	}
	rest := r[3:]
	idx := bytes.Index(rest, []byte("\n---"))
	if idx < 0 {
		return nil, "", fmt.Errorf("skill: unterminated frontmatter (missing closing ---)")
	}
	fm = rest[:idx]
	after := rest[idx+len("\n---"):]
	if nl := bytes.IndexByte(after, '\n'); nl >= 0 {
		after = after[nl+1:]
	} else {
		after = nil
	}
	return fm, string(after), nil
}

func loadAssets(dir string) ([]Asset, error) {
	var assets []Asset
	for _, sub := range []string{"scripts", "references", "assets"} {
		root := filepath.Join(dir, sub)
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("skill: accessing %s: %w", sub, err)
		}
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			assets = append(assets, Asset{RelPath: filepath.ToSlash(rel), Bytes: b})
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("skill: loading %s: %w", sub, err)
		}
	}
	return assets, nil
}
