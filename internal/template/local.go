package template

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type LocalSource struct {
	BasePath string
}

func (s LocalSource) ListTemplates() ([]Template, error) {
	data, err := os.ReadFile(filepath.Join(s.BasePath, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m.Templates, nil
}

func (s LocalSource) Fetch(id, destDir string) error {
	return copyDir(filepath.Join(s.BasePath, "templates", id), destDir)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
