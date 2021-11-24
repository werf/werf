package secret

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"testing"
)

var (
	AesSecretKey      = []byte("11ac8312520b5ff037bae386ea2e8a07")
	supportedKeySizes = []int{16, 24, 32}
)

func TestGenerateAesSecretKey(t *testing.T) {
	key, err := GenerateAesSecretKey()
	if err != nil {
		t.Fatal(err)
	}

	if len(key) != 32 {
		t.Errorf("Got unexpected key %q", key)
	}
}

func TestNewAesSecret_positive(t *testing.T) {
	for _, size := range supportedKeySizes {
		randomBinary := make([]byte, size)
		if _, err := io.ReadFull(rand.Reader, randomBinary); err != nil {
			t.Fatal(err.Error())
		}

		key := []byte(hex.EncodeToString(randomBinary))

		t.Run(fmt.Sprintf("%v|%v", size, string(key)), func(t *testing.T) {
			_, err := NewAesEncoder(key)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestNewAesSecret_negative(t *testing.T) {
	tests := []struct {
		name         string
		key          []byte
		errorMessage string
	}{
		{
			name:         "invalid key size",
			key:          []byte("12"),
			errorMessage: "crypto/aes: invalid key size 1",
		},
		{
			name:         "odd length hex string",
			key:          []byte("1"),
			errorMessage: "encoding/hex: odd length hex string",
		},
		{
			name:         "invalid byte",
			key:          []byte("xx"),
			errorMessage: "encoding/hex: invalid byte: U+0078 'x'",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewAesEncoder(test.key)
			if err == nil {
				t.Errorf("Expected error: %s", test.errorMessage)
			} else if err.Error() != test.errorMessage {
				t.Errorf("\n[EXPECTED]: %s\n[GOT]: %s", test.errorMessage, err.Error())
			}
		})
	}
}

func TestAesSecret_Generate(t *testing.T) {
	s, err := NewAesEncoder(AesSecretKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Encrypt([]byte("flant"))
	if err != nil {
		t.Error(err)
	}
}

func TestAesSecret_Extract_positive(t *testing.T) {
	s, err := NewAesEncoder(AesSecretKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Decrypt([]byte("10000f13a718d019612ab8ad30d9bec8e2c09df0f2d168c179bef954e78371bf6a5a"))
	if err != nil {
		t.Error(err)
	}
}

func TestAesSecret_Extract_negative(t *testing.T) {
	tests := []struct {
		name         string
		encodedData  []byte
		errorMessage string
	}{
		{
			name:         "odd length hex string",
			encodedData:  []byte("1"),
			errorMessage: "encoding/hex: odd length hex string",
		},
		{
			name:         "invalid byte",
			encodedData:  []byte("1x"),
			errorMessage: "encoding/hex: invalid byte: U+0078 'x'",
		},
		{
			name:         "minimum required data length",
			encodedData:  []byte("12"),
			errorMessage: "minimum required data length: '68'",
		},
		{
			name:         "inconsistent data, unpad failed",
			encodedData:  []byte("10000f13a718d019612ab8ad30d9bec8e2c09df0f2d168c179bef954e78371bf6a5b"),
			errorMessage: "inconsistent data, unpad failed",
		},
	}

	s, err := NewAesEncoder(AesSecretKey)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err = s.Decrypt(test.encodedData)
			if err == nil {
				t.Errorf("Expected error: %s", test.errorMessage)
			} else if err.Error() != test.errorMessage {
				t.Errorf("\n[EXPECTED]: %s\n[GOT]: %s", test.errorMessage, err.Error())
			}
		})
	}
}

func TestAesSecret(t *testing.T) {
	tests := []string{"", "value"}

	for _, size := range supportedKeySizes {
		randomBinary := make([]byte, size)
		if _, err := io.ReadFull(rand.Reader, randomBinary); err != nil {
			t.Fatal(err.Error())
		}

		key := []byte(hex.EncodeToString(randomBinary))

		s, err := NewAesEncoder(key)
		if err != nil {
			t.Fatal(err)
		}

		t.Run(fmt.Sprintf("%v|%v", size, string(key)), func(t *testing.T) {
			for _, test := range tests {
				t.Run(test, func(t *testing.T) {
					encodedData, err := s.Encrypt([]byte(test))
					if err != nil {
						t.Fatal(err)
					}

					result, err := s.Decrypt(encodedData)
					if err != nil {
						t.Fatal(err)
					}

					if test != string(result) {
						t.Errorf("\n[EXPECTED]: %s\n[GOT]: %s", test, result)
					}
				})
			}
		})
	}
}
