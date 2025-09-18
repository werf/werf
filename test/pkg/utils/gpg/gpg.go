package gpg

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// ImportKey imports secret or (public) key
func ImportKey(ctx context.Context, keyData []byte) error {
	cmd := exec.CommandContext(ctx, "gpg", "--import")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	_, err = stdin.Write(keyData)
	if err != nil {
		return fmt.Errorf("failed to write key data to stdin: %v", err)
	}

	if err = stdin.Close(); err != nil {
		return fmt.Errorf("failed to close stdin: %v", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to import key: %v, output: %s", err, string(output))
	}
	return nil
}

func SafeDeleteKey(ctx context.Context, fingerprint string) error {
	cmd := exec.CommandContext(ctx, "gpg", "--batch", "--yes", "--delete-key", fingerprint)

	output, err := cmd.CombinedOutput()
	if bytes.Contains(output, []byte("Not found")) {
		return nil // Key not found, nothing to delete
	} else if err != nil {
		return fmt.Errorf("failed to delete key: %w, output: %s", err, string(output))
	}

	return nil
}

func SafeDeleteSecretKey(ctx context.Context, fingerprintID string) error {
	cmd := exec.CommandContext(ctx, "gpg", "--batch", "--yes", "--delete-secret-key", fingerprintID)

	output, err := cmd.CombinedOutput()
	if bytes.Contains(output, []byte("Not found")) {
		return nil // Key not found, nothing to delete
	} else if err != nil {
		return fmt.Errorf("failed to delete secret key: %w, output: %s", err, string(output))
	}

	return nil
}
