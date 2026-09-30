package e2e_export_test

import (
	"context"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/gomega"
)

func imageDiffIDs(ctx context.Context, reference string) []v1.Hash {
	ref, err := name.ParseReference(reference)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	img, err := remote.Image(ref, remote.WithContext(ctx))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	config, err := img.ConfigFile()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return config.RootFS.DiffIDs
}
