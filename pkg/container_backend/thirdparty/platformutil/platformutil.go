package platformutil

import (
	"fmt"
)

func NormalizeUserParams(platformParams []string) ([]string, error) {
	specs, err := Parse(platformParams)
	if err != nil {
		return nil, fmt.Errorf("unable to parse platform specs: %w", err)
	}

	return Format(Dedupe(specs)), nil
}
