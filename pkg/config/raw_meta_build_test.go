package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("rawMetaBuild", func() {
	DescribeTable("toMetaBuild",
		func(raw *rawMetaBuild, expectation MetaBuild) {
			Expect(raw.toMetaBuild()).To(Equal(expectation))
		},
		Entry("should work with RawImageSpec=nil",
			&rawMetaBuild{
				CacheVersion: "some-cache-token",
				Platform:     []string{"linux"},
				Staged:       true,
			},
			MetaBuild{
				CacheVersion: "some-cache-token",
				Platform:     []string{"linux"},
				Staged:       true,
			},
		),
		Entry("should work with RawImageSpec != nil",
			&rawMetaBuild{
				CacheVersion: "some-cache-token",
				Platform:     []string{"linux"},
				Staged:       true,
				RawImageSpec: &rawImageSpecGlobal{},
			},
			MetaBuild{
				CacheVersion: "some-cache-token",
				Platform:     []string{"linux"},
				Staged:       true,
				ImageSpec:    new(rawImageSpecGlobal).toDirective(),
			},
		),
		Entry("should work with RawSbom=nil",
			&rawMetaBuild{
				CacheVersion: "some-cache-token",
				Platform:     []string{"linux"},
				Staged:       true,
			},
			MetaBuild{
				CacheVersion: "some-cache-token",
				Platform:     []string{"linux"},
				Staged:       true,
			},
		),
		Entry("should work with RawSbom != nil",
			&rawMetaBuild{
				CacheVersion: "some-cache-token",
				Platform:     []string{"linux"},
				Staged:       true,
				RawSbom:      new(rawSbom),
			},
			MetaBuild{
				CacheVersion: "some-cache-token",
				Platform:     []string{"linux"},
				Staged:       true,
				Sbom:         new(rawSbom).toDirective(),
			},
		),
	)
})
