package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"git.golder.lan/rossgolderltd/vault-tool/internal/logger"
	"git.golder.lan/rossgolderltd/vault-tool/pkg/otp"
)

var (
	decodeTokenOTP     string
	decodeTokenVerbose bool
	decodeTokenDebug   bool
)

var decodeTokenCmd = &cobra.Command{
	Use:   "decode-token <encoded-token>",
	Short: "Decode an OTP-wrapped root token",
	Long:  "Decode a root token that was generated with OTP protection. The OTP can be provided via --otp flag or will be prompted if not provided.",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if decodeTokenDebug {
			logger.SetDebug()
		} else if decodeTokenVerbose {
			logger.SetVerbose()
		}
	},
	RunE: runDecodeToken,
}

func init() {
	decodeTokenCmd.Flags().StringVar(&decodeTokenOTP, "otp", "", "One-time password (will prompt if not provided)")
	decodeTokenCmd.Flags().BoolVarP(&decodeTokenVerbose, "verbose", "v", false, "Enable verbose (info-level) logging")
	decodeTokenCmd.Flags().BoolVarP(&decodeTokenDebug, "debug", "d", false, "Enable debug-level logging")
}

func runDecodeToken(cmd *cobra.Command, args []string) error {
	encodedToken := args[0]

	logger.Info("Decoding OTP-wrapped root token")

	// Get OTP if not provided
	otpPassword := decodeTokenOTP
	if otpPassword == "" {
		logger.Info("OTP not provided, prompting for input")
		fmt.Fprint(os.Stderr, "Enter OTP: ")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			return fmt.Errorf("failed to read OTP from stdin: %w", err)
		}
		otpPassword = input
	}

	logger.Info("Decoding token using OTP")
	decodedToken, err := otp.DecodeToken(encodedToken, otpPassword)
	if err != nil {
		return fmt.Errorf("failed to decode token: %w", err)
	}

	logger.Info("Token decoded successfully")
	fmt.Println(decodedToken)
	return nil
}
