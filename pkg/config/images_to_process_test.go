package config

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("ImagesToProcess", func() {
	var (
		mockCtrl     *gomock.Controller
		werfConfig   *WerfConfig
		mockBackend  *MockImageInterface
		mockFrontend *MockImageInterface
		mockDb       *MockImageInterface
		mockRedis    *MockImageInterface
		mockApp1     *MockImageInterface
		mockApp2     *MockImageInterface
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		werfConfig = new(WerfConfig)

		mockBackend = NewMockImageInterface(mockCtrl)
		mockFrontend = NewMockImageInterface(mockCtrl)
		mockDb = NewMockImageInterface(mockCtrl)
		mockRedis = NewMockImageInterface(mockCtrl)
		mockApp1 = NewMockImageInterface(mockCtrl)
		mockApp2 = NewMockImageInterface(mockCtrl)

		mockBackend.EXPECT().GetName().Return("backend").AnyTimes()
		mockBackend.EXPECT().IsFinal().Return(true).AnyTimes()

		mockFrontend.EXPECT().GetName().Return("frontend").AnyTimes()
		mockFrontend.EXPECT().IsFinal().Return(true).AnyTimes()

		mockDb.EXPECT().GetName().Return("db").AnyTimes()
		mockDb.EXPECT().IsFinal().Return(false).AnyTimes()

		mockRedis.EXPECT().GetName().Return("redis").AnyTimes()
		mockRedis.EXPECT().IsFinal().Return(false).AnyTimes()

		mockApp1.EXPECT().GetName().Return("app1").AnyTimes()
		mockApp1.EXPECT().IsFinal().Return(true).AnyTimes()

		mockApp2.EXPECT().GetName().Return("app2").AnyTimes()
		mockApp2.EXPECT().IsFinal().Return(true).AnyTimes()

		werfConfig.images = []ImageInterface{
			mockBackend, mockFrontend, mockDb, mockRedis, mockApp1, mockApp2,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
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

		Context("with exact image names", func() {
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
		})

		Context("with glob patterns", func() {
			It("should match images by prefix", func() {
				result, err := NewImagesToProcess(werfConfig, []string{"app*"}, false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ImageNameList).To(ConsistOf("app1", "app2"))
				Expect(result.FinalImageNameList).To(ConsistOf("app1", "app2"))
			})

			It("should match images by suffix", func() {
				// Добавляем тестовые образы
				mockTestBackend := NewMockImageInterface(mockCtrl)
				mockTestBackend.EXPECT().GetName().Return("test-backend").AnyTimes()
				mockTestBackend.EXPECT().IsFinal().Return(true).AnyTimes()

				mockTestFrontend := NewMockImageInterface(mockCtrl)
				mockTestFrontend.EXPECT().GetName().Return("test-frontend").AnyTimes()
				mockTestFrontend.EXPECT().IsFinal().Return(true).AnyTimes()

				werfConfig.images = append(werfConfig.images, mockTestBackend, mockTestFrontend)

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
			It("should return error for invalid glob syntax", func() {
				_, err := NewImagesToProcess(werfConfig, []string{"[invalid"}, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid pattern"))
			})

			It("should reject recursive glob (**)", func() {
				_, err := NewImagesToProcess(werfConfig, []string{"**"}, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("recursive glob (**) not supported"))
			})

			It("should reject too long patterns", func() {
				longPattern := strings.Repeat("a", 101)
				_, err := NewImagesToProcess(werfConfig, []string{longPattern}, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("pattern too long"))
			})

			It("should reject special characters", func() {
				_, err := NewImagesToProcess(werfConfig, []string{"pattern&"}, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("special characters not allowed"))
			})
		})
	})

	Describe("parsePattern", func() {
		It("should parse inclusion pattern", func() {
			include, exclude, isExclusion, err := parsePattern("pattern")
			Expect(err).NotTo(HaveOccurred())
			Expect(include).To(Equal("pattern"))
			Expect(exclude).To(BeEmpty())
			Expect(isExclusion).To(BeFalse())
		})

		It("should parse exclusion pattern", func() {
			include, exclude, isExclusion, err := parsePattern("!pattern")
			Expect(err).NotTo(HaveOccurred())
			Expect(include).To(BeEmpty())
			Expect(exclude).To(Equal("pattern"))
			Expect(isExclusion).To(BeTrue())
		})

		It("should reject recursive glob (**)", func() {
			_, _, _, err := parsePattern("**")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("recursive glob (**) not supported"))
		})

		It("should reject too long patterns", func() {
			longPattern := strings.Repeat("a", 101)
			_, _, _, err := parsePattern(longPattern)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("pattern too long"))
		})

		It("should reject special characters", func() {
			_, _, _, err := parsePattern("pat|tern")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("special characters not allowed"))
		})
	})
})
