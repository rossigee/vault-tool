package otp

import (
	"encoding/base64"
	"fmt"
)

// DecodeToken decodes an OTP-wrapped root token using the one-time password
// This follows Vault's documented procedure (github.com/hashicorp/vault/sdk/helper/roottoken)
func DecodeToken(encodedToken, otp string) (string, error) {
	if len(otp) == 0 {
		return "", fmt.Errorf("OTP cannot be empty")
	}

	// Decode base64 encoded token (Vault uses RawStdEncoding without padding)
	tokenBytes, err := base64.RawStdEncoding.DecodeString(encodedToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 token: %w", err)
	}

	// The OTP is XORed with the token for wrapping/unwrapping
	// Both must be the same length for XOR operation
	if len(tokenBytes) != len(otp) {
		return "", fmt.Errorf("token and OTP length mismatch: token=%d bytes, otp=%d characters", len(tokenBytes), len(otp))
	}

	// XOR the token with OTP to unwrap it
	decodedToken := make([]byte, len(tokenBytes))
	for i := range tokenBytes {
		decodedToken[i] = tokenBytes[i] ^ otp[i]
	}

	// Return the decoded token directly (Vault does not use checksums)
	return string(decodedToken), nil
}
