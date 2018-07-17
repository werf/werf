package config

import (
	"fmt"
	"testing"
)

func TestRawGit_getNameFromUrl(t *testing.T) {
	var positiveExpectations = []struct {
		url  string
		name string
	}{
		{
			"git@github.com:company/name.git",
			"company/name",
		},
		{
			"https://github.com/company/name.git",
			"company/name",
		},
	}

	for _, expectation := range positiveExpectations {
		rawGit := RawGit{Url: expectation.url}

		name, err := rawGit.getNameFromUrl()
		if err != nil {
			t.Fatal(err)
		}

		if name != expectation.name {
			t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", expectation.name, name)
		}
	}

	var negativeExpectations = []string{
		"git@github.com:company/name",
	}

	for _, expectation := range negativeExpectations {
		rawGit := RawGit{Url: expectation}
		_, err := rawGit.getNameFromUrl()
		expectedError := fmt.Sprintf("Cannot determine repo name from `url: %s`: url is not fit `.*?([^:/ ]+/[^/ ]+)\\.git$` regex!", expectation)
		if err == nil {
			t.Errorf("\n[EXPECTED]: %s", expectedError)
		} else if err.Error() != expectedError {
			t.Errorf("\n[EXPECTED]: %s\n[GOT]: %s", expectedError, err.Error())
		}
	}
}
