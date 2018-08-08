package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/spaolacci/murmur3"
)

func MurmurHash(args ...string) string {
	h32 := murmur3.New32()
	h32.Write([]byte(prepareHashArgs(args...)))

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, h32.Sum32())
	if err != nil {
		panic(fmt.Errorf("cannot make hashsum for `%v`: %s", args, err))
	}

	return fmt.Sprintf("%x", buf.Bytes())
}

func Sha256Hash(args ...string) string {
	sum := sha256.Sum256([]byte(prepareHashArgs(args...)))
	return fmt.Sprintf("%x", sum)
}

func prepareHashArgs(args ...string) string {
	return strings.Join(args, ":::")
}
