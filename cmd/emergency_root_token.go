package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"git.golder.lan/rossgolderltd/vault-tool/internal/logger"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/gpg"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/otp"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/vault"
)

var (
	emergencyAddr        string
	emergencyKeysFile    string
	emergencyPassphrase  string
	emergencyTokenTTL    string
	emergencyQuiet       bool
	emergencyUseOTP      bool
)

var emergencyRootTokenCmd = &cobra.Command{
	Use:   "emergency-root-token",
	Short: "Generate an emergency root token",
	Long:  "Use unseal keys to generate a temporary root token for emergency maintenance when OIDC auth is unavailable",
	PreRun: func(cmd *cobra.Command, args []string) {
		logger.SetQuiet(emergencyQuiet)
	},
	RunE: runEmergencyRootToken,
}

func init() {
	emergencyRootTokenCmd.Flags().StringVar(&emergencyAddr, "addr", "https://vault.bankrut.lan", "Vault address")
	emergencyRootTokenCmd.Flags().StringVar(&emergencyKeysFile, "keys-file", os.ExpandEnv("$HOME/.config/vault/bankrut-unseal-keys.gpg"), "Path to GPG-encrypted unseal keys")
	emergencyRootTokenCmd.Flags().StringVar(&emergencyPassphrase, "passphrase", "", "GPG passphrase (read from stdin if not provided)")
	emergencyRootTokenCmd.Flags().StringVar(&emergencyTokenTTL, "ttl", "24h", "Token TTL (default 24 hours)")
	emergencyRootTokenCmd.Flags().BoolVarP(&emergencyQuiet, "quiet", "q", false, "Suppress all logging output")
	emergencyRootTokenCmd.Flags().BoolVar(&emergencyUseOTP, "use-otp", false, "Return OTP-wrapped token separately for secure distribution (normally auto-decoded)")
}

func runEmergencyRootToken(cmd *cobra.Command, args []string) error {
	logger.Info("Starting emergency root token generation", "addr", emergencyAddr, "keys_file", emergencyKeysFile)

	logger.Info("Decrypting unseal keys from GPG file", "file", emergencyKeysFile)
	keys, err := gpg.DecryptKeys(emergencyKeysFile, emergencyPassphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt keys: %w", err)
	}

	if len(keys) == 0 {
		return fmt.Errorf("no unseal keys found in %s", emergencyKeysFile)
	}
	logger.Info("Decrypted unseal keys", "count", len(keys))

	// Validate TTL is parseable
	logger.Info("Validating TTL format", "ttl", emergencyTokenTTL)
	_, err = time.ParseDuration(emergencyTokenTTL)
	if err != nil {
		return fmt.Errorf("invalid TTL: %w", err)
	}
	logger.Info("TTL is valid", "ttl", emergencyTokenTTL)

	client := vault.NewClient(emergencyAddr)

	logger.Info("Initializing root token generation process")
	initResp, err := client.InitRootTokenGeneration(cmd.Context(), emergencyTokenTTL)
	if err != nil {
		return fmt.Errorf("failed to initialize root token generation: %w", err)
	}

	logger.Info("Root token generation initialized", "nonce", initResp.Nonce, "threshold", initResp.T, "total", initResp.N)
	logger.Info("Submitting unseal keys to complete generation", "keys_needed", initResp.T, "total_keys", len(keys))

	for i, key := range keys {
		if key == "" {
			continue
		}

		logger.Info("Submitting unseal key", "key_number", i+1, "total_keys", len(keys))

		updateResp, err := client.UpdateRootTokenGeneration(cmd.Context(), initResp.Nonce, key)
		if err != nil {
			return fmt.Errorf("failed to update root token generation at key %d: %w", i+1, err)
		}

		logger.Info("Key submitted successfully", "progress", updateResp.Progress, "required", updateResp.Required)

		if updateResp.Complete {
			logger.Info("Root token generation completed successfully")

			if emergencyUseOTP && initResp.OTP != "" {
				// User explicitly requested OTP-wrapped output
				fmt.Printf("OTP=%s\n", initResp.OTP)
				fmt.Printf("ENCODED_TOKEN=%s\n", updateResp.EncodedToken)
				fmt.Fprintf(os.Stderr, "\n✓ Token generation complete with OTP protection\n")
				fmt.Fprintf(os.Stderr, "To decode the token, run:\n")
				fmt.Fprintf(os.Stderr, "  vault-tool decode-token ENCODED_TOKEN --otp OTP\n")
			} else if initResp.OTP != "" {
				// OTP is available but user wants plaintext - decode automatically
				logger.Info("Decoding OTP-wrapped token automatically")
				decodedToken, err := otp.DecodeToken(updateResp.EncodedToken, initResp.OTP)
				if err != nil {
					return fmt.Errorf("failed to decode OTP-wrapped token: %w", err)
				}
				fmt.Println(decodedToken)
			} else {
				// No OTP protection - token is plaintext
				fmt.Println(updateResp.EncodedToken)
			}
			return nil
		}
	}

	return fmt.Errorf("root token generation incomplete after submitting all keys")
}
