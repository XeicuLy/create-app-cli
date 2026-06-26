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

type GitHubSource struct {
	ManifestURL string
	TarURL      string
}

func NewGitHubSource() *GitHubSource {
	return &GitHubSource{
		ManifestURL: "https://github.com/XeicuLy/starter-kit/raw/main/manifest.json",
		TarURL:      "https://codeload.github.com/XeicuLy/starter-kit/tar.gz/refs/heads/main",
	}
}

func (g *GitHubSource) ListTemplates() ([]Template, error) {
	resp, err := http.Get(g.ManifestURL) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("manifest の取得に失敗しました: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest の取得に失敗しました: HTTP %d", resp.StatusCode)
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("manifest のデコードに失敗しました: %w", err)
	}
	return manifest.Templates, nil
}

func (g *GitHubSource) Fetch(id, destDir string) error {
	resp, err := http.Get(g.TarURL) //nolint:noctx
	if err != nil {
		return fmt.Errorf("テンプレートの取得に失敗しました: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("テンプレートの取得に失敗しました: HTTP %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gzip の読み込みに失敗しました: %w", err)
	}
	defer func() { _ = gz.Close() }()

	prefix := "starter-kit-main/templates/" + id + "/"
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar の読み込みに失敗しました: %w", err)
		}

		if !strings.HasPrefix(hdr.Name, prefix) {
			continue
		}

		rel := strings.TrimPrefix(hdr.Name, prefix)
		if rel == "" {
			continue
		}

		if hdr.Typeflag == tar.TypeSymlink || hdr.Typeflag == tar.TypeLink {
			continue
		}

		target := filepath.Join(destDir, filepath.FromSlash(rel))
		destAbs, err := filepath.Abs(filepath.Clean(destDir))
		if err != nil {
			return fmt.Errorf("展開先パスの解決に失敗しました: %w", err)
		}
		targetAbs, err := filepath.Abs(filepath.Clean(target))
		if err != nil {
			return fmt.Errorf("展開先パスの解決に失敗しました: %w", err)
		}
		if !strings.HasPrefix(targetAbs, destAbs+string(os.PathSeparator)) && targetAbs != destAbs {
			return fmt.Errorf("不正なパスを検出しました: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
			}
		default:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
			}
			f, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("ファイルの作成に失敗しました: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
			}
			if err := f.Close(); err != nil {
				return fmt.Errorf("ファイルのクローズに失敗しました: %w", err)
			}
		}
	}
	return nil
}
