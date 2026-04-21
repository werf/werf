package docker_registry

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type ResolveImplementationEntry struct {
	imagesRepoAddress string
	expectation       string
}

var _ = DescribeTable("ResolveImplementation", func(entry ResolveImplementationEntry) {
	resolvedImplementation, err := ResolveImplementation(entry.imagesRepoAddress, "")
	Expect(err).ShouldNot(HaveOccurred())

	Expect(resolvedImplementation).Should(Equal(entry.expectation))
},
	Entry("ecr", ResolveImplementationEntry{
		imagesRepoAddress: "123456789012.dkr.ecr.test.amazonaws.com/repo",
		expectation:       "ecr",
	}),
	Entry("acr", ResolveImplementationEntry{
		imagesRepoAddress: "test.azurecr.io/repo",
		expectation:       "acr",
	}),
	Entry("default", ResolveImplementationEntry{
		imagesRepoAddress: "localhost:5000/repo",
		expectation:       "default",
	}),
	Entry("dockerhub", ResolveImplementationEntry{
		imagesRepoAddress: "account/repo",
		expectation:       "dockerhub",
	}),
	Entry("gcr", ResolveImplementationEntry{
		imagesRepoAddress: "container.cloud.google.com/registry/repo",
		expectation:       "gcr",
	}),
	Entry("github", ResolveImplementationEntry{
		imagesRepoAddress: "ghcr.io/package",
		expectation:       "github",
	}),
	Entry("harbor", ResolveImplementationEntry{
		imagesRepoAddress: "harbor.company.com/project/repo",
		expectation:       "harbor",
	}),
	Entry("quay", ResolveImplementationEntry{
		imagesRepoAddress: "quay.io/account/repo",
		expectation:       "quay",
	}),
)

var _ = Describe("trySequentially", func() {
	It("should continue after error and return next successful value", func() {
		calls := []string{}
		value, found, err := trySequentially([]string{"broken", "empty", "good"}, func(item string) (string, bool, error) {
			calls = append(calls, item)
			switch item {
			case "broken":
				return "", false, errors.New("lookup failed")
			case "empty":
				return "", false, nil
			case "good":
				return "value", true, nil
			default:
				return "", false, nil
			}
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(value).To(Equal("value"))
		Expect(calls).To(Equal([]string{"broken", "empty", "good"}))
	})

	It("should skip failing first entry and still use the next mirror", func() {
		calls := []string{}
		value, found, err := trySequentially([]string{"bad-mirror", "good-mirror"}, func(item string) (string, bool, error) {
			calls = append(calls, item)
			if item == "bad-mirror" {
				return "", false, errors.New("dns failure")
			}
			return item, true, nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(value).To(Equal("good-mirror"))
		Expect(calls).To(Equal([]string{"bad-mirror", "good-mirror"}))
	})

	It("should report not found when all entries fail or miss", func() {
		value, found, err := trySequentially([]string{"broken", "empty"}, func(item string) (string, bool, error) {
			switch item {
			case "broken":
				return "", false, errors.New("lookup failed")
			case "empty":
				return "", false, nil
			default:
				return "", false, nil
			}
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())
		Expect(value).To(BeEmpty())
	})
})
