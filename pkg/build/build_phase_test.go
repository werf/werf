package build

import (
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildPhase", func() {
	DescribeTable("image environment variable names", func(werfImageName, imageName, expectedImageEnv, expectedPrefix string) {
		Expect(GenerateImageEnv(werfImageName, imageName)).To(Equal(expectedImageEnv))
		Expect(imageEnvPrefix(werfImageName)).To(Equal(expectedPrefix))
	},
		Entry("plus signs", "libstdc++", "registry.example.com/libstdc++:v1", "WERF_LIBSTDC___DOCKER_IMAGE_NAME=registry.example.com/libstdc++:v1", "WERF_LIBSTDC___"),
	)

	Describe("stage digest mutex lifecycle", func() {
		var (
			digestMutex *sync.Mutex

			simulateCalculateStage func(shouldError bool) (bool, func(), error)
		)

		BeforeEach(func() {
			digestMutex = &sync.Mutex{}

			simulateCalculateStage = func(shouldError bool) (bool, func(), error) {
				digestMutex.Lock()
				if shouldError {
					return false, digestMutex.Unlock, fmt.Errorf("simulated error after lock")
				}
				return true, digestMutex.Unlock, nil
			}
		})

		type onImageStageVariant struct {
			name          string
			callCleanup   bool
			expectRelease bool
		}

		DescribeTable("cleanupFunc handling on error path",
			func(v onImageStageVariant) {
				simulateOnImageStage := func(shouldError bool) error {
					var cleanupFunc func()

					err := func() error {
						var err error
						_, cleanupFunc, err = simulateCalculateStage(shouldError)
						if err != nil {
							return err
						}
						return nil
					}()
					if err != nil {
						if v.callCleanup && cleanupFunc != nil {
							cleanupFunc()
						}
						return err
					}

					if cleanupFunc != nil {
						defer cleanupFunc()
					}
					return nil
				}

				err := simulateOnImageStage(true)
				Expect(err).To(HaveOccurred())

				done := make(chan struct{})
				// The buggy variant leaves the mutex locked on purpose, so this
				// goroutine blocks; it takes the mutex through a local of its
				// own, because the variable is reassigned for the next spec.
				mutex := digestMutex
				go func() {
					mutex.Lock()
					mutex.Unlock()
					close(done)
				}()

				if v.expectRelease {
					Eventually(done, 3*time.Second).Should(BeClosed())
				} else {
					Consistently(done, 500*time.Millisecond).ShouldNot(BeClosed())
					mutex.Unlock()
					Eventually(done, 3*time.Second).Should(BeClosed())
				}
			},
			Entry("buggy: cleanupFunc not called on error path leaks mutex", onImageStageVariant{
				name:          "buggy",
				callCleanup:   false,
				expectRelease: false,
			}),
			Entry("fixed: cleanupFunc called on error path releases mutex", onImageStageVariant{
				name:          "fixed",
				callCleanup:   true,
				expectRelease: true,
			}),
		)
	})
})
