package buildah_test

import (
	"context"
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
	Skip("not working inside go test yet")

	var b buildah.Buildah

	BeforeEach(func() {
		var err error
		b, err = buildah.NewBuildah(buildah.ModeDockerWithFuse)
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
			d, c := loadDockerfileAndContext(desc.DockerfilePath, desc.ContextPath)

			_, err := b.BuildFromDockerfile(context.Background(), d, buildah.BuildFromDockerfileOpts{
				CommonOpts: buildah.CommonOpts{LogWriter: os.Stdout},
				ContextTar: c,
			})

			Expect(err).To(Succeed())
		}
	})
})

func loadDockerfileAndContext(dockerfilePath string, contextPath string) ([]byte, io.Reader) {
	data, err := ioutil.ReadFile(dockerfilePath)
	Expect(err).To(Succeed())

	if contextPath == "" {
		return data, nil
	}

	reader := util.ReadDirAsTar(contextPath)

	return data, reader
}
