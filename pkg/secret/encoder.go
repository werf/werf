package secret

type Encoder interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(encodedData []byte) ([]byte, error)
}
