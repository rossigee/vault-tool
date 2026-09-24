package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"git.golder.lan/rossgolderltd/vault-tool/internal/version"
)

var rootCmd = &cobra.Command{
	Use:   "vault-tool",
	Short: "Vault emergency access and maintenance tool",
	Long:  "CLI tool for Vault unsealing and emergency root token generation",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("vault-tool version %s\n", version.Version)
		fmt.Printf("Commit: %s\n", version.Commit)
		fmt.Printf("Build Time: %s\n", version.BuildTime)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(unsealCmd)
	rootCmd.AddCommand(emergencyRootTokenCmd)
	rootCmd.AddCommand(versionCmd)
}
