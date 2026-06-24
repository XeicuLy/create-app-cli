# xeikit Go CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `xeikit new` コマンドで TUI テンプレート選択 → GitHub tar.gz 取得 → プロジェクト展開 を実現する Go CLI ツールを v0.1.0 としてビルド・リリースできる状態にする。

**Architecture:** Cobra でコマンドルーティング、charmbracelet/huh で TUI フォーム、`TemplateSource` インターフェースでテンプレート取得層を抽象化する。テスト時は `LocalSource`（ローカル testdata）を使い、GitHub への実アクセスは CI で行わない。

**Tech Stack:** Go 1.23、github.com/spf13/cobra、github.com/charmbracelet/huh、goreleaser、release-please

---

## ファイル構成

```
create-app-cli/                          ← リポジトリルート
├── cmd/xeikit/main.go                   # エントリポイント。version 変数を ldflags で注入
├── internal/
│   ├── template/
│   │   ├── source.go                    # Template / Manifest 型 + TemplateSource インターフェース
│   │   ├── source_test.go               # 型テスト（ほぼ不要だが export 確認用）
│   │   ├── local.go                     # LocalSource: ローカル dir から読む（テスト用）
│   │   ├── local_test.go                # testdata を使った LocalSource テスト
│   │   ├── github.go                    # GitHubSource: manifest 取得 + tar.gz 展開
│   │   └── github_test.go              # httptest モックサーバーを使ったテスト
│   ├── ui/
│   │   └── prompt.go                    # huh フォーム（プロジェクト名 + テンプレート選択）
│   └── cli/
│       ├── root.go                      # cobra.Command ルート（Use: "xeikit"）
│       ├── root_test.go                 # --version フラグのテスト
│       └── new.go                       # `xeikit new` サブコマンド本体
├── testdata/
│   ├── manifest.json                    # E2E テスト用フィクスチャ
│   └── templates/
│       └── tanstack-start-hono-cf/
│           └── package.json             # 最小限のファイル
├── .github/workflows/
│   ├── ci.yml                           # PR 時: golangci-lint + go test
│   └── release.yml                      # release-please → goreleaser
├── .goreleaser.yaml                     # クロスプラットフォームビルド設定
├── go.mod
└── go.sum
```

---

## Task 1: Go モジュールと依存関係のセットアップ

**Files:**

- Create: `go.mod`
- Create: `cmd/xeikit/main.go`

- [ ] **Step 1: go.mod を初期化する**

```bash
cd /path/to/create-app-cli
go mod init github.com/XeicuLy/create-app-cli
```

Expected output: `go: creating new go.mod: module github.com/XeicuLy/create-app-cli`

- [ ] **Step 2: 依存パッケージを追加する**

```bash
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/huh@latest
go mod tidy
```

- [ ] **Step 3: エントリポイントを作成する**

`cmd/xeikit/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/XeicuLy/create-app-cli/internal/cli"
)

var version = "dev"

func main() {
	cmd := cli.NewRootCmd(version)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: ビルドが通ることを確認する（この時点では internal/cli がないためエラーになる — 次タスクで解決）**

```bash
go build ./... 2>&1 || true
```

Expected: `cannot find package "github.com/XeicuLy/create-app-cli/internal/cli"` — これは正常

- [ ] **Step 5: コミット**

```bash
git add go.mod go.sum cmd/
git commit -m "chore: initialize Go module and entry point"
```

---

## Task 2: Template 型と TemplateSource インターフェース

**Files:**

- Create: `internal/template/source.go`

- [ ] **Step 1: 型定義を作成する**

`internal/template/source.go`:

```go
package template

// Template はテンプレートのメタデータを表す。
type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Manifest は manifest.json のスキーマを表す。
type Manifest struct {
	Templates []Template `json:"templates"`
}

// TemplateSource はテンプレートの取得元を抽象化するインターフェース。
// 本番実装は GitHubSource、テスト実装は LocalSource を使う。
type TemplateSource interface {
	ListTemplates() ([]Template, error)
	Fetch(id, destDir string) error
}
```

- [ ] **Step 2: コンパイルが通ることを確認する**

```bash
go build ./internal/template/...
```

Expected: エラーなし（出力なし）

- [ ] **Step 3: コミット**

```bash
git add internal/template/source.go
git commit -m "feat: add Template types and TemplateSource interface"
```

---

## Task 3: testdata フィクスチャを作成する

**Files:**

- Create: `testdata/manifest.json`
- Create: `testdata/templates/tanstack-start-hono-cf/package.json`

- [ ] **Step 1: manifest.json を作成する**

`testdata/manifest.json`:

```json
{
  "templates": [
    {
      "id": "tanstack-start-hono-cf",
      "name": "TanStack Start + Hono (CF Workers)",
      "description": "TanStack Start + Hono deployed to Cloudflare Workers"
    }
  ]
}
```

- [ ] **Step 2: フィクスチャテンプレートを作成する**

`testdata/templates/tanstack-start-hono-cf/package.json`:

```json
{
  "name": "tanstack-start-hono-cf",
  "version": "0.0.1",
  "private": true
}
```

- [ ] **Step 3: コミット**

```bash
git add testdata/
git commit -m "test: add fixture data for template tests"
```

---

## Task 4: LocalSource の実装とテスト

**Files:**

- Create: `internal/template/local.go`
- Create: `internal/template/local_test.go`

- [ ] **Step 1: failing テストを書く**

`internal/template/local_test.go`:

```go
package template_test

import (
	"os"
	"path/filepath"
	"testing"

	tmpl "github.com/XeicuLy/create-app-cli/internal/template"
)

func TestLocalSource_ListTemplates(t *testing.T) {
	src := &tmpl.LocalSource{BasePath: "../../testdata"}
	templates, err := src.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(templates) == 0 {
		t.Fatal("expected at least one template, got 0")
	}
	if templates[0].ID != "tanstack-start-hono-cf" {
		t.Errorf("expected id tanstack-start-hono-cf, got %s", templates[0].ID)
	}
}

func TestLocalSource_Fetch(t *testing.T) {
	src := &tmpl.LocalSource{BasePath: "../../testdata"}
	destDir := t.TempDir()

	if err := src.Fetch("tanstack-start-hono-cf", destDir); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "package.json")); err != nil {
		t.Errorf("expected package.json in output dir: %v", err)
	}
}
```

- [ ] **Step 2: テストが失敗することを確認する**

```bash
go test ./internal/template/... -run TestLocalSource -v
```

Expected: `FAIL` — `LocalSource` が未定義

- [ ] **Step 3: LocalSource を実装する**

`internal/template/local.go`:

```go
package template

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// LocalSource はローカルディレクトリからテンプレートを読み込む。テスト用。
type LocalSource struct {
	BasePath string
}

func (s *LocalSource) ListTemplates() ([]Template, error) {
	data, err := os.ReadFile(filepath.Join(s.BasePath, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return m.Templates, nil
}

func (s *LocalSource) Fetch(id, destDir string) error {
	srcDir := filepath.Join(s.BasePath, "templates", id)
	return copyDir(srcDir, destDir)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0644)
	})
}
```

- [ ] **Step 4: テストが通ることを確認する**

```bash
go test ./internal/template/... -run TestLocalSource -v
```

Expected:

```
--- PASS: TestLocalSource_ListTemplates (0.00s)
--- PASS: TestLocalSource_Fetch (0.00s)
PASS
```

- [ ] **Step 5: コミット**

```bash
git add internal/template/local.go internal/template/local_test.go
git commit -m "feat: implement LocalSource for template loading"
```

---

## Task 5: GitHubSource の実装とテスト

**Files:**

- Create: `internal/template/github.go`
- Create: `internal/template/github_test.go`

- [ ] **Step 1: failing テストを書く**

`internal/template/github_test.go`:

```go
package template_test

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

	tmpl "github.com/XeicuLy/create-app-cli/internal/template"
)

func TestGitHubSource_ListTemplates(t *testing.T) {
	manifest := tmpl.Manifest{
		Templates: []tmpl.Template{
			{ID: "test-tmpl", Name: "Test", Description: "desc"},
		},
	}
	data, _ := json.Marshal(manifest)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer srv.Close()

	src := &tmpl.GitHubSource{ManifestURL: srv.URL, TarURL: ""}
	templates, err := src.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(templates) != 1 || templates[0].ID != "test-tmpl" {
		t.Errorf("unexpected templates: %+v", templates)
	}
}

func TestGitHubSource_Fetch(t *testing.T) {
	tarData := buildTestTar(t, "test-tmpl")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(tarData)
	}))
	defer srv.Close()

	src := &tmpl.GitHubSource{ManifestURL: "", TarURL: srv.URL}
	destDir := t.TempDir()

	if err := src.Fetch("test-tmpl", destDir); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "package.json")); err != nil {
		t.Errorf("expected package.json in extracted dir: %v", err)
	}
}

// buildTestTar は指定テンプレート ID を含む tar.gz バイト列を返す。
func buildTestTar(t *testing.T, templateID string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	content := []byte(`{"name":"test"}`)
	hdr := &tar.Header{
		Name:     "starter-kit-main/templates/" + templateID + "/package.json",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	tw.Write(content)
	tw.Close()
	gw.Close()
	return buf.Bytes()
}
```

- [ ] **Step 2: テストが失敗することを確認する**

```bash
go test ./internal/template/... -run TestGitHubSource -v
```

Expected: `FAIL` — `GitHubSource` が未定義

- [ ] **Step 3: GitHubSource を実装する**

`internal/template/github.go`:

```go
package template

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultManifestURL = "https://github.com/XeicuLy/starter-kit/raw/main/manifest.json"
	defaultTarURL      = "https://codeload.github.com/XeicuLy/starter-kit/tar.gz/refs/heads/main"
	tarPrefix          = "starter-kit-main/templates/"
)

// GitHubSource は GitHub リポジトリからテンプレートを取得する。
type GitHubSource struct {
	ManifestURL string
	TarURL      string
}

// NewGitHubSource は本番用のデフォルト URL を設定した GitHubSource を返す。
func NewGitHubSource() *GitHubSource {
	return &GitHubSource{
		ManifestURL: defaultManifestURL,
		TarURL:      defaultTarURL,
	}
}

func (s *GitHubSource) ListTemplates() ([]Template, error) {
	resp, err := http.Get(s.ManifestURL)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch manifest: HTTP %d", resp.StatusCode)
	}
	var m Manifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return m.Templates, nil
}

func (s *GitHubSource) Fetch(id, destDir string) error {
	resp, err := http.Get(s.TarURL)
	if err != nil {
		return fmt.Errorf("download tar: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download tar: HTTP %d", resp.StatusCode)
	}
	return extractTemplate(resp.Body, id, destDir)
}

func extractTemplate(r io.Reader, id, destDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	prefix := tarPrefix + id + "/"

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}
		if !strings.HasPrefix(hdr.Name, prefix) {
			continue
		}
		rel := strings.TrimPrefix(hdr.Name, prefix)
		if rel == "" {
			continue
		}
		dstPath := filepath.Join(destDir, rel)
		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}
```

- [ ] **Step 4: テストが通ることを確認する**

```bash
go test ./internal/template/... -v
```

Expected:

```
--- PASS: TestLocalSource_ListTemplates (0.00s)
--- PASS: TestLocalSource_Fetch (0.00s)
--- PASS: TestGitHubSource_ListTemplates (0.00s)
--- PASS: TestGitHubSource_Fetch (0.00s)
PASS
```

- [ ] **Step 5: コミット**

```bash
git add internal/template/github.go internal/template/github_test.go
git commit -m "feat: implement GitHubSource with tar.gz template extraction"
```

---

## Task 6: huh TUI プロンプトの実装

**Files:**

- Create: `internal/ui/prompt.go`

- [ ] **Step 1: prompt.go を実装する**

`internal/ui/prompt.go`:

```go
package ui

import (
	"fmt"

	"github.com/XeicuLy/create-app-cli/internal/template"
	"github.com/charmbracelet/huh"
)

// ProjectConfig はユーザーが TUI フォームで入力した設定を保持する。
type ProjectConfig struct {
	Name       string
	TemplateID string
}

// AskProjectConfig はプロジェクト名とテンプレートを対話的に尋ねる。
func AskProjectConfig(templates []template.Template) (ProjectConfig, error) {
	var cfg ProjectConfig

	options := make([]huh.Option[string], len(templates))
	for i, t := range templates {
		options[i] = huh.NewOption(t.Name+" — "+t.Description, t.ID)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("プロジェクト名を入力してください").
				Value(&cfg.Name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("プロジェクト名は必須です")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("テンプレートを選択してください").
				Options(options...).
				Value(&cfg.TemplateID),
		),
	)

	if err := form.Run(); err != nil {
		return ProjectConfig{}, err
	}
	return cfg, nil
}
```

- [ ] **Step 2: コンパイルが通ることを確認する**

```bash
go build ./internal/ui/...
```

Expected: エラーなし（出力なし）

- [ ] **Step 3: コミット**

```bash
git add internal/ui/prompt.go
git commit -m "feat: add huh TUI prompt for project config"
```

---

## Task 7: Cobra ルートコマンドの実装とテスト

**Files:**

- Create: `internal/cli/root.go`
- Create: `internal/cli/root_test.go`

- [ ] **Step 1: failing テストを書く**

`internal/cli/root_test.go`:

```go
package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/XeicuLy/create-app-cli/internal/cli"
)

func TestRootCmd_Version(t *testing.T) {
	cmd := cli.NewRootCmd("v0.1.0-test")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), "v0.1.0-test") {
		t.Errorf("expected version in output, got: %q", buf.String())
	}
}

func TestRootCmd_HasNewSubcommand(t *testing.T) {
	cmd := cli.NewRootCmd("dev")
	for _, sub := range cmd.Commands() {
		if sub.Use == "new" {
			return
		}
	}
	t.Error("expected 'new' subcommand to be registered")
}
```

- [ ] **Step 2: テストが失敗することを確認する**

```bash
go test ./internal/cli/... -v
```

Expected: `FAIL` — `cli` パッケージが未定義

- [ ] **Step 3: root.go を実装する**

`internal/cli/root.go`:

```go
package cli

import "github.com/spf13/cobra"

// NewRootCmd は xeikit コマンドのルートを返す。
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "xeikit",
		Short:   "xeikit — scaffold your next project",
		Version: version,
	}
	cmd.AddCommand(newNewCmd())
	return cmd
}
```

- [ ] **Step 4: テストが通ることを確認する**

```bash
go test ./internal/cli/... -v
```

Expected:

```
--- PASS: TestRootCmd_Version (0.00s)
--- PASS: TestRootCmd_HasNewSubcommand (0.00s)
PASS
```

- [ ] **Step 5: コミット**

```bash
git add internal/cli/root.go internal/cli/root_test.go
git commit -m "feat: add cobra root command"
```

---

## Task 8: new サブコマンドの実装

**Files:**

- Create: `internal/cli/new.go`

- [ ] **Step 1: new.go を実装する**

`internal/cli/new.go`:

```go
package cli

import (
	"fmt"

	"github.com/XeicuLy/create-app-cli/internal/template"
	"github.com/XeicuLy/create-app-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Create a new project from a template",
		RunE:  runNew,
	}
}

func runNew(cmd *cobra.Command, _ []string) error {
	src := template.NewGitHubSource()

	templates, err := src.ListTemplates()
	if err != nil {
		return fmt.Errorf("テンプレート一覧の取得に失敗しました: %w", err)
	}

	cfg, err := ui.AskProjectConfig(templates)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ テンプレートを取得中...\n")

	if err := src.Fetch(cfg.TemplateID, cfg.Name); err != nil {
		return fmt.Errorf("テンプレートの取得に失敗しました: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ %s/ を作成しました\n\n", cfg.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "  cd %s\n  pnpm install\n", cfg.Name)

	return nil
}
```

- [ ] **Step 2: 全テストが通ることを確認する**

```bash
go test ./... -v
```

Expected: 全テスト PASS

- [ ] **Step 3: ビルドが通ることを確認する**

```bash
go build -o /tmp/xeikit ./cmd/xeikit
/tmp/xeikit --version
```

Expected: `xeikit version dev`

- [ ] **Step 4: コミット**

```bash
git add internal/cli/new.go
git commit -m "feat: implement xeikit new command"
```

---

## Task 9: CI ワークフローの設定

**Files:**

- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: ci.yml を作成する**

`.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

      - name: Test
        run: go test ./... -v -race
```

- [ ] **Step 2: コミット**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add lint and test workflow"
```

---

## Task 10: goreleaser 設定

**Files:**

- Create: `.goreleaser.yaml`

- [ ] **Step 1: .goreleaser.yaml を作成する**

`.goreleaser.yaml`:

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: xeikit
    main: ./cmd/xeikit
    binary: xeikit
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: '{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}'
    format_overrides:
      - goos: windows
        format: zip

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
```

- [ ] **Step 2: goreleaser の構文チェック（goreleaser CLI が手元にある場合）**

```bash
goreleaser check 2>/dev/null || echo "goreleaser not installed locally — skip"
```

- [ ] **Step 3: コミット**

```bash
git add .goreleaser.yaml
git commit -m "chore: add goreleaser configuration"
```

---

## Task 11: release-please + goreleaser リリースワークフロー

**Files:**

- Create: `.github/workflows/release.yml`
- Create: `.release-please-manifest.json`
- Create: `release-please-config.json`

- [ ] **Step 1: release-please 設定ファイルを作成する**

`release-please-config.json`:

```json
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "release-type": "go",
  "packages": {
    ".": {}
  }
}
```

`.release-please-manifest.json`:

```json
{
  ".": "0.0.0"
}
```

- [ ] **Step 2: release.yml を作成する**

`.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      version_type:
        description: 'Release type (auto uses conventional commits)'
        type: choice
        options: [auto, patch, minor, major]
        default: auto

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    runs-on: ubuntu-latest
    outputs:
      release_created: ${{ steps.release.outputs.release_created }}
      tag_name: ${{ steps.release.outputs.tag_name }}
    steps:
      - uses: googleapis/release-please-action@v4
        id: release
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json

  goreleaser:
    needs: release-please
    if: needs.release-please.outputs.release_created == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 3: コミット**

```bash
git add .github/workflows/release.yml release-please-config.json .release-please-manifest.json
git commit -m "ci: add release-please and goreleaser workflow"
```

---

## Task 12: README 更新

**Files:**

- Modify: `README.md`

- [ ] **Step 1: README を更新する**

`README.md`:

````markdown
# create-app-cli

Go 製のプロジェクトスキャフォールディング CLI。

## インストール

```bash
go install github.com/XeicuLy/create-app-cli/cmd/xeikit@latest
```
````

## 使い方

```bash
xeikit new
```

プロジェクト名とテンプレートを選択すると、カレントディレクトリに展開されます。

## テンプレート

テンプレートは [XeicuLy/starter-kit](https://github.com/XeicuLy/starter-kit) で管理されています。

| ID                       | 説明                                       |
| ------------------------ | ------------------------------------------ |
| `tanstack-start-hono-cf` | TanStack Start + Hono (Cloudflare Workers) |

## 開発

```bash
# テスト
go test ./... -v

# ローカルビルド
go build -o xeikit ./cmd/xeikit

# 実行
./xeikit new
```

## リリース

conventional commits に従ってコミットすると、release-please が自動でリリース PR を作成します。PR をマージすると goreleaser がバイナリをビルドして GitHub Releases に公開します。

````

- [ ] **Step 2: コミット**

```bash
git add README.md
git commit -m "docs: update README with installation and usage"
````

---

## 動作確認チェックリスト

全タスク完了後、以下を手動確認すること：

- [ ] `go test ./... -race` が全て PASS
- [ ] `go build -o /tmp/xeikit ./cmd/xeikit && /tmp/xeikit --version` が `xeikit version dev` を出力
- [ ] `go build -o /tmp/xeikit ./cmd/xeikit && /tmp/xeikit new` でプロンプトが表示される（ネットワーク接続が必要）
