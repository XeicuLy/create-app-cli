package ui

import (
	"fmt"

	"github.com/XeicuLy/create-app-cli/internal/template"
	"github.com/charmbracelet/huh"
)

// ProjectConfig holds the user's choices from the TUI form.
type ProjectConfig struct {
	Name       string
	TemplateID string
}

// AskProjectConfig runs an interactive huh form and returns the user's input.
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
