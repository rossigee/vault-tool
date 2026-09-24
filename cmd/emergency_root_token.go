package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/gpg"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/vault"
)

var (
	emergencyAddr        string
	emergencyKeysFile    string
	emergencyPassphrase  string
	emergencyTokenTTL    string
)

var emergencyRootTokenCmd = &cobra.Command{
	Use:   "emergency-root-token",
	Short: "Generate an emergency root token",
	Long:  "Use unseal keys to generate a temporary root token for emergency maintenance when OIDC auth is unavailable",
	RunE:  runEmergencyRootToken,
}

func init() {
	emergencyRootTokenCmd.Flags().StringVar(&emergencyAddr, "addr", "https://vault.bankrut.lan", "Vault address")
	emergencyRootTokenCmd.Flags().StringVar(&emergencyKeysFile, "keys-file", os.ExpandEnv("$HOME/.config/vault/bankrut-unseal-keys.gpg"), "Path to GPG-encrypted unseal keys")
	emergencyRootTokenCmd.Flags().StringVar(&emergencyPassphrase, "passphrase", "", "GPG passphrase (read from stdin if not provided)")
	emergencyRootTokenCmd.Flags().StringVar(&emergencyTokenTTL, "ttl", "24h", "Token TTL (default 24 hours)")
}

func runEmergencyRootToken(cmd *cobra.Command, args []string) error {
	keys, err := gpg.DecryptKeys(emergencyKeysFile, emergencyPassphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt keys: %w", err)
	}

	if len(keys) == 0 {
		return fmt.Errorf("no unseal keys found in %s", emergencyKeysFile)
	}

	// Validate TTL is parseable
	_, err = time.ParseDuration(emergencyTokenTTL)
	if err != nil {
		return fmt.Errorf("invalid TTL: %w", err)
	}

	client := vault.NewClient(emergencyAddr)

	fmt.Fprintf(os.Stderr, "Starting root token generation process...\n")

	initResp, err := client.InitRootTokenGeneration(cmd.Context(), emergencyTokenTTL)
	if err != nil {
		return fmt.Errorf("failed to initialize root token generation: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Root token generation initialized. Nonce: %s\n", initResp.Nonce)
	fmt.Fprintf(os.Stderr, "Submitting %d unseal key(s) to complete generation...\n", len(keys))

	for i, key := range keys {
		if key == "" {
			continue
		}

		fmt.Fprintf(os.Stderr, "Submitting key %d/%d...\n", i+1, len(keys))

		updateResp, err := client.UpdateRootTokenGeneration(cmd.Context(), initResp.Nonce, key)
		if err != nil {
			return fmt.Errorf("failed to update root token generation at key %d: %w", i+1, err)
		}

		fmt.Fprintf(os.Stderr, "  progress: %d/%d\n", updateResp.Progress, updateResp.Required)

		if updateResp.Complete {
			fmt.Fprintf(os.Stderr, "Root token generation completed.\n")
			fmt.Println(updateResp.EncodedToken)
			return nil
		}
	}

	return fmt.Errorf("root token generation incomplete after submitting all keys")
}
