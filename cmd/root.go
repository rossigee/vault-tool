package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vault-tool",
	Short: "Vault emergency access and maintenance tool",
	Long:  "CLI tool for Vault unsealing and emergency root token generation",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(unsealCmd)
	rootCmd.AddCommand(emergencyRootTokenCmd)
}
