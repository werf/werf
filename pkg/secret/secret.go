package secret

type Secret interface {
	Generate(data []byte) ([]byte, error)
	Extract(encodedData []byte) ([]byte, error)
}

func NewSecret(key []byte) (Secret, error) {
	s, err := NewAesSecret(key)
	if err != nil {
		return nil, err
	}

	return s, nil
}
