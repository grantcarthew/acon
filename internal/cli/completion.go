package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell autocompletion",
	Long: `Generate autocompletion script for the specified shell.

To load completions:

Bash:
  $ source <(acon completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ acon completion bash > /etc/bash_completion.d/acon
  # macOS:
  $ acon completion bash > $(brew --prefix)/etc/bash_completion.d/acon

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ acon completion zsh > "${fpath[1]}/_acon"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ acon completion fish | source
  # To load completions for each session, execute once:
  $ acon completion fish > ~/.config/fish/completions/acon.fish

PowerShell:
  PS> acon completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> acon completion powershell > acon.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	completionCmd.GroupID = "utility"
	rootCmd.AddCommand(completionCmd)
}
