package signing

import (
	"archive/tar"
	"bytes"
	"context"
	"debug/elf"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	"github.com/werf/werf/v2/pkg/tmp_manager"
	werfExec "github.com/werf/werf/v2/pkg/werf/exec"
)

// ELF state machine removed in favor of debug/elf

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

func isELFFileStream(readerAt io.ReaderAt) (bool, error) {
	ef, err := elf.NewFile(readerAt)
	if err != nil {

		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return false, nil
		}

		if _, isFormatError := err.(*elf.FormatError); isFormatError {
			// not ELF
			return false, nil
		}
		return false, err
	}
	defer ef.Close()

	switch ef.Machine {
	case elf.EM_386, elf.EM_X86_64:
		// Good ELF
		return true, nil
	default:
		// Bad ELF
		return false, nil
	}
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

		cmd := werfExec.CommandContextCancellation(ctx, "bsign", "-N", "-s", "--pgoptions="+pgOptionsString, path)
		cmd.Env = append(os.Environ(), cmdExtraEnv...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return formatBsignError(path, output, err)
		}
	}

	return nil
}

var bsignExitCodeMessages = map[int]string{
	1:  "permission denied - insufficient privilege for operation",
	2:  "file not found - specified file not found",
	12: "no memory - memory allocation failed",
	21: "is directory - argument is a directory and must not be",
	22: "invalid argument - invalid command line argument",
	24: "too many open files - pipe call failed",
	26: "file busy - unable to rewrite file as it is in use",
	28: "no space - output device full, no space for new file",
	36: "name too long - pathname too long",
	64: "no hash - hash missing and was expected",
	65: "no signature - signature missing and was expected",
	66: "bad hash - hash failed verification",
	67: "bad signature - signature failed verification",
	68: "unsupported file type - type of file not (yet) supported",
	69: "bad passphrase - passphrase given to gpg was incorrect (check PGPPrivateKeyPassphrase)",
	70: "rewrite failed - error rewriting file",
	71: "quit - premature application termination",
	72: "program not found - exec failed because program wasn't found (check gpg installation)",
}

func formatBsignError(path string, output []byte, err error) error {
	baseMsg := fmt.Sprintf("bsign sign %q failed", path)

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode := exitErr.ExitCode()
		if msg, exists := bsignExitCodeMessages[exitCode]; exists {
			return fmt.Errorf("%s: %s: %w\n\n%s", baseMsg, msg, err, output)
		}
		return fmt.Errorf("%s: unknown exit code %d: %w\n\n%s", baseMsg, exitCode, err, output)
	}

	return fmt.Errorf("%s: %w\n\n%s", baseMsg, err, output)
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

		switch header.Typeflag {
		case tar.TypeReg:
			// For regular files, sign the content if it is an ELF file
			if err := func() error {
				tmpFile, err := tmp_manager.TempFile("elf-file-*.tmp")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()

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

				if !isELF {
					// Reset the file offset to the beginning
					if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
						return fmt.Errorf("failed to seek temp file: %w", err)
					}

					if err := writeTarFile(tarWriter, header, tmpFile); err != nil {
						return fmt.Errorf("failed to write tar file: %w", err)
					}

					return nil
				}

				if err := logboek.Context(ctx).Default().LogProcessInline(header.Name).DoError(func() error {
					if err := signELFFile(ctx, tmpFile.Name(), elfSigningOptions); err != nil {
						return fmt.Errorf("failed to sign ELF file %q: %w", header.Name, err)
					}

					return nil
				}); err != nil {
					return err
				}

				// We must create a new file descriptor to read actual file size and content
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

				if err := writeTarFile(tarWriter, header, signedFile); err != nil {
					return fmt.Errorf("failed to write tar file: %w", err)
				}

				return nil
			}(); err != nil {
				return nil, err
			}
		default:
			// For non-regular files, copy the header and content as-is
			if err := writeTarFile(tarWriter, header, tarReader); err != nil {
				return nil, fmt.Errorf("failed to write tar file: %w", err)
			}
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}

	return &buffer, nil
}

func writeTarFile(tarWriter *tar.Writer, header *tar.Header, body io.Reader) error {
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar file header: %w", err)
	}
	if _, err := io.Copy(tarWriter, body); err != nil {
		return fmt.Errorf("failed to write tar file body: %w", err)
	}
	return nil
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

	var deferredActions []func()
	defer func() {
		for _, action := range deferredActions {
			action()
		}
	}()

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
		deferredActions = append(deferredActions, func() { _ = rc.Close() })

		modifiedLayerBuffer, err := mutateELFFiles(ctx, rc, elfSigningOptions)
		if err != nil {
			return "", fmt.Errorf("failed to mutate ELF files: %w", err)
		}

		tmpFile, err := tmp_manager.TempFile("layer-*.tmp")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		deferredActions = append(deferredActions, func() { _ = tmpFile.Close() })
		deferredActions = append(deferredActions, func() { _ = os.Remove(tmpFile.Name()) })

		if _, err := io.Copy(tmpFile, modifiedLayerBuffer); err != nil {
			return "", fmt.Errorf("failed to write to temp file: %w", err)
		}

		newLayer, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
			return os.Open(tmpFile.Name())
		})
		if err != nil {
			return "", fmt.Errorf("failed to create layer from opener: %w", err)
		}
		newLayers = append(newLayers, mutate.Addendum{
			Layer: newLayer,
		})

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
