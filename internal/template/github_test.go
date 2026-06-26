package template

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func buildTestTar(t *testing.T, id string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	content := []byte(`{"name":"tanstack-start-hono-cf","version":"0.0.1","private":true}`)
	entryPath := "starter-kit-main/templates/" + id + "/package.json"
	_ = tw.WriteHeader(&tar.Header{
		Name: entryPath,
		Mode: 0644,
		Size: int64(len(content)),
	})
	_, _ = tw.Write(content)
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func TestGitHubSource_ListTemplates(t *testing.T) {
	manifest := Manifest{
		Templates: []Template{
			{ID: "tanstack-start-hono-cf", Name: "TanStack Start + Hono (CF Workers)", Description: "TanStack Start + Hono deployed to Cloudflare Workers"},
		},
	}
	body, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	src := &GitHubSource{ManifestURL: srv.URL, TarURL: ""}
	templates, err := src.ListTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}
	if templates[0].ID != "tanstack-start-hono-cf" {
		t.Errorf("expected ID tanstack-start-hono-cf, got %s", templates[0].ID)
	}
}

func TestGitHubSource_ListTemplates_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	src := &GitHubSource{ManifestURL: srv.URL, TarURL: ""}
	_, err := src.ListTemplates()
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

func TestGitHubSource_Fetch(t *testing.T) {
	id := "tanstack-start-hono-cf"
	tarData := buildTestTar(t, id)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-gzip")
		_, _ = w.Write(tarData)
	}))
	defer srv.Close()

	destDir := t.TempDir()
	src := &GitHubSource{ManifestURL: "", TarURL: srv.URL}
	if err := src.Fetch(id, destDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pkgPath := filepath.Join(destDir, "package.json")
	if _, err := os.Stat(pkgPath); err != nil {
		t.Errorf("expected package.json to exist at %s: %v", pkgPath, err)
	}
}

func TestGitHubSource_Fetch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	src := &GitHubSource{ManifestURL: "", TarURL: srv.URL}
	err := src.Fetch("tanstack-start-hono-cf", t.TempDir())
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}
