package util

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/spaolacci/murmur3"
	"golang.org/x/crypto/sha3"
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

func prepareHashArgs(args ...string) string {
	return strings.Join(args, ":::")
}
