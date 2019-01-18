package slug

import (
	"strings"
	"testing"

	"github.com/flant/werf/pkg/util"
)

var (
	servicePartSize = len(util.MurmurHash()) + len(slugSeparator)
)

func TestSlug(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		result string
	}{
		{
			name:   "empty",
			data:   "",
			result: "",
		},
		{
			name:   "shouldNotBeSlugged",
			data:   "data",
			result: "data",
		},
		{
			name:   "notEqualWithSluggedData",
			data:   "A",
			result: "a-cef7dc54",
		},
		{
			name:   "maxSizeExceeded",
			data:   strings.Repeat("x", slugMaxSize+1),
			result: strings.Repeat("x", slugMaxSize-servicePartSize) + "-2ff0e227",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Slug(test.data)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}
		})
	}
}

func TestDockerTag(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		result string
	}{
		{
			name:   "empty",
			data:   "",
			result: "",
		},
		{
			name:   "notMatchRegexp_startWithDash",
			data:   "-data",
			result: "data-6ed871a8",
		},
		{
			name:   "notMatchRegexp_unsupportedChar",
			data:   "da/ta",
			result: "da-ta-f896fa0a",
		},
		{
			name:   "maxSizeExceeded",
			data:   strings.Repeat("x", dockerTagMaxSize+1),
			result: strings.Repeat("x", dockerTagMaxSize-servicePartSize) + "-eb70ca8c",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DockerTag(test.data)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}
		})
	}
}

func TestHelmRelease(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		result string
	}{
		{
			name:   "empty",
			data:   "",
			result: "",
		},
		{
			name:   "shouldNotBeSlugged",
			data:   "data",
			result: "data",
		},
		{
			name:   "notMatchRegexp_startWithDash",
			data:   "-data",
			result: "data-6ed871a8",
		},
		{
			name:   "notMatchRegexp_unsupportedChar",
			data:   "da/ta",
			result: "da-ta-f896fa0a",
		},
		{
			name:   "maxSizeExceeded",
			data:   strings.Repeat("x", helmReleaseMaxSize+1),
			result: strings.Repeat("x", helmReleaseMaxSize-servicePartSize) + "-b9dfc518",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := HelmRelease(test.data)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}
		})
	}
}

func TestKubernetesNamespace(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		result string
	}{
		{
			name:   "empty",
			data:   "",
			result: "",
		},
		{
			name:   "shouldNotBeSlugged",
			data:   "data",
			result: "data",
		},
		{
			name:   "notMatchRegexp_unsupportedChar",
			data:   "da_ta",
			result: "da-ta-149d333b",
		},
		{
			name:   "maxSizeExceeded",
			data:   strings.Repeat("x", kubernetesNamespaceMaxSize+1),
			result: strings.Repeat("x", kubernetesNamespaceMaxSize-servicePartSize) + "-cdefd4af",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := KubernetesNamespace(test.data)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}
		})
	}
}
