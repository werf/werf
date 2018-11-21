package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

type AesSecret struct {
	CipherBlock cipher.Block
}

func GenerateAexSecretKey() ([]byte, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}

	result := []byte(hex.EncodeToString(randomBytes))

	return result, nil
}

func NewAesSecret(key []byte) (*AesSecret, error) {
	key, err := hexToBinary(key)
	if err != nil {
		return nil, err
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	secret := &AesSecret{c}
	return secret, nil
}

func (s *AesSecret) Generate(data []byte) ([]byte, error) {
	dataToEncrypt := pad([]byte(data))

	cipherData := make([]byte, aes.BlockSize+len(dataToEncrypt))
	iv := cipherData[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(s.CipherBlock, iv)
	mode.CryptBlocks(cipherData[aes.BlockSize:], dataToEncrypt)

	ivSize := make([]byte, 2)
	binary.LittleEndian.PutUint16(ivSize, aes.BlockSize)

	var args []byte
	args = append(args, ivSize...)
	args = append(args, cipherData...)

	result := make([]byte, hex.EncodedLen(len(args)))
	hex.Encode(result, args)

	return result, nil
}

func (s *AesSecret) Extract(data []byte) ([]byte, error) {
	dataToExtract, err := hexToBinary(data)
	if err != nil {
		return nil, err
	}

	ivLengthInfoSize := 2
	ivSize := aes.BlockSize
	paddingMaxSize := aes.BlockSize
	minimalDataBinarySize := ivLengthInfoSize + ivSize + paddingMaxSize
	minimalDataSize := minimalDataBinarySize * 2
	if len(dataToExtract) < minimalDataBinarySize { // iv + padding
		return nil, fmt.Errorf("minimum required data length: '%v'", minimalDataSize)
	}

	iv := dataToExtract[ivLengthInfoSize : ivLengthInfoSize+ivSize]
	cipherText := dataToExtract[ivLengthInfoSize+ivSize:]

	if len(cipherText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("data isn't a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(s.CipherBlock, iv)
	mode.CryptBlocks(cipherText, cipherText)

	result, err := unpad(cipherText)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func pad(data []byte) []byte {
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func unpad(data []byte) ([]byte, error) {
	length := len(data)
	unpadding := int(data[length-1])

	if unpadding > length {
		return nil, fmt.Errorf("inconsistent data, unpad failed")
	}

	return data[:(length - unpadding)], nil
}

func hexToBinary(data []byte) ([]byte, error) {
	result := make([]byte, hex.DecodedLen(len(data)))
	if _, err := hex.Decode(result, data); err != nil {
		return nil, err
	}

	return result, nil
}
