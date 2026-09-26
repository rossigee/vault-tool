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
	emergencyTokenUnlimited bool
	emergencyVerbose     bool
	emergencyDebug       bool
	emergencyUseOTP      bool
	emergencyTokenFile   string
)

var emergencyRootTokenCmd = &cobra.Command{
	Use:   "emergency-root-token",
	Short: "Generate an emergency root token",
	Long:  "Use unseal keys to generate a temporary root token for emergency maintenance when OIDC auth is unavailable",
	PreRun: func(cmd *cobra.Command, args []string) {
		if emergencyDebug {
			logger.SetDebug()
		} else if emergencyVerbose {
			logger.SetVerbose()
		}
	},
	RunE: runEmergencyRootToken,
}

func init() {
	emergencyRootTokenCmd.Flags().StringVar(&emergencyAddr, "addr", "https://vault.bankrut.lan", "Vault address")
	emergencyRootTokenCmd.Flags().StringVar(&emergencyKeysFile, "keys-file", os.ExpandEnv("$HOME/.config/vault/bankrut-unseal-keys.gpg"), "Path to GPG-encrypted unseal keys")
	emergencyRootTokenCmd.Flags().StringVar(&emergencyPassphrase, "passphrase", "", "GPG passphrase (read from stdin if not provided)")
	emergencyRootTokenCmd.Flags().StringVar(&emergencyTokenTTL, "ttl", "24h", "Token TTL (default 24 hours, mutually exclusive with --unlimited)")
	emergencyRootTokenCmd.Flags().BoolVar(&emergencyTokenUnlimited, "unlimited", false, "Create token without expiry (mutually exclusive with --ttl)")
	emergencyRootTokenCmd.Flags().StringVarP(&emergencyTokenFile, "file", "f", os.ExpandEnv("$HOME/.vault-token"), "Output file for token (default: ~/.vault-token, use - for stdout)")
	emergencyRootTokenCmd.Flags().BoolVarP(&emergencyVerbose, "verbose", "v", false, "Enable verbose (info-level) logging")
	emergencyRootTokenCmd.Flags().BoolVarP(&emergencyDebug, "debug", "d", false, "Enable debug-level logging")
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

	// Handle TTL or unlimited flag
	var tokenTTL string
	if emergencyTokenUnlimited {
		if emergencyTokenTTL != "24h" {
			return fmt.Errorf("--unlimited and --ttl are mutually exclusive")
		}
		logger.Info("Creating unlimited root token (no expiry)")
		tokenTTL = ""
	} else {
		// Validate TTL is parseable
		logger.Info("Validating TTL format", "ttl", emergencyTokenTTL)
		_, err = time.ParseDuration(emergencyTokenTTL)
		if err != nil {
			return fmt.Errorf("invalid TTL: %w", err)
		}
		logger.Info("TTL is valid", "ttl", emergencyTokenTTL)
		tokenTTL = emergencyTokenTTL
	}

	client := vault.NewClient(emergencyAddr)

	logger.Info("Initializing root token generation process")
	initResp, err := client.InitRootTokenGeneration(cmd.Context(), tokenTTL)
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
				logger.Info("Decoding OTP-wrapped token automatically", "token_len", len(updateResp.EncodedToken), "otp_len", len(initResp.OTP), "token", updateResp.EncodedToken)
				decodedToken, err := otp.DecodeToken(updateResp.EncodedToken, initResp.OTP)
				if err != nil {
					return fmt.Errorf("failed to decode OTP-wrapped token: %w", err)
				}
				writeToken(decodedToken, tokenTTL, emergencyTokenUnlimited)
			} else {
				// No OTP protection - token is plaintext
				writeToken(updateResp.EncodedToken, tokenTTL, emergencyTokenUnlimited)
			}
			return nil
		}
	}

	return fmt.Errorf("root token generation incomplete after submitting all keys")
}

func writeToken(token, ttl string, unlimited bool) error {
	if emergencyTokenFile == "-" {
		fmt.Println(token)
		return nil
	}

	expandedPath := os.ExpandEnv(emergencyTokenFile)
	err := os.WriteFile(expandedPath, []byte(token+"\n"), 0600)
	if err != nil {
		return fmt.Errorf("failed to write token to %s: %w", expandedPath, err)
	}

	fmt.Fprintf(os.Stderr, "\n✓ Root token written to: %s\n", expandedPath)
	if unlimited {
		fmt.Fprintf(os.Stderr, "Token TTL: unlimited (no expiry)\n")
	} else {
		fmt.Fprintf(os.Stderr, "Token TTL: %s\n", ttl)
	}
	return nil
}
