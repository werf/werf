package deploy

import (
	"os"
	"strings"
	"testing"
)

func TestGetSecret_env(t *testing.T) {
	os.Setenv("DAPP_SECRET_KEY", "de470824107b4818acc3d626d67181a9")
	s, err := GetSecret("")
	os.Clearenv()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if s == nil {
		t.Errorf("secret not found")
	}
}

func TestGetSecret_projectDir(t *testing.T) {
	projectDir := "testdata/.dapp"
	s, err := GetSecret(projectDir)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if s == nil {
		t.Errorf("secret not found")
	}
}

func TestGetSecret_home(t *testing.T) {
	os.Setenv("HOME", "testdata")
	s, err := GetSecret("")
	os.Clearenv()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if s == nil {
		t.Errorf("secret not found")
	}
}

func TestGetSecret_negative(t *testing.T) {
	expectedErrorPrefix := "encryption key not found in: "

	_, err := GetSecret("")
	if err == nil {
		t.Error("expected error")
	} else if !strings.HasPrefix(err.Error(), expectedErrorPrefix) {
		t.Errorf("\n[EXPECTED]: %s...\n[GOT] %s\n", expectedErrorPrefix, err.Error())
	}

	expectedErrorPrefix = "check encryption key: "

	os.Setenv("DAPP_SECRET_KEY", "bad key")
	_, err = GetSecret("")
	os.Clearenv()
	if err == nil {
		t.Error("expected error")
	} else if !strings.HasPrefix(err.Error(), expectedErrorPrefix) {
		t.Errorf("\n[EXPECTED]: %s: ...\n[GOT] %s\n", expectedErrorPrefix, err.Error())
	}
}
