# xeikit

Go 製の CLI プロジェクトスキャフォールディングツール。`xeikit new` コマンドを実行すると、TUI でプロジェクト名とテンプレートを選択し、[XeicuLy/starter-kit](https://github.com/XeicuLy/starter-kit) からテンプレートを取得・展開します。

## インストール

```bash
go install github.com/XeicuLy/create-app-cli/cmd/xeikit@latest
```

インストール後、`xeikit` コマンドが使用可能になります。

## 使い方

```bash
xeikit new
```

実行すると以下の手順でプロジェクトを作成します：

1. プロジェクト名を入力
2. テンプレートを選択（TUI）
3. 選択したテンプレートが `./<プロジェクト名>/` に展開される

### 実行例

```bash
$ xeikit new
? プロジェクト名: my-app
? テンプレートを選択:
  > tanstack-start-hono-cf

✓ プロジェクト my-app を作成しました
```

## テンプレート一覧

| テンプレート ID          | 説明                                                                      |
| ------------------------ | ------------------------------------------------------------------------- |
| `tanstack-start-hono-cf` | TanStack Start + Hono + Cloudflare Workers 構成のフルスタックテンプレート |

## 開発

### 前提条件

- Go 1.22 以上
- [golangci-lint](https://golangci-lint.run/)

### セットアップ・ビルド

```bash
# ビルド
go build -o xeikit ./cmd/xeikit

# 動作確認
./xeikit new
```

### テスト

```bash
# 全テスト実行（レースコンディション検出付き）
go test ./... -v -race

# 特定パッケージのテスト
go test ./internal/template/... -run TestLocalSource -v
```

テストは GitHub への実際のアクセスを行いません。`GitHubSource` は `net/http/httptest` モックサーバーを使用し、`LocalSource` は `testdata/` のフィクスチャを使用します。

### Lint

```bash
golangci-lint run
```

## アーキテクチャ

```text
cmd/xeikit/main.go          — エントリポイント（version を cli.NewRootCmd に渡す）
internal/cli/
  root.go                   — cobra ルートコマンド
  new.go                    — xeikit new サブコマンド（fetch → prompt → extract を orchestrate）
internal/template/
  source.go                 — Template / Manifest 型 + TemplateSource インターフェース
  github.go                 — GitHubSource: starter-kit から manifest.json + tar.gz を取得
  local.go                  — LocalSource: ローカルディレクトリから読み込み（テスト用）
internal/ui/
  prompt.go                 — charmbracelet/huh TUI フォーム
```

`TemplateSource` インターフェースによる DI パターンで、本番は `GitHubSource`、テストは `LocalSource` に差し替えられます。

## リリースフロー

[Conventional Commits](https://www.conventionalcommits.org/) に従ったコミットメッセージを使用します：

| プレフィックス               | バージョン変更 |
| ---------------------------- | -------------- |
| `fix:`                       | patch          |
| `feat:`                      | minor          |
| `feat!:` / `BREAKING CHANGE` | major          |

リリース手順：

1. `fix:` / `feat:` コミットをマージ
2. [release-please](https://github.com/googleapis/release-please) が自動でリリース PR を作成
3. リリース PR をマージすると [goreleaser](https://goreleaser.com/) が起動
4. `darwin/amd64`・`darwin/arm64`・`linux/amd64`・`windows/amd64` 向けバイナリが自動ビルド・リリース

## ライセンス

MIT
