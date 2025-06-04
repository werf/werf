package pass

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"golang.org/x/term"
)

const (
	VariablePassword = "WERF_PASSWORD"
)

// PassFunc is the function to be called to retrieve the signer password. If
// nil, then it assumes that no password is provided.
type PassFunc func(bool) ([]byte, error)

func GetPass(confirm bool) ([]byte, error) {
	read := readPasswordFn(confirm)
	return read()
}

func readPasswordFn(confirm bool) func() ([]byte, error) {
	pw, ok := os.LookupEnv(VariablePassword)
	switch {
	case ok:
		return func() ([]byte, error) {
			return []byte(pw), nil
		}
	case isTerminal():
		return func() ([]byte, error) {
			return getPassFromTerm(confirm)
		}
	// Handle piped in passwords.
	default:
		return func() ([]byte, error) {
			return io.ReadAll(os.Stdin)
		}
	}
}

func getPassFromTerm(confirm bool) ([]byte, error) {
	fmt.Fprint(os.Stderr, "Enter password for private key: ")
	// Unnecessary convert of syscall.Stdin on *nix, but Windows is a uintptr
	//nolint:unconverted
	pw1, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(os.Stderr)
	if !confirm {
		return pw1, nil
	}
	fmt.Fprint(os.Stderr, "Enter password for private key again: ")
	// Unnecessary convert of syscall.Stdin on *nix, but Windows is a uintptr
	//nolint:unconverted
	confirmpw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return nil, err
	}

	if string(pw1) != string(confirmpw) {
		return nil, errors.New("passwords do not match")
	}
	return pw1, nil
}

func isTerminal() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}
