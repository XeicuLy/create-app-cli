# xeikit Go CLI — 設計仕様書

## 概要

Go 製の CLI ツール `xeikit` を開発する。`xeikit new` コマンドでインタラクティブにテンプレートを選択し、プロジェクトをスキャフォールドする。TypeScript 製の [create-xeikit-app](https://github.com/xeikit/create-xeikit-app) の Go 版に相当する。

---

## リポジトリ構成

個人リポジトリとして作成する（org リポジトリだと GitHub プロフィールの言語解析に乗らないため）。

| リポジトリ               | 用途                 |
| ------------------------ | -------------------- |
| `XeicuLy/create-app-cli` | CLI ツール本体（Go） |
| `XeicuLy/starter-kit`    | テンプレート置き場   |

---

## コマンド設計

### バイナリ名

```
xeikit
```

### サブコマンド

```bash
xeikit new        # プロジェクト作成（メインコマンド）
```

### 実行フロー

```
$ xeikit new

? プロジェクト名を入力してください: my-app
? テンプレートを選択してください:
  ▸ tanstack-start-hono-cf  TanStack Start + Hono (Cloudflare Workers)

✓ テンプレートを取得中...
✓ my-app/ を作成しました

  cd my-app
  pnpm install
```

---

## アーキテクチャ

### CLI リポジトリ構造

```
create-app-cli/
├── cmd/
│   └── xeikit/
│       └── main.go            # エントリポイント
├── internal/
│   ├── cli/
│   │   ├── root.go            # cobra root command
│   │   └── new.go             # `xeikit new` サブコマンド
│   ├── template/
│   │   ├── source.go          # TemplateSource インターフェース
│   │   ├── github.go          # GitHub tar.gz ダウンロード実装
│   │   └── local.go           # ローカル実装（テスト用）
│   └── ui/
│       └── prompt.go          # huh フォーム定義
├── .github/
│   └── workflows/
│       ├── ci.yml             # PR 時: lint + test
│       └── release.yml        # v* タグ push 時: goreleaser
├── .goreleaser.yaml
├── go.mod
└── go.sum
```

### テンプレートリポジトリ構造

```
starter-kit/
├── templates/
│   └── tanstack-start-hono-cf/   # 初期テンプレート（1種類から開始）
└── manifest.json                 # テンプレート一覧・説明メタデータ
```

### manifest.json 形式

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

---

## 技術スタック

| 役割               | 採用技術                                                  | 選定理由                                                 |
| ------------------ | --------------------------------------------------------- | -------------------------------------------------------- |
| CLI フレームワーク | [Cobra](https://github.com/spf13/cobra)                   | Go CLI のデファクトスタンダード、サブコマンド拡張が容易  |
| TUI                | [charmbracelet/huh](https://github.com/charmbracelet/huh) | フォーム特化型で実装コスト低、charmbracelet エコシステム |
| テンプレート取得   | Go 標準ライブラリ（`archive/tar` + `compress/gzip`）      | 外部依存ゼロ                                             |
| リリース           | release-please + goreleaser                               | Google 製で活発にメンテ、goreleaser との統合実績豊富     |
| CI                 | golangci-lint + go test                                   | 標準的な Go CI 構成                                      |

---

## テンプレート取得の仕組み

```
xeikit new
  │
  ├─ ① manifest.json を取得
  │    https://github.com/XeicuLy/starter-kit/raw/main/manifest.json
  │    → テンプレート名・説明一覧をロード → huh で選択肢表示
  │
  ├─ ② ユーザーがプロジェクト名・テンプレートを選択
  │
  └─ ③ tar.gz ダウンロード → 展開 → リネーム
       https://codeload.github.com/XeicuLy/starter-kit/tar.gz/refs/heads/main
       └─ starter-kit-main/templates/<選択テンプレート>/ → ./<プロジェクト名>/
```

### TemplateSource インターフェース

テスト時にローカル実装へ差し替えられるよう、インターフェースで抽象化する。

```go
type TemplateSource interface {
    ListTemplates() ([]Template, error)
    Fetch(id, destDir string) error
}

// 本番: GitHub tar.gz ダウンロード
type GitHubSource struct { ... }

// テスト: ローカルディレクトリ（testdata/templates/）
type LocalSource struct { BasePath string }
```

---

## リリース管理

### バージョン戦略

Semantic Versioning に従う。

| コミットプレフィックス       | バージョン変動             |
| ---------------------------- | -------------------------- |
| `fix:`                       | patch（例: 1.0.0 → 1.0.1） |
| `feat:`                      | minor（例: 1.0.0 → 1.1.0） |
| `feat!:` / `BREAKING CHANGE` | major（例: 1.0.0 → 2.0.0） |

マイルストーン目安：

```
v0.1.0  最初の動く版（tanstack-start-hono-cf テンプレートのみ）
v0.x.0  テンプレート追加・機能追加
v1.0.0  複数テンプレート対応、安定版
```

### リリースフロー

```
conventional commits → release-please が Release PR を自動作成（CHANGELOG 付き）
  └─ PR をマージ → v1.x.x タグ作成
       └─ goreleaser 起動 → クロスプラットフォームバイナリビルド
            └─ GitHub Releases に公開
```

初回メジャーリリース（`v1.0.0`）は `workflow_dispatch` で `major` を手動選択。

### 配布バイナリ

goreleaser が以下のターゲット向けにビルド・公開する：

- `xeikit_darwin_amd64`
- `xeikit_darwin_arm64`（M1/M2 Mac）
- `xeikit_linux_amd64`
- `xeikit_windows_amd64.exe`

### インストール方法（ユーザー向け）

```bash
# go install
go install github.com/XeicuLy/create-app-cli/cmd/xeikit@latest

# GitHub Releases から直接 DL（Go 不要）
# goreleaser が install script を自動生成
```

### ワークフロー構成（2ファイル）

```yaml
# .github/workflows/ci.yml（PR 時に実行）
- golangci-lint
- go test ./...

# .github/workflows/release.yml（Release PR マージ時に実行）
- goreleaser release
```

---

## テスト戦略

- `TemplateSource` インターフェースにより、テスト時は `LocalSource`（`testdata/templates/`）を使用
- GitHub への実アクセスは CI では行わない
- E2E テストは最小構成のフィクスチャテンプレートを `testdata/` に置いて展開まで通す

```
internal/template/
  ├── github_test.go     # HTTP サーバーをモックして DL をテスト
  └── local_test.go

internal/cli/
  └── new_test.go        # cobra コマンドの入出力テスト

testdata/
  └── templates/
      └── tanstack-start-hono-cf/   # E2E テスト用フィクスチャ
```

---

## 将来の拡張候補（スコープ外）

- テンプレート追加（tanstack-start-hono-node / tanstack-start-go / nuxt / go-react）
- `xeikit list` — 利用可能なテンプレート一覧を表示
- `--template` フラグによる非インタラクティブモード
- Homebrew tap 対応
