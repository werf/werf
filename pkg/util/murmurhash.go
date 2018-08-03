package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	
	"github.com/spaolacci/murmur3"
)

func MurmurHash(value string) string {
	h32 := murmur3.New32()
	h32.Write([]byte(value))

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, h32.Sum32())
	if err != nil {
		panic(fmt.Errorf("cannot make hashsum for %s: %s", value, err))
	}

	return fmt.Sprintf("%x", buf.Bytes())
}
