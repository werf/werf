package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spaolacci/murmur3"
	"golang.org/x/crypto/sha3"
	"golang.org/x/mod/sumdb/dirhash"
)

// LegacyMurmurHash function returns a hash of non-fixed length (1-8 symbols)
func LegacyMurmurHash(args ...string) string {
	h32 := murmur3.New32()
	h32.Write([]byte(prepareHashArgs(args...)))
	sum := h32.Sum32()
	return fmt.Sprintf("%x", sum)

	// TODO: use byte slice instead of uint32
	// bytes := (*[4]byte)(unsafe.Pointer(&sum))
	// return fmt.Sprintf("%x", *bytes)
}

func Sha3_224Hash(args ...string) string {
	sum := sha3.Sum224([]byte(prepareHashArgs(args...)))
	return fmt.Sprintf("%x", sum)
}

func Sha256Hash(args ...string) string {
	sum := sha256.Sum256([]byte(prepareHashArgs(args...)))
	return fmt.Sprintf("%x", sum)
}

// For file: hash contents of file with its name.
// For directory: hash contents of all files in directory, along with their relative filenames.
func HashContentsAndPathsRecurse(path string) (string, error) {
	path = filepath.Clean(path)

	fi, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("unable to stat %q: %w", path, err)
	}

	var hash string
	if fi.IsDir() {
		if hash, err = dirhash.HashDir(path, "/", dirhash.Hash1); err != nil {
			return "", fmt.Errorf("unable to calculate hash for dir %q: %w", path, err)
		}
	} else {
		if hash, err = dirhash.Hash1([]string{filepath.Base(path)}, func(_ string) (io.ReadCloser, error) {
			return os.Open(path)
		}); err != nil {
			return "", fmt.Errorf("unable to calculate hash for file %q: %w", path, err)
		}
	}

	return hash, nil
}

func prepareHashArgs(args ...string) string {
	return strings.Join(args, ":::")
}
