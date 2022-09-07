package util

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func LookupBoolEnvironment(environmentName string) (*bool, bool) {
	value, isSet := os.LookupEnv(environmentName)
	if !isSet {
		return nil, false
	}

	switch value {
	case "1", "true", "yes":
		t := true
		return &t, true
	case "0", "false", "no":
		f := false
		return &f, true
	}
	return nil, true
}

func GetBoolEnvironment(environmentName string) *bool {
	val, _ := LookupBoolEnvironment(environmentName)
	return val
}

func GetBoolEnvironmentDefaultFalse(environmentName string) bool {
	switch os.Getenv(environmentName) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func GetBoolEnvironmentDefaultTrue(environmentName string) bool {
	switch os.Getenv(environmentName) {
	case "0", "false", "no":
		return false
	default:
		return true
	}
}

func GetFirstExistingEnvVarAsString(envNames ...string) string {
	for _, envName := range envNames {
		if v := os.Getenv(envName); v != "" {
			return v
		}
	}

	return ""
}

func PredefinedValuesByEnvNamePrefix(envNamePrefix string, envNamePrefixesToExcept ...string) []string {
	var result []string

	env := os.Environ()
	sort.Strings(env)

environLoop:
	for _, keyValue := range env {
		parts := strings.SplitN(keyValue, "=", 2)
		if strings.HasPrefix(parts[0], envNamePrefix) {
			for _, exceptEnvNamePrefix := range envNamePrefixesToExcept {
				if strings.HasPrefix(parts[0], exceptEnvNamePrefix) {
					continue environLoop
				}
			}

			result = append(result, parts[1])
		}
	}

	return result
}

func GetInt64EnvVar(varName string) (*int64, error) {
	if v := os.Getenv(varName); v != "" {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value %q: %w", varName, v, err)
		}

		res := new(int64)
		*res = vInt

		return res, nil
	}

	return nil, nil
}

func GetIntEnvVar(varName string) (*int64, error) {
	if v := os.Getenv(varName); v != "" {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value %q: %w", varName, v, err)
		}

		res := new(int64)
		*res = vInt

		return res, nil
	}

	return nil, nil
}

func GetUint64EnvVar(varName string) (*uint64, error) {
	if v := os.Getenv(varName); v != "" {
		vUint, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value %q: %w", varName, v, err)
		}

		res := new(uint64)
		*res = vUint

		return res, nil
	}

	return nil, nil
}
