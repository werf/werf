package buildah_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/werf/werf/pkg/util"

	"github.com/werf/werf/pkg/docker"

	"github.com/werf/werf/pkg/werf"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/buildah"
)

func TestBuildah(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildah suite")
}

var _ = BeforeSuite(func() {
	Expect(werf.Init("", "")).To(Succeed())
	Expect(docker.Init(context.Background(), "", false, false, "")).To(Succeed())
})

var _ = Describe("Buildah client", func() {
	var b buildah.Buildah

	BeforeEach(func() {
		Skip("not working inside go test yet")

		var err error
		b, err = buildah.NewBuildah(buildah.ModeDockerWithFuse, buildah.BuildahOpts{})
		Expect(err).To(Succeed())
	})

	It("Builds projects using Dockerfile", func() {
		for _, desc := range []struct {
			DockerfilePath string
			ContextPath    string
		}{
			{"./buildah_test/Dockerfile.ghost", ""},
			{"./buildah_test/Dockerfile.neovim", ""},
			{"./buildah_test/app1/Dockerfile", "./buildah_test/app1"},
			{"./buildah_test/app2/Dockerfile", "./buildah_test/app2"},
			{"./buildah_test/app3/Dockerfile", "./buildah_test/app3"},
		} {
			errCh := make(chan error, 0)
			buildDoneCh := make(chan string, 0)

			d, err := ioutil.ReadFile(desc.DockerfilePath)
			Expect(err).To(Succeed())

			var c io.Reader
			if desc.ContextPath != "" {
				c = util.BufferedPipedWriterProcess(func(w io.WriteCloser) {
					if err := util.WriteDirAsTar((desc.ContextPath), w); err != nil {
						errCh <- fmt.Errorf("unable to write dir %q as tar: %s", desc.ContextPath, err)
						return
					}

					if err := w.Close(); err != nil {
						errCh <- fmt.Errorf("unable to close buffered piped writer for context dir %q: %s", desc.ContextPath, err)
						return
					}
				})
			}

			go func() {
				imageID, err := b.BuildFromDockerfile(context.Background(), d, buildah.BuildFromDockerfileOpts{
					CommonOpts: buildah.CommonOpts{LogWriter: os.Stdout},
					ContextTar: c,
				})
				if err != nil {
					errCh <- fmt.Errorf("BuildFromDockerfile failed: %s", err)
					return
				}

				buildDoneCh <- imageID
				close(buildDoneCh)
			}()

			select {
			case err := <-errCh:
				close(errCh)
				Expect(err).NotTo(HaveOccurred())

			case imageID := <-buildDoneCh:
				fmt.Fprintf(os.Stdout, "INFO: built imageId is %s\n", imageID)
			}
		}
	})
})
