package utils

import (
	"fmt"
	"os"
)

func GetRequiredEnv(name string) string {
	envValue := os.Getenv(name)
	if envValue == "" {
		panic(fmt.Sprintf("environment variable %s must be specified", name))
	}

	return envValue
}
