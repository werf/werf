package common

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"

	"github.com/samber/lo"
	"golang.org/x/crypto/openpgp"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/build/signing"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/build/verify_annotation"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/thirdparty/platformutil"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/slug"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf/exec"
)

func GetConveyorOptions(ctx context.Context, commonCmdData *CmdData, imagesToProcess config.ImagesToProcess) (build.ConveyorOptions, error) {
	conveyorOptions := build.ConveyorOptions{
		LocalGitRepoVirtualMergeOptions: stage.VirtualMergeOptions{
			VirtualMerge: GetVirtualMerge(commonCmdData),
		},
		ImagesToProcess: imagesToProcess,
	}

	if len(commonCmdData.GetPlatform()) > 0 {
		platforms, err := platformutil.NormalizeUserParams(commonCmdData.GetPlatform())
		if err != nil {
			return build.ConveyorOptions{}, fmt.Errorf("unable to normalize platform params %v: %w", commonCmdData.GetPlatform(), err)
		}
		conveyorOptions.TargetPlatforms = platforms
	}

	conveyorOptions.DeferBuildLog = GetDeferredBuildLog(ctx, commonCmdData)
	if commonCmdData.SkipImageSpecStage != nil {
		conveyorOptions.SkipImageSpecStage = *commonCmdData.SkipImageSpecStage
	}
	return conveyorOptions, nil
}

// GetDeferredBuildLog returns true if build log should be catched and printed on error.
// Default rules are follows:
// - If --require-built-images is specified catch log and print on error.
// - Hide log messages if --log-quiet is specified.
// - Print "live" logs by default or if --log-verbose is specified.
func GetDeferredBuildLog(ctx context.Context, commonCmdData *CmdData) bool {
	requireBuiltImage := GetRequireBuiltImages(commonCmdData)
	isVerbose := logboek.Context(ctx).IsAcceptedLevel(level.Default)
	return requireBuiltImage || !isVerbose
}

func GetConveyorOptionsWithParallel(ctx context.Context, commonCmdData *CmdData, imagesToProcess config.ImagesToProcess, buildStagesOptions build.BuildOptions) (build.ConveyorOptions, error) {
	conveyorOptions, err := GetConveyorOptions(ctx, commonCmdData, imagesToProcess)
	if err != nil {
		return conveyorOptions, err
	}

	conveyorOptions.Parallel = !(buildStagesOptions.ImageBuildOptions.IntrospectAfterError || buildStagesOptions.ImageBuildOptions.IntrospectBeforeError || len(buildStagesOptions.Targets) != 0) && GetParallel(commonCmdData)
	conveyorOptions.ParallelTasksLimit = GetParallelTasksLimit(commonCmdData)
	conveyorOptions.ManifestSigningOptions = buildStagesOptions.ManifestSigningOptions
	conveyorOptions.VerityAnnotationOptions = buildStagesOptions.VerityAnnotationOptions

	return conveyorOptions, nil
}

func GetShouldBeBuiltOptions(commonCmdData *CmdData, imagesToProcess config.ImagesToProcess) (options build.ShouldBeBuiltOptions, err error) {
	customTagFuncList, err := getCustomTagFuncList(getCustomTagOptionValues(commonCmdData), commonCmdData, imagesToProcess)
	if err != nil {
		return options, err
	}

	options = build.ShouldBeBuiltOptions{CustomTagFuncList: customTagFuncList}

	if GetSaveBuildReport(commonCmdData) {
		options.ReportPath, options.ReportFormat, err = GetBuildReportPathAndFormat(commonCmdData)
		if err != nil {
			return options, fmt.Errorf("getting build report path failed: %w", err)
		}
	}

	return options, nil
}

func GetBuildOptions(ctx context.Context, commonCmdData *CmdData, werfConfig *config.WerfConfig, imagesToProcess config.ImagesToProcess) (buildOptions build.BuildOptions, err error) {
	introspectOptions, err := GetIntrospectOptions(commonCmdData, werfConfig)
	if err != nil {
		return buildOptions, err
	}

	customTagFuncList, err := getCustomTagFuncList(getCustomTagOptionValues(commonCmdData), commonCmdData, imagesToProcess)
	if err != nil {
		return buildOptions, err
	}

	signerOptions, err := getSignerOptions(commonCmdData)
	if err != nil {
		return buildOptions, fmt.Errorf("getting signer options: %w", err)
	}

	signer, err := signing.NewSigner(ctx, signerOptions)
	if err != nil {
		return buildOptions, fmt.Errorf("creating signer: %w", err)
	}

	manifestSigningOptions, err := getManifestSigningOptions(commonCmdData, signer)
	if err != nil {
		return buildOptions, fmt.Errorf("getting manifest signing options: %w", err)
	}

	elfSigningOptions, err := getELFSigningOptions(commonCmdData, signer)
	if err != nil {
		return buildOptions, err
	}

	verityAnnotationOptions, err := getVerityAnnotationOptions(commonCmdData)
	if err != nil {
		return build.BuildOptions{}, fmt.Errorf("getting verity annotation options: %w", err)
	}

	buildOptions = build.BuildOptions{
		SkipAddManagedImagesRecords:  werfConfig.Meta.Cleanup.DisableCleanup,
		SkipImageMetadataPublication: *commonCmdData.Dev || werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy || werfConfig.Meta.Cleanup.DisableCleanup,
		CustomTagFuncList:            customTagFuncList,
		ImageBuildOptions: container_backend.BuildOptions{
			IntrospectAfterError:  GetIntrospectAfterError(commonCmdData),
			IntrospectBeforeError: GetIntrospectBeforeError(commonCmdData),
		},
		IntrospectOptions:       introspectOptions,
		ManifestSigningOptions:  manifestSigningOptions,
		ELFSigningOptions:       elfSigningOptions,
		VerityAnnotationOptions: verityAnnotationOptions,
	}

	if GetSaveBuildReport(commonCmdData) {
		buildOptions.ReportPath, buildOptions.ReportFormat, err = GetBuildReportPathAndFormat(commonCmdData)
		if err != nil {
			return buildOptions, fmt.Errorf("getting build report path failed: %w", err)
		}
	}

	return buildOptions, nil
}

func getVerityAnnotationOptions(commonCmdData *CmdData) (verify_annotation.Options, error) {
	return verify_annotation.Options{
		Enabled: lo.FromPtr(commonCmdData.AnnotateLayersWithDmvVerityRootHash),
	}, nil
}

func getSignerOptions(commonCmdData *CmdData) (signing.SignerOptions, error) {
	if !GetSignManifest(commonCmdData) && !GetSignELFFiles(commonCmdData) {
		return signing.SignerOptions{}, nil
	}
	if commonCmdData.SignKey == nil || *commonCmdData.SignKey == "" {
		return signing.SignerOptions{}, fmt.Errorf("signing key is required (the private signing key must be specified with --sign-key option)")
	}
	if commonCmdData.SignCert == nil || *commonCmdData.SignCert == "" {
		return signing.SignerOptions{}, fmt.Errorf("signing certificate is required (the public signing certificate must be specified with --sign-cert option)")
	}
	return signing.SignerOptions{
		KeyRef:           lo.FromPtr(commonCmdData.SignKey),
		CertRef:          lo.FromPtr(commonCmdData.SignCert),
		IntermediatesRef: lo.FromPtr(commonCmdData.SignIntermediates),
	}, nil
}

func getManifestSigningOptions(commonCmdData *CmdData, signer *signing.Signer) (signing.ManifestSigningOptions, error) {
	options := signing.NewManifestSigningOptions(signer)
	options.Enabled = GetSignManifest(commonCmdData)
	return options, nil
}

func getELFSigningOptions(commonCmdData *CmdData, signer *signing.Signer) (signing.ELFSigningOptions, error) {
	options := signing.NewELFSigningOptions(signer)

	if !GetSignELFFiles(commonCmdData) && !GetBSignELFFiles(commonCmdData) {
		return options, nil
	}

	if GetSignELFFiles(commonCmdData) {
		options.InHouseEnabled = true
	}

	// bsign
	{
		if !GetBSignELFFiles(commonCmdData) {
			return options, nil
		} else {
			options.BsignEnabled = true
		}

		if *commonCmdData.ELFPGPPrivateKeyPassphrase != "" {
			options.PGPPrivateKeyPassphrase = *commonCmdData.ELFPGPPrivateKeyPassphrase
		}

		if *commonCmdData.ELFPGPPrivateKeyBase64 != "" && *commonCmdData.ELFPGPPrivateKeyFingerprint != "" {
			return options, fmt.Errorf("both --elf-pgp-private-key-base64 and --elf-pgp-private-key-fingerprint params are specified, only one of them should be specified")
		} else if *commonCmdData.ELFPGPPrivateKeyBase64 == "" && *commonCmdData.ELFPGPPrivateKeyFingerprint == "" {
			return options, fmt.Errorf("either --elf-pgp-private-key-base64 or --elf-pgp-private-key-fingerprint param is required")
		}

		if *commonCmdData.ELFPGPPrivateKeyFingerprint != "" {
			options.PGPPrivateKeyFingerprint = *commonCmdData.ELFPGPPrivateKeyFingerprint
			return options, nil
		}

		// Get fingerprint and import key.
		{
			keyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(*commonCmdData.ELFPGPPrivateKeyBase64))
			if err != nil {
				return options, fmt.Errorf("unable to decode PGP key from base64: %w", err)
			}

			pgpKeyString := string(keyBytes)
			entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(keyBytes))
			if err != nil {
				return options, fmt.Errorf("unable to read PGP key: %w", err)
			}

			firstKey := entityList[0].PrimaryKey
			fingerprint := firstKey.Fingerprint
			options.PGPPrivateKeyFingerprint = fmt.Sprintf("%X", fingerprint)

			// Import PGP key.
			{
				ctx := context.Background()
				cmd := exec.CommandContextCancellation(ctx, "gpg", "--import")
				cmd.Stdin = bytes.NewBufferString(pgpKeyString)

				if options.PGPPrivateKeyPassphrase != "" {
					cmd.Args = append(cmd.Args, "--batch")
					cmd.Args = append(cmd.Args, "--passphrase=$WERF_SERVICE_ELF_PGP_PRIVATE_KEY_PASSPHRASE")
					cmd.Env = append(cmd.Env, fmt.Sprintf("WERF_SERVICE_ELF_PGP_PRIVATE_KEY_PASSPHRASE=%s", options.PGPPrivateKeyPassphrase))
				}

				err := cmd.Run()
				if err != nil {
					return options, fmt.Errorf("unable to import PGP key: %w", err)
				}
			}
		}
	}

	return options, nil
}

func getCustomTagFuncList(tagOptionValues []string, commonCmdData *CmdData, imagesToProcess config.ImagesToProcess) ([]image.CustomTagFunc, error) {
	if len(tagOptionValues) == 0 {
		return nil, nil
	}

	if *commonCmdData.Repo.Address == "" || *commonCmdData.Repo.Address == storage.LocalStorageAddress {
		return nil, fmt.Errorf("custom tags can only be used with remote storage: --repo=ADDRESS param required")
	}

	templateName := "--add/use-custom-tag"
	tmpl := template.New(templateName).Delims("%", "%")
	tmpl = tmpl.Funcs(map[string]interface{}{
		"image":                   func() string { return "%[1]s" },
		"image_slug":              func() string { return "%[2]s" },
		"image_safe_slug":         func() string { return "%[3]s" },
		"image_content_based_tag": func() string { return "%[4]s" },
	})

	var tagFuncList []image.CustomTagFunc
	for _, optionValue := range tagOptionValues {
		tmpl, err := tmpl.Parse(optionValue)
		if err != nil {
			return nil, fmt.Errorf("invalid custom tag %q: %w", optionValue, err)
		}

		buf := bytes.NewBuffer(nil)
		if err := tmpl.ExecuteTemplate(buf, templateName, nil); err != nil {
			return nil, fmt.Errorf("invalid custom tag %q: %w", optionValue, err)
		}

		tagOrFormat := buf.String()
		tagFunc := func(imageName, contentBasedTag string) string {
			if strings.ContainsRune(tagOrFormat, '%') {
				return fmt.Sprintf(tagOrFormat, imageName, slug.Slug(imageName), slug.DockerTag(imageName), contentBasedTag)
			} else {
				return tagOrFormat
			}
		}

		contentBasedTagStub := strings.Repeat("x", 70) // 1b77754d35b0a3e603731828ee6f2400c4f937382874db2566c616bb-1624991915332
		var prevImageTag string
		for _, imageName := range imagesToProcess.FinalImageNameList {
			imageTag := tagFunc(imageName, contentBasedTagStub)

			if err := slug.ValidateDockerTag(imageTag); err != nil {
				return nil, fmt.Errorf("invalid custom tag %q: %w", optionValue, err)
			}

			if prevImageTag == "" {
				prevImageTag = imageTag
				continue
			} else if imageTag == prevImageTag {
				return nil, fmt.Errorf("invalid custom tag %q: it is necessary to use the image name in the tag format if there is more than one image in the werf config (e.g., %q)", tagOrFormat, fmt.Sprintf("%s-%s", "%image%", tagOrFormat))
			}
		}

		tagFuncList = append(tagFuncList, tagFunc)
	}

	return tagFuncList, nil
}

func GetUseCustomTagFunc(commonCmdData *CmdData, giterminismManager giterminism_manager.Interface, imagesToProcess config.ImagesToProcess) (image.CustomTagFunc, error) {
	var tagOptionValues []string
	if *commonCmdData.UseCustomTag != "" {
		tagOptionValues = []string{*commonCmdData.UseCustomTag}
	}

	customTagFuncList, err := getCustomTagFuncList(tagOptionValues, commonCmdData, imagesToProcess)
	if err != nil {
		return nil, err
	}

	if len(customTagFuncList) == 0 {
		return nil, nil
	}

	if err := giterminismManager.Inspector().InspectCustomTags(); err != nil {
		return nil, err
	}

	if len(customTagFuncList) != 1 {
		panic("unexpected condition")
	}

	return customTagFuncList[0], nil
}

func getCustomTagOptionValues(commonCmdData *CmdData) []string {
	var tagOptionValues []string

	if commonCmdData.UseCustomTag != nil && *commonCmdData.UseCustomTag != "" {
		tagOptionValues = append(tagOptionValues, *commonCmdData.UseCustomTag)
	}

	if commonCmdData.AddCustomTag != nil {
		tagOptionValues = append(tagOptionValues, getAddCustomTag(commonCmdData)...)
	}

	return tagOptionValues
}
