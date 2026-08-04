package migrate

import (
	"context"
	"testing"
)

func TestRemoveSourceDefaultsToTrue(t *testing.T) {
	t.Setenv("WERF_REMOVE_SOURCE", "")

	cmd := NewCmd(context.Background())
	flag := cmd.Flags().Lookup("remove-source")
	if flag == nil {
		t.Fatal("--remove-source flag must be registered")
	}
	if flag.DefValue != "true" {
		t.Fatalf("--remove-source must default to true, got %q", flag.DefValue)
	}
}
