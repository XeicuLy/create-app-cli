package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalSource_ListTemplates(t *testing.T) {
	src := LocalSource{BasePath: "../../testdata"}

	templates, err := src.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("ListTemplates() len = %d, want 1", len(templates))
	}

	got := templates[0]
	if got.ID != "tanstack-start-hono-cf" {
		t.Errorf("ID = %q, want %q", got.ID, "tanstack-start-hono-cf")
	}
	if got.Name != "TanStack Start + Hono (CF Workers)" {
		t.Errorf("Name = %q, want %q", got.Name, "TanStack Start + Hono (CF Workers)")
	}
	if got.Description != "TanStack Start + Hono deployed to Cloudflare Workers" {
		t.Errorf("Description = %q, want %q", got.Description, "TanStack Start + Hono deployed to Cloudflare Workers")
	}
}

func TestLocalSource_Fetch(t *testing.T) {
	src := LocalSource{BasePath: "../../testdata"}
	destDir := t.TempDir()

	if err := src.Fetch("tanstack-start-hono-cf", destDir); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	pkgPath := filepath.Join(destDir, "package.json")
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("package.json not found in dest: %v", err)
	}

	wantContent := "{\n  \"name\": \"tanstack-start-hono-cf\",\n  \"version\": \"0.0.1\",\n  \"private\": true\n}\n"
	if string(content) != wantContent {
		t.Errorf("package.json content =\n%s\nwant:\n%s", content, wantContent)
	}
}
