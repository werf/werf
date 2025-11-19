package signature

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/deckhouse/delivery-kit-sdk/pkg/signature/elf/inhouse"
	"github.com/deckhouse/delivery-kit-sdk/pkg/signature/image"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/signature/elf/bsign"
	elfTar "github.com/werf/werf/v2/pkg/signature/elf/tar"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/util/parallel"
)

type VerifyOptions struct {
	VerifyManifest      bool
	VerifyELFFiles      bool
	VerifyBSignELFFiles bool
	Roots               []string
	References          []string
}

func Verify(ctx context.Context, options VerifyOptions, parallelOptions parallel.DoTasksOptions) error {
	refsCount := len(options.References)

	return parallel.DoTasks(ctx, refsCount, parallelOptions, func(ctx context.Context, taskId int) error {
		imageRef := options.References[taskId]

		return logboek.Context(ctx).Default().LogProcess("Verifying image (%d/%d)", taskId+1, refsCount).DoError(func() error {
			logboek.Context(ctx).Default().LogF("Using reference: %s\n", imageRef)

			img, err := loadImage(ctx, imageRef)
			if err != nil {
				return err
			}

			if options.VerifyManifest {
				if err = verifyManifest(ctx, options.Roots, img); err != nil {
					return err
				}
			}

			if options.VerifyELFFiles || options.VerifyBSignELFFiles {
				if err = verifyELFFiles(ctx, options.Roots, img, options.VerifyELFFiles, options.VerifyBSignELFFiles); err != nil {
					return err
				}
			}

			return nil
		})
	})
}

func loadImage(ctx context.Context, imageRef string) (v1.Image, error) {
	desc, err := docker_registry.API().GetRepoImageDesc(ctx, imageRef)
	if err != nil {
		return nil, fmt.Errorf("unable to get image description: %w", err)
	}

	img, err := desc.Image()
	if err != nil {
		return nil, fmt.Errorf("unable to get image: %w", err)
	}

	return img, nil
}

func verifyManifest(ctx context.Context, roots []string, img v1.Image) error {
	manifest, err := img.Manifest()
	if err != nil {
		return fmt.Errorf("unable to get manifest: %w", err)
	}

	return logboek.Context(ctx).Default().LogProcessInline("Manifest signature").DoError(func() error {
		if err = image.VerifyImageManifestSignature(ctx, roots, manifest); err != nil {
			return fmt.Errorf("unable to verify image manifest signature: %w", err)
		}

		logboek.Context(ctx).Default().LogF(" ok")
		return nil
	})
}

func verifyELFFiles(ctx context.Context, roots []string, img v1.Image, useInHouseVerification, useBSignVerification bool) error {
	rc := mutate.Extract(img)
	defer rc.Close()

	return logboek.Context(ctx).Default().LogProcess("ELF files signatures").DoError(func() error {
		elfTarReader := elfTar.NewReader(tar.NewReader(rc))

		for {
			header, err := elfTarReader.Next()
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return fmt.Errorf("unable to read tar file: %w", err)
			}

			if !header.IsELF {
				continue
			}

			if err = logboek.Context(ctx).Default().LogProcessInline(header.Name).DoError(func() error {
				tmpFile, err := tmp_manager.TempFile("elf-file-*.tmp")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()

				if _, err := io.Copy(tmpFile, elfTarReader); err != nil {
					return fmt.Errorf("failed to write to temp file: %w", err)
				}

				if useInHouseVerification {
					if err := inhouse.Verify(ctx, roots, tmpFile.Name()); err != nil {
						return err
					}
				}
				if useBSignVerification {
					if err = bsign.Verify(ctx, roots, tmpFile.Name()); err != nil {
						return err
					}
				}

				logboek.Context(ctx).Default().LogF(" ok")
				return nil
			}); err != nil {
				return fmt.Errorf("ELF signature verification failed for %s: %w", header.Name, err)
			}
		}

		return nil
	})
}
