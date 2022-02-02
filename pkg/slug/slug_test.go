package slug

import (
	"strings"
	"testing"

	"github.com/werf/werf/pkg/util"
)

var servicePartSize = len(util.MurmurHash("stub")) + len(slugSeparator)

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
			result: "a-54dcf7ce",
		},
		{
			name:   "maxSizeExceeded",
			data:   strings.Repeat("x", DefaultSlugMaxSize+1),
			result: strings.Repeat("x", DefaultSlugMaxSize-servicePartSize) + "-27e2f02f",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := LimitedSlug(test.data, DefaultSlugMaxSize)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}

			if len(result) > DefaultSlugMaxSize {
				t.Errorf("Max size exceeded: [EXPECTED]: %d [GOT]: %d", DefaultSlugMaxSize, len(result))
			}

			tRunIdempotence(t, test.name, test.data, func(s string) string {
				return LimitedSlug(s, DefaultSlugMaxSize)
			})
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
			result: "data-a871d86e",
		},
		{
			name:   "notMatchRegexp_unsupportedChar",
			data:   "da/ta",
			result: "da-ta-afa96f8",
		},
		{
			name:   "maxSizeExceeded",
			data:   strings.Repeat("x", dockerTagMaxSize+1),
			result: strings.Repeat("x", dockerTagMaxSize-servicePartSize) + "-8cca70eb",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DockerTag(test.data)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}

			if len(result) > dockerTagMaxSize {
				t.Errorf("Max size exceeded: [EXPECTED]: %d [GOT]: %d", dockerTagMaxSize, len(result))
			}
		})

		tRunIdempotence(t, test.name, test.data, DockerTag)
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
			result: "data-a871d86e",
		},
		{
			name:   "notMatchRegexp_unsupportedChar",
			data:   "da/ta",
			result: "da-ta-afa96f8",
		},
		{
			name:   "maxSizeExceeded",
			data:   strings.Repeat("x", helmReleaseMaxSize+1),
			result: strings.Repeat("x", helmReleaseMaxSize-servicePartSize) + "-18c5dfb9",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := HelmRelease(test.data)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}

			if len(result) > helmReleaseMaxSize {
				t.Errorf("Max size exceeded: [EXPECTED]: %d [GOT]: %d", helmReleaseMaxSize, len(result))
			}
		})

		tRunIdempotence(t, test.name, test.data, HelmRelease)
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
			result: "da-ta-3b339d14",
		},
		{
			name:   "maxSizeExceeded",
			data:   strings.Repeat("x", kubernetesNamespaceMaxSize+1),
			result: strings.Repeat("x", kubernetesNamespaceMaxSize-servicePartSize) + "-afd4efcd",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := KubernetesNamespace(test.data)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}

			if len(result) > kubernetesNamespaceMaxSize {
				t.Errorf("Max size exceeded: [EXPECTED]: %d [GOT]: %d", kubernetesNamespaceMaxSize, len(result))
			}
		})

		tRunIdempotence(t, test.name, test.data, KubernetesNamespace)
	}
}

func tRunIdempotence(t *testing.T, testName, testData string, slugger func(string) string) {
	t.Run(testName+"-idempotence", func(t *testing.T) {
		firstResult := slugger(testData)
		secondResult := slugger(firstResult)
		if firstResult != secondResult {
			t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", firstResult, len(firstResult), secondResult, len(secondResult))
		}
	})
}
