package util

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"

	"github.com/google/uuid"

	"github.com/spaolacci/murmur3"
)

func MurmurHash(args ...string) string {
	h32 := murmur3.New32()
	h32.Write([]byte(prepareHashArgs(args...)))
	sum := h32.Sum32()
	return fmt.Sprintf("%x", sum)
}

func Sha3_224Hash(args ...string) string {
	sum := sha3.Sum224([]byte(prepareHashArgs(args...)))
	return fmt.Sprintf("%x", sum)
}

func Sha256Hash(args ...string) string {
	sum := sha256.Sum256([]byte(prepareHashArgs(args...)))
	return fmt.Sprintf("%x", sum)
}

func MD5Hash(args ...string) string {
	h := md5.New()
	h.Write([]byte(prepareHashArgs(args...)))
	return hex.EncodeToString(h.Sum(nil))
}

func prepareHashArgs(args ...string) string {
	return strings.Join(args, ":::")
}

func UUIDToShortString(id uuid.UUID) string {
	return hex.EncodeToString(id[:])
}
