package ui

import (
	"strings"
	"testing"

	"github.com/XeicuLy/create-app-cli/internal/template"
)

func TestAskProjectConfig_EmptyTemplates(t *testing.T) {
	_, err := AskProjectConfig([]template.Template{})
	if err == nil {
		t.Fatal("expected error when templates is empty, got nil")
	}
	if !strings.Contains(err.Error(), "テンプレート") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"tab only", "\t", true},
		{"valid name", "my-project", false},
		{"name with spaces", "my project", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
