package secret

type Secret interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(encodedData []byte) ([]byte, error)
}

func NewSecret(key []byte) (Secret, error) {
	s, err := NewAesSecret(key)
	if err != nil {
		return nil, err
	}

	return s, nil
}
