package cli

import (
	"fmt"
	"os"

	"github.com/XeicuLy/create-app-cli/internal/template"
	"github.com/XeicuLy/create-app-cli/internal/ui"
	"github.com/spf13/cobra"
)

type newCmd struct {
	src      template.TemplateSource
	promptFn func([]template.Template) (ui.ProjectConfig, error)
}

func newNewCmd() *cobra.Command {
	c := &newCmd{
		src:      template.NewGitHubSource(),
		promptFn: ui.AskProjectConfig,
	}
	return newCmdWithDI(c)
}

func newCmdWithDI(c *newCmd) *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "新しいプロジェクトを作成します",
		RunE:  c.run,
	}
}

func (c *newCmd) run(cmd *cobra.Command, _ []string) error {
	templates, err := c.src.ListTemplates()
	if err != nil {
		return fmt.Errorf("テンプレート一覧の取得に失敗しました: %w", err)
	}

	cfg, err := c.promptFn(templates)
	if err != nil {
		return fmt.Errorf("プロンプトに失敗しました: %w", err)
	}

	if _, err := os.Stat(cfg.Name); err == nil {
		return fmt.Errorf("ディレクトリ %q はすでに存在します", cfg.Name)
	}

	cmd.Println("プロジェクトを作成中です...")

	if err := c.src.Fetch(cfg.TemplateID, cfg.Name); err != nil {
		return fmt.Errorf("テンプレートの展開に失敗しました: %w", err)
	}

	cmd.Printf("\n✅ プロジェクト %q を作成しました！\n\n", cfg.Name)
	cmd.Println("次のステップ:")
	cmd.Printf("  cd %s\n", cfg.Name)
	cmd.Println("  pnpm install")

	return nil
}
