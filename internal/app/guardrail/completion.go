package guardrail

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.AddCommand(completionInstallCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for guardrail.

To load completions:

Bash:
  $ source <(guardrail completion bash)
  # To load completions for each session, add to your bashrc:
  # Linux:
  $ guardrail completion bash > /etc/bash_completion.d/guardrail
  # macOS:
  $ guardrail completion bash > $(brew --prefix)/etc/bash_completion.d/guardrail

Zsh:
  $ guardrail completion zsh > "${fpath[1]}/_guardrail"
  # You may need to start a new shell for this to take effect.

Fish:
  $ guardrail completion fish > ~/.config/fish/completions/guardrail.fish

PowerShell:
  PS> guardrail completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, add to your profile:
  PS> guardrail completion powershell >> $PROFILE

Or use 'guardrail completion install' to auto-detect and install.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			_ = rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			_ = rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

var completionInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Auto-detect shell and install completions",
	Long: `Detect the current shell and install completion scripts to the appropriate location.

Supported shells and locations:
  Fish: ~/.config/fish/completions/guardrail.fish
  Zsh:  ~/.oh-my-zsh/completions/_guardrail (if oh-my-zsh detected)
  Bash: ~/.local/share/bash-completion/completions/guardrail

If the install location cannot be determined, the command will error with manual instructions.
`,
	Run: func(cmd *cobra.Command, args []string) {
		shell := detectShell()
		if shell == "" {
			fmt.Fprintln(os.Stderr, "error: cannot detect shell from $SHELL")
			fmt.Fprintln(os.Stderr, "hint: Run 'guardrail completion --help' for manual installation")
			os.Exit(1)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot determine home directory: %v\n", err)
			os.Exit(1)
		}

		var installPath string
		var generator func() error

		switch shell {
		case "fish":
			installPath = filepath.Join(home, ".config", "fish", "completions", "guardrail.fish")
			generator = func() error {
				f, err := os.Create(installPath)
				if err != nil {
					return err
				}
				defer func() { _ = f.Close() }()
				return rootCmd.GenFishCompletion(f, true)
			}

		case "zsh":
			// Check for oh-my-zsh first (most common)
			omzPath := filepath.Join(home, ".oh-my-zsh", "completions")
			if dirExists(omzPath) {
				installPath = filepath.Join(omzPath, "_guardrail")
			} else {
				// Check for ~/.zfunc (common custom setup)
				zfuncPath := filepath.Join(home, ".zfunc")
				if dirExists(zfuncPath) {
					installPath = filepath.Join(zfuncPath, "_guardrail")
				} else {
					fmt.Fprintln(os.Stderr, "error: cannot determine zsh completion directory")
					fmt.Fprintln(os.Stderr, "")
					fmt.Fprintln(os.Stderr, "For oh-my-zsh, create the directory:")
					fmt.Fprintln(os.Stderr, "  mkdir -p ~/.oh-my-zsh/completions")
					fmt.Fprintln(os.Stderr, "  guardrail completion install")
					fmt.Fprintln(os.Stderr, "")
					fmt.Fprintln(os.Stderr, "Or install manually:")
					fmt.Fprintf(os.Stderr, "  guardrail completion zsh > \"${fpath[1]}/_guardrail\"\n")
					os.Exit(1)
				}
			}
			generator = func() error {
				f, err := os.Create(installPath)
				if err != nil {
					return err
				}
				defer func() { _ = f.Close() }()
				return rootCmd.GenZshCompletion(f)
			}

		case "bash":
			// Use XDG standard location
			completionsDir := filepath.Join(home, ".local", "share", "bash-completion", "completions")
			if !dirExists(completionsDir) {
				// Try to create it
				if err := os.MkdirAll(completionsDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "error: cannot create bash completions directory: %v\n", err)
					fmt.Fprintln(os.Stderr, "")
					fmt.Fprintln(os.Stderr, "Install manually:")
					fmt.Fprintln(os.Stderr, "  guardrail completion bash >> ~/.bashrc")
					os.Exit(1)
				}
			}
			installPath = filepath.Join(completionsDir, "guardrail")
			generator = func() error {
				f, err := os.Create(installPath)
				if err != nil {
					return err
				}
				defer func() { _ = f.Close() }()
				return rootCmd.GenBashCompletion(f)
			}

		default:
			fmt.Fprintf(os.Stderr, "error: unsupported shell '%s'\n", shell)
			fmt.Fprintln(os.Stderr, "hint: Run 'guardrail completion --help' for manual installation")
			os.Exit(1)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot create directory: %v\n", err)
			os.Exit(1)
		}

		if err := generator(); err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to write completion file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Installed %s completions to %s\n", shell, installPath)
		if shell == "zsh" || shell == "bash" {
			fmt.Println("Restart your shell or source your profile to enable completions")
		}
	},
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	base := filepath.Base(shell)
	switch {
	case strings.Contains(base, "fish"):
		return "fish"
	case strings.Contains(base, "zsh"):
		return "zsh"
	case strings.Contains(base, "bash"):
		return "bash"
	default:
		return base
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
