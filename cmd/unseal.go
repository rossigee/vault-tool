package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"git.golder.lan/rossgolderltd/vault-tool/internal/logger"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/gpg"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/vault"
)

var (
	unsealAddr       string
	unsealKeysFile   string
	unsealPassphrase string
	unsealQuiet      bool
)

var unsealCmd = &cobra.Command{
	Use:   "unseal",
	Short: "Unseal a Vault instance",
	Long:  "Decrypt unseal keys and submit them to Vault to unseal it",
	PreRun: func(cmd *cobra.Command, args []string) {
		logger.SetQuiet(unsealQuiet)
	},
	RunE: runUnseal,
}

func init() {
	unsealCmd.Flags().StringVar(&unsealAddr, "addr", "https://vault.bankrut.lan", "Vault address")
	unsealCmd.Flags().StringVar(&unsealKeysFile, "keys-file", os.ExpandEnv("$HOME/.config/vault/bankrut-unseal-keys.gpg"), "Path to GPG-encrypted unseal keys")
	unsealCmd.Flags().StringVar(&unsealPassphrase, "passphrase", "", "GPG passphrase (read from stdin if not provided)")
	unsealCmd.Flags().BoolVarP(&unsealQuiet, "quiet", "q", false, "Suppress all logging output")
}

func runUnseal(cmd *cobra.Command, args []string) error {
	logger.Info("Starting Vault unseal process", "addr", unsealAddr, "keys_file", unsealKeysFile)

	logger.Info("Decrypting unseal keys from GPG file", "file", unsealKeysFile)
	keys, err := gpg.DecryptKeys(unsealKeysFile, unsealPassphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt keys: %w", err)
	}

	if len(keys) == 0 {
		return fmt.Errorf("no unseal keys found in %s", unsealKeysFile)
	}
	logger.Info("Decrypted unseal keys", "count", len(keys))

	client := vault.NewClient(unsealAddr)

	for i, key := range keys {
		if key == "" {
			continue
		}

		logger.Info("Submitting unseal key", "key_number", i+1, "total_keys", len(keys))

		resp, err := client.Unseal(cmd.Context(), key)
		if err != nil {
			return fmt.Errorf("unseal failed at key %d: %w", i+1, err)
		}

		logger.Info("Unseal key accepted", "sealed", resp.Sealed, "progress", resp.Progress, "threshold", resp.T)

		if !resp.Sealed {
			logger.Info("Vault unsealed successfully", "keys_used", i+1)
			return nil
		}
	}

	// If we get here, vault is still sealed
	statusResp, err := client.Status(cmd.Context())
	if err == nil {
		logger.Info("Vault still sealed after all keys", "progress", statusResp.Progress, "threshold", statusResp.T)
	}

	return nil
}
