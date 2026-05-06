package cmd

import (
	"fmt"
	"os"

	"agentai-go/cmd/chat"
	"agentai-go/cmd/config"
	"agentai-go/cmd/version"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

// Build-time version (set via ldflags).
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "agentai",
	Short: "Agentic AI Code Assistant",
	Long:  "Agentic AI Code Assistant - An intelligent code generation tool.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// Execute runs the root command and all subcommands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// RootCmd returns the root command for testing.
func RootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(chat.NewCommand())
	rootCmd.AddCommand(config.NewCommand())
	rootCmd.AddCommand(version.NewCommand(func() (string, string, string) {
		return Version, Commit, BuildDate
	}))
}
