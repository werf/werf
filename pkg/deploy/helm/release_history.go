package helm

import (
	"encoding/json"
	"fmt"
)

type ReleaseHistory []ReleaseHistoryRecord

type ReleaseHistoryRecord struct {
	Revision    int    `json:"revision"`
	Updated     string `json:"updated"`
	Status      string `json:"status"`
	Chart       string `json:"chart"`
	Description string `json:"description"`
}

func GetReleaseHistory(releaseName string) (ReleaseHistory, error) {
	stdout, stderr, err := HelmCmd("history", releaseName, "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("%s %s\n%s", stdout, stderr, err)
	}

	res := ReleaseHistory{}

	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		return nil, fmt.Errorf("bad release json:\n%s\n%s", stdout, err)
	}

	return res, nil
}
