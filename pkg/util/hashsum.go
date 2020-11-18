package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/werf/lockgate/pkg/util"

	"github.com/spaolacci/murmur3"
	"golang.org/x/crypto/sha3"
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

func prepareHashArgs(args ...string) string {
	return strings.Join(args, ":::")
}

func ObjectToHashKey(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Sprintf("unable to marshal object %#v: %s", obj, err))
	}
	return util.Sha256Hash(string(data))
}
