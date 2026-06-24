# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`xeikit` is a Go CLI tool that scaffolds projects by fetching templates from GitHub. The user runs `xeikit new`, picks a project name and template via a TUI, and the tool downloads and extracts the template from [XeicuLy/starter-kit](https://github.com/XeicuLy/starter-kit).

Module path: `github.com/XeicuLy/create-app-cli`
Binary name: `xeikit`
Entry point: `cmd/xeikit/main.go` — injects `version` via `-ldflags "-X main.version=..."`

## Commands

```bash
# Build
go build -o xeikit ./cmd/xeikit

# Run all tests (with race detector)
go test ./... -race -v

# Run a single test
go test ./internal/template/... -run TestLocalSource -v

# Lint
golangci-lint run
```

## Architecture

```
cmd/xeikit/main.go          — entry point; passes version to cli.NewRootCmd
internal/cli/
  root.go                   — cobra root command (Use: "xeikit")
  new.go                    — `xeikit new` subcommand; orchestrates fetch → prompt → extract
internal/template/
  source.go                 — Template, Manifest types + TemplateSource interface
  github.go                 — GitHubSource: fetches manifest.json + tar.gz from starter-kit repo
  local.go                  — LocalSource: reads from a local dir (used in tests)
internal/ui/
  prompt.go                 — charmbracelet/huh TUI form (project name + template select)
testdata/
  manifest.json             — fixture for tests
  templates/<id>/           — minimal template fixtures for E2E tests
```

### TemplateSource Interface

`TemplateSource` in `internal/template/source.go` is the key abstraction:

```go
type TemplateSource interface {
    ListTemplates() ([]Template, error)
    Fetch(id, destDir string) error
}
```

- **Production**: `GitHubSource` — downloads `manifest.json` then `tar.gz` from `codeload.github.com`
- **Tests**: `LocalSource{BasePath: "../../testdata"}` — reads from local fixtures; no GitHub access in CI

### Template Fetch Flow

1. `manifest.json` fetched from `https://github.com/XeicuLy/starter-kit/raw/main/manifest.json`
2. User selects template via huh TUI
3. Full repo tar.gz downloaded from `https://codeload.github.com/XeicuLy/starter-kit/tar.gz/refs/heads/main`
4. Entries under `starter-kit-main/templates/<selected-id>/` extracted into `./<project-name>/`

## Tech Stack

| Role           | Library                                     |
| -------------- | ------------------------------------------- |
| CLI framework  | `github.com/spf13/cobra`                    |
| TUI forms      | `github.com/charmbracelet/huh`              |
| Template fetch | Go stdlib (`archive/tar` + `compress/gzip`) |
| Release        | release-please + goreleaser                 |
| CI             | golangci-lint + `go test ./... -race`       |

## Testing Strategy

- `GitHubSource` tests use `net/http/httptest` mock servers — never hit real GitHub in CI
- `LocalSource` tests use `testdata/` fixtures
- Test for tar extraction by constructing a synthetic tar.gz in `buildTestTar` helper

## Release

Follows Conventional Commits → release-please creates Release PR → merge triggers goreleaser:

- `fix:` → patch, `feat:` → minor, `feat!:` / `BREAKING CHANGE` → major
- goreleaser builds for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `windows/amd64`
- Version is injected at build time: `-ldflags "-s -w -X main.version={{.Version}}"`
