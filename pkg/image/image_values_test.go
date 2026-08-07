package image

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildImageValuesMap", func() {
	DescribeTable("parses image reference into values map",
		func(ref, digest string, expected map[string]interface{}) {
			getter := NewInfoGetter("myimage", ref, digest, InfoGetterOptions{})
			result, err := BuildImageValuesMap(getter)
			Expect(err).NotTo(HaveOccurred())

			for key, expectedVal := range expected {
				Expect(result).To(HaveKeyWithValue(key, expectedVal), "key %q mismatch", key)
			}
		},

		Entry("standard registry with namespace",
			"ghcr.io/zalando/postgres-operator:v1.15.1",
			"sha256:abc123",
			map[string]interface{}{
				"registry":       "ghcr.io",
				"repository":     "zalando/postgres-operator",
				"namespace":      "zalando",
				"name":           "postgres-operator",
				"tag":            "v1.15.1",
				"digest":         "sha256:abc123",
				"tag_digest":     "v1.15.1@sha256:abc123",
				"image":          "ghcr.io/zalando/postgres-operator",
				"ref":            "ghcr.io/zalando/postgres-operator:v1.15.1@sha256:abc123",
				"ref_tag":        "ghcr.io/zalando/postgres-operator:v1.15.1",
				"repository_ref": "zalando/postgres-operator:v1.15.1@sha256:abc123",
				"repository_tag": "zalando/postgres-operator:v1.15.1",
				"name_ref":       "postgres-operator:v1.15.1@sha256:abc123",
				"name_tag":       "postgres-operator:v1.15.1",
			}),

		Entry("deep namespace path",
			"example.org/apps/team/myapp:a243949601ddc3d4-1598024377816",
			"sha256:def456",
			map[string]interface{}{
				"registry":       "example.org",
				"repository":     "apps/team/myapp",
				"namespace":      "apps/team",
				"name":           "myapp",
				"tag":            "a243949601ddc3d4-1598024377816",
				"digest":         "sha256:def456",
				"image":          "example.org/apps/team/myapp",
				"ref_tag":        "example.org/apps/team/myapp:a243949601ddc3d4-1598024377816",
				"repository_tag": "apps/team/myapp:a243949601ddc3d4-1598024377816",
				"name_tag":       "myapp:a243949601ddc3d4-1598024377816",
			}),

		Entry("Docker Hub implicit registry",
			"index.docker.io/library/nginx:latest",
			"sha256:789abc",
			map[string]interface{}{
				"registry":   "index.docker.io",
				"repository": "library/nginx",
				"namespace":  "library",
				"name":       "nginx",
				"tag":        "latest",
				"digest":     "sha256:789abc",
				"image":      "index.docker.io/library/nginx",
				"ref_tag":    "index.docker.io/library/nginx:latest",
			}),

		Entry("empty digest",
			"ghcr.io/org/app:sometag",
			"",
			map[string]interface{}{
				"registry":       "ghcr.io",
				"repository":     "org/app",
				"namespace":      "org",
				"name":           "app",
				"tag":            "sometag",
				"digest":         "",
				"tag_digest":     "sometag@",
				"ref":            "ghcr.io/org/app:sometag@",
				"ref_tag":        "ghcr.io/org/app:sometag",
				"repository_ref": "org/app:sometag@",
				"repository_tag": "org/app:sometag",
				"name_ref":       "app:sometag@",
				"name_tag":       "app:sometag",
			}),
	)
})

var _ = Describe("BuildStubImageValuesMap", func() {
	It("returns all expected keys with stub placeholders", func() {
		result := BuildStubImageValuesMap("REPO", "TAG")

		Expect(result).To(HaveKeyWithValue("registry", "REGISTRY"))
		Expect(result).To(HaveKeyWithValue("namespace", "NAMESPACE"))
		Expect(result).To(HaveKeyWithValue("name", "NAME"))
		Expect(result).To(HaveKeyWithValue("tag", "TAG"))
		Expect(result).To(HaveKeyWithValue("digest", "DIGEST"))
		Expect(result).To(HaveKeyWithValue("tag_digest", "TAG@DIGEST"))
		Expect(result).To(HaveKeyWithValue("image", "REPO"))
		Expect(result).To(HaveKeyWithValue("repository", "NAMESPACE/NAME"))
		Expect(result).To(HaveKeyWithValue("ref", "REPO:TAG@DIGEST"))
		Expect(result).To(HaveKeyWithValue("ref_tag", "REPO:TAG"))
		Expect(result).To(HaveKeyWithValue("repository_ref", "NAMESPACE/NAME:TAG@DIGEST"))
		Expect(result).To(HaveKeyWithValue("repository_tag", "NAMESPACE/NAME:TAG"))
		Expect(result).To(HaveKeyWithValue("name_ref", "NAME:TAG@DIGEST"))
		Expect(result).To(HaveKeyWithValue("name_tag", "NAME:TAG"))
	})

	It("has same keys as BuildImageValuesMap", func() {
		stubResult := BuildStubImageValuesMap("REPO", "TAG")

		getter := NewInfoGetter("img", "ghcr.io/org/app:v1", "sha256:abc", InfoGetterOptions{})
		realResult, err := BuildImageValuesMap(getter)
		Expect(err).NotTo(HaveOccurred())

		for key := range realResult {
			Expect(stubResult).To(HaveKey(key), "stub missing key %q", key)
		}
		for key := range stubResult {
			Expect(realResult).To(HaveKey(key), "real missing key %q", key)
		}
	})
})
