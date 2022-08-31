package slug

import (
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	legacyCaseWithTwoHyphensMaxSize := 48

	tests := []struct {
		name    string
		data    string
		maxSize *int
		result  string
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
			result: strings.Repeat("x", DefaultSlugMaxSize-len("-27e2f02f")) + "-27e2f02f",
		},
		{
			name:    "legacyCaseWithTwoHyphen",
			data:    "postgres-feature-31981-change-delivery-date-del-result",
			maxSize: &legacyCaseWithTwoHyphensMaxSize,
			result:  "postgres-feature-31981-change-delivery--852739dc",
		},
		{
			name:    "legacyCaseWithTwoHyphen_2",
			data:    "php_fpm_exporter-monitoring-dev-encrypt-1",
			maxSize: &legacyCaseWithTwoHyphensMaxSize,
			result:  "php-fpm-exporter-monitoring-dev-encrypt--83286e5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			maxSize := DefaultSlugMaxSize
			if test.maxSize != nil {
				maxSize = *test.maxSize
			}

			result := LimitedSlug(test.data, maxSize)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}

			if len(result) > maxSize {
				t.Errorf("Max size exceeded: [EXPECTED]: %d [GOT]: %d", maxSize, len(result))
			}

			tRunIdempotence(t, test.name, test.data, func(s string) string {
				return LimitedSlug(s, maxSize)
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
			data:   strings.Repeat("x", DockerTagMaxSize+1),
			result: strings.Repeat("x", DockerTagMaxSize-len("-8cca70eb")) + "-8cca70eb",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DockerTag(test.data)
			if test.result != result {
				t.Errorf("\n[EXPECTED]: %s (%d)\n[GOT]: %s (%d)", test.result, len(test.result), result, len(result))
			}

			if len(result) > DockerTagMaxSize {
				t.Errorf("Max size exceeded: [EXPECTED]: %d [GOT]: %d", DockerTagMaxSize, len(result))
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
			name:   "validWithDot",
			data:   "release.name",
			result: "release.name",
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
			result: strings.Repeat("x", helmReleaseMaxSize-len("-18c5dfb9")) + "-18c5dfb9",
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
			name:   "notMatchRegexp_dot",
			data:   "kubernetes.namespace",
			result: "kubernetes-namespace-1507f75c",
		},
		{
			name:   "maxSizeExceeded",
			data:   strings.Repeat("x", kubernetesNamespaceMaxSize+1),
			result: strings.Repeat("x", kubernetesNamespaceMaxSize-len("-afd4efcd")) + "-afd4efcd",
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
