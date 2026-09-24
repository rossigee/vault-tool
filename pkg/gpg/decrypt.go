package gpg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func DecryptKeys(keyFilePath, passphrase string) ([]string, error) {
	// Use gpg command-line tool for decryption since it handles keyring management
	// This is simpler and more reliable than using crypto/openpgp
	cmd := exec.Command("gpg", "--quiet", "--decrypt", keyFilePath)

	// Suppress stdout from command
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	// If passphrase is provided, pass it via stdin
	if passphrase != "" {
		stdin, _ := cmd.StdinPipe()
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, passphrase+"\n")
		}()
	} else {
		// Connect user's terminal for passphrase prompt
		cmd.Stdin = os.Stdin
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start gpg: %w", err)
	}

	// Read decrypted output
	decrypted, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted output: %w", err)
	}

	// Check for errors
	if err := cmd.Wait(); err != nil {
		errMsg, _ := io.ReadAll(stderr)
		return nil, fmt.Errorf("gpg decryption failed: %w (stderr: %s)", err, string(errMsg))
	}

	// Split into lines and filter empty ones
	var keys []string
	scanner := bufio.NewScanner(strings.NewReader(string(decrypted)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			keys = append(keys, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading decrypted content: %w", err)
	}

	return keys, nil
}


