package deploy

import (
	"bytes"
	"io/ioutil"
	"testing"
)

type SecretMock struct{}

func (s *SecretMock) Generate(data []byte) ([]byte, error) {
	return []byte("encoded data"), nil
}

func (s *SecretMock) Extract(data []byte) ([]byte, error) {
	return []byte("data"), nil
}

func TestSecretGenerator_GenerateYamlData(t *testing.T) {
	gEncode, err := NewSecretEncodeGenerator(&SecretMock{})
	if err != nil {
		t.Fatal(err)
	}

	fileData, err := ioutil.ReadFile("testdata/secret.yaml")
	if err != nil {
		t.Fatal(err)
	}

	encodedData, err := gEncode.GenerateYamlData(fileData)
	if err != nil {
		t.Fatal(err)
	}

	gDecode, err := NewSecretDecodeGenerator(&SecretMock{})
	if err != nil {
		t.Fatal(err)
	}

	resultData, err := gDecode.GenerateYamlData(encodedData)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(fileData, resultData) {
		t.Errorf("\n[EXPECTED]\n%s\n[GOT]\n%s\n", string(fileData), string(resultData))
	}
}
