package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("hasExplicitTagOrDigest", func(ref string, expected bool) {
	Expect(hasExplicitTagOrDigest(ref)).To(Equal(expected))
},
	Entry("bare name", "ubuntu", false),
	Entry("name with tag", "ubuntu:20.04", true),
	Entry("name with digest", "ubuntu@sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", true),
	Entry("registry host:port with no tag", "registry.example.com:5000/base-image", false),
	Entry("registry host:port with tag", "registry.example.com:5000/base-image:latest", true),
	Entry("invalid reference", "", false),
)
