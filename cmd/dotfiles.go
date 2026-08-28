package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nilay-banerjee/noob-cli/internal/dotfiles"
)

var dotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "Manage the ~/dotfiles repo and its symlinks",
}

var dotfilesInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Harvest this machine's configs into ~/dotfiles and symlink them back",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dotfiles.Harvest(dryRun)
	},
}

var dotfilesLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Symlink ~/dotfiles configs into place (clones first with --dotfiles-repo)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dotfiles.Link(dotfilesRepo, dryRun)
	},
}

func init() {
	dotfilesCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print what would happen without changing anything")
	dotfilesLinkCmd.Flags().StringVar(&dotfilesRepo, "dotfiles-repo", "", "git URL to clone into ~/dotfiles if it's missing")
	dotfilesCmd.AddCommand(dotfilesInitCmd, dotfilesLinkCmd)
	rootCmd.AddCommand(dotfilesCmd)
}
