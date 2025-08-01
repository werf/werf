package signing

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/containers/buildah/docker"
	"github.com/deckhouse/delivery-kit-sdk/pkg/signature/elf/inhouse"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/werf/exec"
)

type ELFState int

const (
	Initial ELFState = iota
	Seen7F
	Seen45
	Seen4C
	ELF
	NotELF
)

type ELFSigningOptions struct {
	InHouseEnabled bool
	BsignEnabled   bool

	PGPPrivateKeyFingerprint string
	PGPPrivateKeyPassphrase  string

	signer *Signer
}

func (o ELFSigningOptions) Enabled() bool {
	return o.InHouseEnabled || o.BsignEnabled
}

func (o ELFSigningOptions) Signer() *Signer {
	return o.signer
}

func NewELFSigningOptions(signer *Signer) ELFSigningOptions {
	return ELFSigningOptions{
		signer: signer,
	}
}

func isELFFileStream(reader io.Reader) (bool, error) {
	state := Initial
	buffer := make([]byte, 1)

	for {
		_, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}

		b := buffer[0]
		switch state {
		case Initial:
			if b == 0x7f {
				state = Seen7F
			} else {
				state = NotELF
			}
		case Seen7F:
			if b == 0x45 {
				state = Seen45
			} else {
				state = NotELF
			}
		case Seen45:
			if b == 0x4c {
				state = Seen4C
			} else {
				state = NotELF
			}
		case Seen4C:
			if b == 0x46 {
				state = ELF
			} else {
				state = NotELF
			}
		case ELF:
			return true, nil
		case NotELF:
			return false, nil
		}
	}

	return state == ELF, nil
}

func signELFFile(ctx context.Context, path string, elfSigningOptions ELFSigningOptions) error {
	if elfSigningOptions.InHouseEnabled {
		if err := inhouse.Sign(ctx, elfSigningOptions.Signer().SignerVerifier(), path); err != nil {
			return fmt.Errorf("inhouse sign %q: %w", path, err)
		}
	}

	if elfSigningOptions.BsignEnabled {
		var cmdExtraEnv []string
		pgOptionsString := fmt.Sprintf("--batch --default-key=%s", elfSigningOptions.PGPPrivateKeyFingerprint)
		if elfSigningOptions.PGPPrivateKeyPassphrase != "" {
			pgOptionsString += " --pinentry-mode=loopback"
			pgOptionsString += fmt.Sprintf(" --passphrase=$WERF_SERVICE_ELF_PGP_PRIVATE_KEY_PASSPHRASE")
			cmdExtraEnv = append(cmdExtraEnv, fmt.Sprintf("WERF_SERVICE_ELF_PGP_PRIVATE_KEY_PASSPHRASE=%s", elfSigningOptions.PGPPrivateKeyPassphrase))
		}

		cmd := exec.CommandContextCancellation(ctx, "bsign", "-N", "-s", "--pgoptions="+pgOptionsString, path)
		cmd.Env = append(os.Environ(), cmdExtraEnv...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("bsign sign %q: %w", path, err)
		}
	}

	return nil
}

func mutateELFFiles(ctx context.Context, reader io.Reader, elfSigningOptions ELFSigningOptions) (*bytes.Buffer, error) {
	tarReader := tar.NewReader(reader)
	var buffer bytes.Buffer
	tarWriter := tar.NewWriter(&buffer)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// We have reached the end of the archive
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar archive: %w", err)
		}

		if header.Typeflag == tar.TypeReg {
			if err := func() error {
				tmpFile, err := ioutil.TempFile("", "elf-file-*.tmp")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				defer os.Remove(tmpFile.Name())

				// Stream the file content to the temp file
				if _, err := io.Copy(tmpFile, tarReader); err != nil {
					return fmt.Errorf("failed to write to temp file: %w", err)
				}

				// Reset the file offset to the beginning
				if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
					return fmt.Errorf("failed to seek temp file: %w", err)
				}

				isELF, err := isELFFileStream(tmpFile)
				if err != nil {
					return fmt.Errorf("failed to check ELF file %q: %w", header.Name, err)
				}

				if isELF {
					if err := logboek.Context(ctx).Default().LogProcessInline(header.Name).DoError(func() error {
						if err := signELFFile(ctx, tmpFile.Name(), elfSigningOptions); err != nil {
							return fmt.Errorf("failed to sign ELF file %q: %w", header.Name, err)
						}

						return nil
					}); err != nil {
						return err
					}

					signedFile, err := os.Open(tmpFile.Name())
					if err != nil {
						return err
					}
					defer signedFile.Close()

					// Update the header size to the signed file size
					info, err := signedFile.Stat()
					if err != nil {
						return err
					}
					header.Size = info.Size()

					// Reset the file offset to the beginning
					if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
						return fmt.Errorf("failed to seek temp file: %w", err)
					}

					if err := tarWriter.WriteHeader(header); err != nil {
						return fmt.Errorf("failed to write tar header: %w", err)
					}
					if _, err := io.Copy(tarWriter, signedFile); err != nil {
						return fmt.Errorf("failed to write signed file content: %w", err)
					}
				} else {
					// Reset the file offset to the beginning
					if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
						return fmt.Errorf("failed to seek temp file: %w", err)
					}

					if err := tarWriter.WriteHeader(header); err != nil {
						return fmt.Errorf("failed to write tar header: %w", err)
					}
					if _, err := io.Copy(tarWriter, tmpFile); err != nil {
						return fmt.Errorf("failed to write file content: %w", err)
					}
				}

				return nil
			}(); err != nil {
				return nil, err
			}
		} else {
			// For non-regular files, copy the header and content as-is
			if err := tarWriter.WriteHeader(header); err != nil {
				return nil, fmt.Errorf("failed to write tar header: %w", err)
			}
			if _, err := io.Copy(tarWriter, tarReader); err != nil {
				return nil, fmt.Errorf("failed to write file content: %w", err)
			}
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}

	return &buffer, nil
}

func Sign(ctx context.Context, refBase, refFinal string, elfSigningOptions ELFSigningOptions) (string, error) {
	img, err := daemon.Image(IDReference{ID: refBase})
	if err != nil {
		return "", err
	}

	baseCfg, err := img.ConfigFile()
	if err != nil {
		return "", fmt.Errorf("failed to get config file: %w", err)
	}

	if len(baseCfg.RootFS.DiffIDs) == 0 {
		logboek.Context(ctx).Info().LogLn("Skipping empty image")
		return refBase, nil
	}

	// Extract the image layers and mutate ELF files
	layers, err := img.Layers()
	if err != nil {
		return "", err
	}

	extraLabels := map[string]string{}
	var newLayers []mutate.Addendum
	for _, layer := range layers {
		h, err := layer.DiffID()
		if err != nil {
			return "", fmt.Errorf("failed to get layer digest: %w", err)
		}

		if baseCfg.Config.Labels[fmt.Sprintf("werf-elf-signed-layer-%s", h.String())] == "true" {
			logboek.Context(ctx).Debug().LogLn("Skipping already signed layer")
			newLayers = append(newLayers, mutate.Addendum{
				Layer: layer,
			})
			continue
		}

		rc, err := layer.Uncompressed()
		if err != nil {
			return "", fmt.Errorf("failed to get uncompressed layer: %w", err)
		}

		modifiedLayerBuffer, err := mutateELFFiles(ctx, rc, elfSigningOptions)
		if err != nil {
			return "", fmt.Errorf("failed to mutate ELF files: %w", err)
		}

		newLayer, err := tarball.LayerFromReader(io.NopCloser(modifiedLayerBuffer))
		if err != nil {
			return "", fmt.Errorf("failed to create layer from reader: %w", err)
		}

		newLayers = append(newLayers, mutate.Addendum{
			Layer: newLayer,
		})
		rc.Close()

		h, err = newLayer.DiffID()
		if err != nil {
			return "", fmt.Errorf("failed to get new layer digest: %w", err)
		}

		extraLabels[fmt.Sprintf("werf-elf-signed-layer-%s", h.String())] = "true"
	}

	cfg := &v1.ConfigFile{
		Architecture: baseCfg.Architecture,
		Author:       baseCfg.Author,
		Config:       baseCfg.Config,
		OS:           baseCfg.OS,
		Created:      v1.Time{Time: time.Now()},
		OSVersion:    baseCfg.OSVersion,
		Variant:      baseCfg.Variant,
		OSFeatures:   baseCfg.OSFeatures,
		RootFS: v1.RootFS{
			Type: docker.TypeLayers,
		},
	}

	cfg.Config.Labels = util.MergeMaps(cfg.Config.Labels, extraLabels)

	// Create a new image with mutated layers
	newImg, err := mutate.ConfigFile(empty.Image, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to append new layers: %w", err)
	}

	newImg, err = mutate.Append(newImg, newLayers...)
	if err != nil {
		return "", fmt.Errorf("failed to append new layers: %w", err)
	}

	newImg, err = mutate.CreatedAt(newImg, v1.Time{Time: time.Now()})
	if err != nil {
		return "", fmt.Errorf("failed to set time: %w", err)
	}

	// Save the new image with a new tag
	newRef, err := name.NewTag(refFinal)
	if err != nil {
		return "", fmt.Errorf("failed to parse new reference: %w", err)
	}

	if _, err := daemon.Write(newRef, newImg); err != nil {
		return "", fmt.Errorf("failed to write new image: %w", err)
	}

	logboek.Context(ctx).Debug().LogLn("Successfully created new image: %s\n", newRef.Name())

	return refFinal, nil
}

type IDReference struct {
	fmt.Stringer
	ID string
}

// Реализуем метод String() для интерфейса fmt.Stringer.
func (r IDReference) String() string {
	return r.ID
}

// Реализуем метод Context(), который пока не реализован.
func (r IDReference) Context() name.Repository {
	panic("not implemented")
}

// Реализуем метод Identifier(), который пока не реализован.
func (r IDReference) Identifier() string {
	panic("not implemented")
}

// Реализуем метод Name(), который возвращает ID.
func (r IDReference) Name() string {
	return r.ID
}

// Реализуем метод Scope(), который пока не реализован.
func (r IDReference) Scope(s string) string {
	panic("not implemented")
}
