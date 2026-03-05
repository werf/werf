package common

import (
	"os"
	"testing"
)

func TestProcessEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		initialEnv  string // value of cmdData.Environment before ProcessEnvironment
		werfEnvVar  string // value of WERF_ENV environment variable
		expectedEnv string // expected value of cmdData.Environment after ProcessEnvironment
	}{
		{
			name:        "empty --env flag with WERF_ENV set should use WERF_ENV",
			initialEnv:  "",
			werfEnvVar:  "production",
			expectedEnv: "production",
		},
		{
			name:        "empty --env flag without WERF_ENV should remain empty",
			initialEnv:  "",
			werfEnvVar:  "",
			expectedEnv: "",
		},
		{
			name:        "non-empty --env flag should be preserved even with WERF_ENV set",
			initialEnv:  "staging",
			werfEnvVar:  "production",
			expectedEnv: "staging",
		},
		{
			name:        "non-empty --env flag should be preserved without WERF_ENV",
			initialEnv:  "development",
			werfEnvVar:  "",
			expectedEnv: "development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Preserve original WERF_ENV and restore it after the subtest
			origWerfEnv, hadWerfEnv := os.LookupEnv("WERF_ENV")
			t.Cleanup(func() {
				if hadWerfEnv {
					os.Setenv("WERF_ENV", origWerfEnv)
				} else {
					os.Unsetenv("WERF_ENV")
				}
			})

			// Setup test environment
			cmdData := &CmdData{
				Environment: tt.initialEnv,
			}

			// Set or unset WERF_ENV for this test case
			if tt.werfEnvVar != "" {
				os.Setenv("WERF_ENV", tt.werfEnvVar)
			} else {
				os.Unsetenv("WERF_ENV")
			}

			// Execute
			ProcessEnvironment(cmdData)

			// Verify
			if cmdData.Environment != tt.expectedEnv {
				t.Errorf("ProcessEnvironment() = %q, want %q", cmdData.Environment, tt.expectedEnv)
			}
		})
	}
}
