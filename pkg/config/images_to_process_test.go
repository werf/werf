package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockImage struct {
	name         string
	final        bool
	stapel       bool
	platform     []string
	cacheVersion string
}

func (m *mockImage) GetName() string {
	return m.name
}

func (m *mockImage) IsFinal() bool {
	return m.final
}

func (m *mockImage) IsStapel() bool {
	return m.stapel
}

func (m *mockImage) Platform() []string {
	return m.platform
}

func (m *mockImage) CacheVersion() string {
	return m.cacheVersion
}

func (m *mockImage) dependsOn() DependsOn {
	return DependsOn{}
}

func (m *mockImage) rawDoc() *doc {
	return nil
}

var _ = Describe("ImagesToProcess", func() {
	var werfConfig *WerfConfig

	BeforeEach(func() {
		werfConfig = &WerfConfig{
			images: []ImageInterface{
				&mockImage{name: "backend", final: true, stapel: false},
				&mockImage{name: "frontend", final: true, stapel: false},
				&mockImage{name: "db", final: false, stapel: true},
				&mockImage{name: "redis", final: false, stapel: true},
				&mockImage{name: "app1", final: true, stapel: false},
				&mockImage{name: "app2", final: true, stapel: false},
			},
		}
	})

	Describe("NewImagesToProcess", func() {
		Context("when withoutImages is true", func() {
			It("should return WithoutImages=true regardless of other params", func() {
				result, err := NewImagesToProcess(werfConfig, []string{"*"}, true, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.WithoutImages).To(BeTrue())
				Expect(result.ImageNameList).To(BeEmpty())
				Expect(result.FinalImageNameList).To(BeEmpty())
			})
		})

		Context("when imageNameList is empty", func() {
			It("should include all images by default", func() {
				result, err := NewImagesToProcess(werfConfig, []string{}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("backend", "frontend", "db", "redis", "app1", "app2"))
				Expect(result.FinalImageNameList).To(ConsistOf("backend", "frontend", "app1", "app2"))
			})

			It("should respect exclusion patterns even with empty include list", func() {
				result, err := NewImagesToProcess(werfConfig, []string{"!db", "!redis"}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("backend", "frontend", "app1", "app2"))
			})
		})

		Context("standard behavior without glob patterns", func() {
			It("should process exact image names", func() {
				result, err := NewImagesToProcess(werfConfig, []string{"backend", "db"}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("backend", "db"))
				Expect(result.FinalImageNameList).To(ConsistOf("backend"))
			})

			It("should return error for non-existing image", func() {
				_, err := NewImagesToProcess(werfConfig, []string{"nonexistent"}, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no image matches pattern"))
			})
		})

		Context("with image exclusion", func() {
			It("should exclude single image", func() {
				result, err := NewImagesToProcess(werfConfig, []string{"!db"}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("backend", "frontend", "redis", "app1", "app2"))
			})

			It("should exclude multiple images", func() {
				result, err := NewImagesToProcess(werfConfig, []string{"!db", "!redis"}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("backend", "frontend", "app1", "app2"))
			})

			It("should handle exclusion only patterns", func() {
				result, err := NewImagesToProcess(werfConfig, []string{"!db", "!redis"}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("backend", "frontend", "app1", "app2"))
			})
		})

		Context("with glob patterns", func() {
			It("should match images by prefix", func() {
				result, err := NewImagesToProcess(werfConfig, []string{"app*"}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("app1", "app2"))
				Expect(result.FinalImageNameList).To(ConsistOf("app1", "app2"))
			})

			It("should match images by suffix", func() {
				werfConfig.images = append(werfConfig.images,
					&mockImage{name: "test-backend", final: true},
					&mockImage{name: "test-frontend", final: true},
				)
				result, err := NewImagesToProcess(werfConfig, []string{"*end"}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("backend", "frontend", "test-backend", "test-frontend"))
				Expect(result.FinalImageNameList).To(ConsistOf("backend", "frontend", "test-backend", "test-frontend"))
			})
		})

		Context("with onlyFinal flag", func() {
			It("should return only final images when onlyFinal=true", func() {
				result, err := NewImagesToProcess(werfConfig, []string{}, true, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("backend", "frontend", "app1", "app2"))
				Expect(result.FinalImageNameList).To(ConsistOf("backend", "frontend", "app1", "app2"))
			})

			It("should return all matching images when onlyFinal=false", func() {
				result, err := NewImagesToProcess(werfConfig, []string{}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("backend", "frontend", "db", "redis", "app1", "app2"))
			})
		})

		Context("with invalid patterns", func() {
			It("should return error for invalid pattern", func() {
				_, err := NewImagesToProcess(werfConfig, []string{"[invalid"}, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid pattern"))
			})
		})
	})

	Describe("parsePattern", func() {
		It("should parse inclusion pattern", func() {
			include, exclude, isExclusion := parsePattern("pattern")
			Expect(include).To(Equal("pattern"))
			Expect(exclude).To(BeEmpty())
			Expect(isExclusion).To(BeFalse())
		})

		It("should parse exclusion pattern", func() {
			include, exclude, isExclusion := parsePattern("!pattern")
			Expect(include).To(BeEmpty())
			Expect(exclude).To(Equal("pattern"))
			Expect(isExclusion).To(BeTrue())
		})
	})
})
