package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/gpg"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/vault"
)

var (
	unsealAddr      string
	unsealKeysFile  string
	unsealPassphrase string
)

var unsealCmd = &cobra.Command{
	Use:   "unseal",
	Short: "Unseal a Vault instance",
	Long:  "Decrypt unseal keys and submit them to Vault to unseal it",
	RunE:  runUnseal,
}

func init() {
	unsealCmd.Flags().StringVar(&unsealAddr, "addr", "https://vault.bankrut.lan", "Vault address")
	unsealCmd.Flags().StringVar(&unsealKeysFile, "keys-file", os.ExpandEnv("$HOME/.config/vault/bankrut-unseal-keys.gpg"), "Path to GPG-encrypted unseal keys")
	unsealCmd.Flags().StringVar(&unsealPassphrase, "passphrase", "", "GPG passphrase (read from stdin if not provided)")
}

func runUnseal(cmd *cobra.Command, args []string) error {
	keys, err := gpg.DecryptKeys(unsealKeysFile, unsealPassphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt keys: %w", err)
	}

	if len(keys) == 0 {
		return fmt.Errorf("no unseal keys found in %s", unsealKeysFile)
	}

	client := vault.NewClient(unsealAddr)

	for i, key := range keys {
		if key == "" {
			continue
		}

		fmt.Fprintf(os.Stderr, "Unsealing (key %d/%d)...\n", i+1, len(keys))

		resp, err := client.Unseal(cmd.Context(), key)
		if err != nil {
			return fmt.Errorf("unseal failed at key %d: %w", i+1, err)
		}

		fmt.Fprintf(os.Stderr, "  sealed: %v  progress: %d/%d\n", resp.Sealed, resp.Progress, resp.T)

		if !resp.Sealed {
			fmt.Fprintf(os.Stderr, "Vault unsealed after %d keys\n", i+1)
			return nil
		}
	}

	// If we get here, vault is still sealed
	statusResp, err := client.Status(cmd.Context())
	if err == nil {
		fmt.Fprintf(os.Stderr, "Vault still sealed after %d keys. Progress: %d/%d\n", len(keys), statusResp.Progress, statusResp.T)
	}

	return nil
}
