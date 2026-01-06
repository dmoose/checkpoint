package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for checkpoint.

To load completions:

Bash:
  $ source <(checkpoint completion bash)
  # To load completions for each session, add to your bashrc:
  # Linux:
  $ checkpoint completion bash > /etc/bash_completion.d/checkpoint
  # macOS:
  $ checkpoint completion bash > $(brew --prefix)/etc/bash_completion.d/checkpoint

Zsh:
  $ checkpoint completion zsh > "${fpath[1]}/_checkpoint"
  # You may need to start a new shell for this to take effect.

Fish:
  $ checkpoint completion fish > ~/.config/fish/completions/checkpoint.fish

PowerShell:
  PS> checkpoint completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, add to your profile:
  PS> checkpoint completion powershell >> $PROFILE
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}
