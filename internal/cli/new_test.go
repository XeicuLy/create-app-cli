package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/XeicuLy/create-app-cli/internal/template"
	"github.com/XeicuLy/create-app-cli/internal/ui"
)

func TestNewCmd_ExistingDirectory(t *testing.T) {
	dir := t.TempDir()
	existingName := filepath.Join(dir, "my-project")
	if err := os.Mkdir(existingName, 0o755); err != nil {
		t.Fatal(err)
	}

	c := &newCmd{
		src: &template.LocalSource{BasePath: "../../testdata"},
		promptFn: func(_ []template.Template) (ui.ProjectConfig, error) {
			return ui.ProjectConfig{Name: existingName, TemplateID: "tanstack-start-hono-cf"}, nil
		},
	}

	rootCmd := NewRootCmd("dev")
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "new" {
			sub.RunE = c.run
		}
	}

	rootCmd.SetArgs([]string{"new"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for existing directory, got nil")
	}
	if !strings.Contains(err.Error(), "すでに存在") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "すでに存在")
	}
}

func TestNewCmd_Success(t *testing.T) {
	destParent := t.TempDir()
	projectName := filepath.Join(destParent, "my-new-project")

	src := &template.LocalSource{BasePath: "../../testdata"}
	c := &newCmd{
		src: src,
		promptFn: func(_ []template.Template) (ui.ProjectConfig, error) {
			return ui.ProjectConfig{Name: projectName, TemplateID: "tanstack-start-hono-cf"}, nil
		},
	}

	cmd := newCmdWithDI(c)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	pkgPath := filepath.Join(projectName, "package.json")
	if _, err := os.Stat(pkgPath); err != nil {
		t.Errorf("package.json not found in project dir: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "pnpm install") {
		t.Errorf("output = %q, want to contain pnpm install instructions", out)
	}
}
