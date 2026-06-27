package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_Version(t *testing.T) {
	cmd := NewRootCmd("1.2.3")
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "1.2.3") {
		t.Errorf("--version output = %q, want to contain %q", got, "1.2.3")
	}
}

func TestRootCmd_HasNewSubcommand(t *testing.T) {
	cmd := NewRootCmd("dev")
	for _, sub := range cmd.Commands() {
		if sub.Use == "new" {
			return
		}
	}
	t.Error("root command does not have 'new' subcommand")
}
