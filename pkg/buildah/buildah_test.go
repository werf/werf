package buildah_test

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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

	r, w := io.Pipe()

	go func() {
		tarWriter := tar.NewWriter(w)

		err := filepath.Walk(contextPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error accessing %q: %s", path, err)
			}

			relPath, err := filepath.Rel(contextPath, path)
			if err != nil {
				return err
			}

			if info.Mode().IsDir() {
				header := &tar.Header{
					Name:     relPath,
					Size:     info.Size(),
					Mode:     int64(info.Mode()),
					ModTime:  info.ModTime(),
					Typeflag: tar.TypeDir,
				}

				fmt.Printf("WRITE HEADER %#v\n", header)

				err = tarWriter.WriteHeader(header)
				if err != nil {
					return fmt.Errorf("could not tar write header for %q: %s", path, err)
				}

				return nil
			}

			header := &tar.Header{
				Name:     relPath,
				Size:     info.Size(),
				Mode:     int64(info.Mode()),
				ModTime:  info.ModTime(),
				Typeflag: tar.TypeReg,
			}

			err = tarWriter.WriteHeader(header)
			if err != nil {
				return fmt.Errorf("could not tar write header for %q: %s", path, err)
			}

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("unable to open %q: %s", path, err)
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return fmt.Errorf("unable to write %q into tar: %s", path, err)
			}

			return nil
		})

		if err != nil {
			panic(err.Error())
		}

		if err := tarWriter.Close(); err != nil {
			panic(err.Error())
		}

		if err := w.Close(); err != nil {
			panic(err.Error())
		}
	}()

	return data, r
}
