package cli

import "github.com/spf13/cobra"

func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "xeikit",
		Short:   "xeikit — プロジェクトスキャフォールディングツール",
		Version: version,
	}
	cmd.AddCommand(newNewCmd())
	return cmd
}
