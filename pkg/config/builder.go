package config

import (
	"fmt"
	"math/rand"
	"time"
)

func LoadConfig(DappfilePath string) (interface{}, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	if rand.Float64() > 0.5 {
		return nil, fmt.Errorf("BADBADBAD")
	} else {
		return map[string]map[string]string{"hello": map[string]string{"world": "74"}}, nil
	}
}
