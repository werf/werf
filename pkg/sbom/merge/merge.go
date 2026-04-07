package merge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/opencontainers/go-digest"
	"github.com/samber/lo"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/sbom/convert"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
	"github.com/werf/werf/v2/pkg/sbom/image"
)

type Options struct {
	Input        string
	IsprasFormat string
	AppName      string
	AppVersion   string
	Manufacturer string
	Output       string
}

func Run(ctx context.Context, registry docker_registry.Interface, repo string, opts Options) error {
	if err := ValidateOptions(opts); err != nil {
		return err
	}

	var (
		mapping map[string]string
		images  []*convert.ImageSBOM
		result  *cdx.BOM
	)

	err := logboek.Context(ctx).Default().LogProcess("Reading input mapping").DoError(func() error {
		var err error
		mapping, err = ReadInputMapping(opts.Input)
		if err != nil {
			return fmt.Errorf("unable to read input mapping: %w", err)
		}
		logboek.Context(ctx).Default().LogFDetails("%d image(s) found\n", len(mapping))
		return nil
	})
	if err != nil {
		return err
	}

	err = logboek.Context(ctx).Default().LogProcess("Pulling and parsing SBOMs").DoError(func() error {
		var err error
		images, err = PullAndParseImages(ctx, registry, repo, mapping)
		return err
	})
	if err != nil {
		return err
	}

	err = logboek.Context(ctx).Default().LogProcess("Merging SBOMs into %s format", opts.IsprasFormat).DoError(func() error {
		assembler, err := convert.NewAssembler(opts.IsprasFormat)
		if err != nil {
			return fmt.Errorf("unable to select assembler: %w", err)
		}

		total := len(images)
		for i, img := range images {
			logboek.Context(ctx).Default().LogF("[%d/%d] %s\n", i+1, total, img.Name)
		}

		result, err = (&convert.Converter{Assembler: assembler}).Convert(ctx, images, convert.ProductMeta{
			AppName:      opts.AppName,
			AppVersion:   opts.AppVersion,
			Manufacturer: opts.Manufacturer,
		})
		if err != nil {
			return fmt.Errorf("unable to convert: %w", err)
		}

		logboek.Context(ctx).Default().LogFDetails(
			"result: %d component(s), %d dependenc(ies)\n",
			lo.Ternary(result.Components != nil, len(*result.Components), 0),
			lo.Ternary(result.Dependencies != nil, len(*result.Dependencies), 0),
		)
		return nil
	})
	if err != nil {
		return err
	}

	return logboek.Context(ctx).Default().LogProcess("Writing output").DoError(func() error {
		jsonBytes, err := cyclonedxutil.ToJSON(result)
		if err != nil {
			return fmt.Errorf("unable to serialize result: %w", err)
		}

		dest := opts.Output
		if dest == "" {
			dest = "stdout"
		}
		logboek.Context(ctx).Default().LogFDetails("destination: %s\n", dest)

		return WriteOutput(jsonBytes, opts.Output)
	})
}

func ValidateOptions(opts Options) error {
	var missing []string

	if opts.Input == "" {
		missing = append(missing, "--input")
	}
	if opts.IsprasFormat == "" {
		missing = append(missing, "--ispras-format")
	}
	if opts.AppName == "" {
		missing = append(missing, "--app-name")
	}
	if opts.AppVersion == "" {
		missing = append(missing, "--app-version")
	}
	if opts.Manufacturer == "" {
		missing = append(missing, "--manufacturer")
	}

	if len(missing) > 0 {
		return fmt.Errorf("required flag(s) not set: %v", missing)
	}

	if opts.IsprasFormat != "oss" && opts.IsprasFormat != "container" {
		return fmt.Errorf("--ispras-format must be \"oss\" or \"container\", got %q", opts.IsprasFormat)
	}

	return nil
}

func ReadInputMapping(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %q: %w", path, err)
	}

	var mapping map[string]string
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("unable to parse JSON from %q: %w", path, err)
	}

	if len(mapping) == 0 {
		return nil, fmt.Errorf("empty mapping in %q", path)
	}

	if err := ValidateInputMapping(mapping); err != nil {
		return nil, fmt.Errorf("unable to validate mapping in %q: %w", path, err)
	}

	return mapping, nil
}

func ValidateInputMapping(mapping map[string]string) error {
	for imageName, imageDigest := range mapping {
		if strings.TrimSpace(imageName) == "" {
			return fmt.Errorf("image name must not be empty")
		}
		if _, err := digest.Parse(imageDigest); err != nil {
			return fmt.Errorf("image %q has invalid digest %q: %w", imageName, imageDigest, err)
		}
	}

	return nil
}

func PullAndParseImages(ctx context.Context, registry docker_registry.Interface, repo string, mapping map[string]string) ([]*convert.ImageSBOM, error) {
	total := len(mapping)
	images := make([]*convert.ImageSBOM, 0, total)

	logboek.Context(ctx).Default().LogF("Pulling SBOMs: %d image(s)\n", total)

	idx := 0
	for imageName, imageDigest := range mapping {
		idx++
		reference := fmt.Sprintf("%s@%s", repo, imageDigest)

		err := logboek.Context(ctx).Default().
			LogProcess("[%d/%d] %s", idx, total, imageName).
			DoError(func() error {
				logboek.Context(ctx).Default().LogFDetails("reference: %s\n", reference)

				bom, err := image.PullCycloneDX16BOM(ctx, registry, reference)
				if err != nil {
					return fmt.Errorf("unable to pull SBOM for %q (%s): %w", imageName, reference, err)
				}

				img, err := convert.NewImageSBOMFromCycloneDX16(ctx, imageName, bom)
				if err != nil {
					return fmt.Errorf("unable to parse SBOM for %q: %w", imageName, err)
				}

				images = append(images, img)
				return nil
			})
		if err != nil {
			return nil, err
		}
	}

	return images, nil
}

func WriteOutput(data []byte, outputPath string) error {
	if outputPath == "" {
		return logboek.Streams().DoErrorWithoutProxyStreamDataFormatting(func() error {
			_, err := os.Stdout.Write(data)
			return err
		})
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("write output to %q: %w", outputPath, err)
	}

	return nil
}
