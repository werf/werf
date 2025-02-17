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
	)
})
