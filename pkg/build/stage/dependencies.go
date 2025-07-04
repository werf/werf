package stage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker"
	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/stapel"
	"github.com/werf/werf/v2/pkg/storage"
)

type getImportsOptions struct {
	Before StageName
	After  StageName
}

func getImports(imageBaseConfig *config.StapelImageBase, options *getImportsOptions) []*config.Import {
	var imports []*config.Import
	for _, elm := range imageBaseConfig.Import {
		if options.Before != "" && elm.Before != "" && elm.Before == string(options.Before) {
			imports = append(imports, elm)
		} else if options.After != "" && elm.After != "" && elm.After == string(options.After) {
			imports = append(imports, elm)
		}
	}

	return imports
}

func getDependencies(imageBaseConfig *config.StapelImageBase, options *getImportsOptions) []*config.Dependency {
	var dependencies []*config.Dependency

	for _, dep := range imageBaseConfig.Dependencies {
		switch {
		case dep.Before != "" && string(options.Before) == dep.Before:
			dependencies = append(dependencies, dep)
		case dep.After != "" && string(options.After) == dep.After:
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies
}

func newDependenciesStage(imports []*config.Import, dependencies []*config.Dependency, name StageName, baseStageOptions *BaseStageOptions) *DependenciesStage {
	s := &DependenciesStage{}
	s.imports = imports
	s.dependencies = dependencies
	s.BaseStage = NewBaseStage(name, baseStageOptions)
	return s
}

type DependenciesStage struct {
	*BaseStage

	imports      []*config.Import
	dependencies []*config.Dependency
}

func (s *DependenciesStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	var args []string

	if len(s.imports) != 0 {
		if err := logboek.Context(ctx).Default().LogProcess("Calculating import checksums").DoError(func() error {
			for ind, elm := range s.imports {
				sourceChecksum, err := s.getImportSourceChecksum(ctx, c, cb, elm)
				if err != nil {
					return fmt.Errorf("unable to get import %d source checksum: %w", ind, err)
				}

				var importTitle string
				{
					importTitle = fmt.Sprintf("image=%s add=%s to=%s", elm.ImageName, elm.Add, elm.To)
					if len(elm.IncludePaths) != 0 {
						importTitle += fmt.Sprintf(" includePaths=%v", elm.IncludePaths)
					}
					if len(elm.ExcludePaths) != 0 {
						importTitle += fmt.Sprintf(" excludePaths=%v", elm.ExcludePaths)
					}
					importTitle = fmt.Sprintf("import[%s]", importTitle)
				}

				logboek.Context(ctx).Default().LogF("%s: %s\n", sourceChecksum, importTitle)

				args = append(args, sourceChecksum)
				args = append(args, elm.To)
				args = append(args, elm.Group, elm.Owner)
			}
			return nil
		}); err != nil {
			return "", err
		}
	}

	for _, dep := range s.dependencies {
		args = append(args, "Dependency", c.GetImageNameForLastImageStage(s.targetPlatform, dep.ImageName))
		for _, imp := range dep.Imports {
			args = append(args, "DependencyImport", getDependencyImportID(imp))
		}
	}

	return util.Sha256Hash(args...), nil
}

func (s *DependenciesStage) prepareImageWithLegacyStapelBuilder(ctx context.Context, c Conveyor, cr container_backend.ContainerBackend, _, stageImage *StageImage) error {
	imageServiceCommitChangeOptions := stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions()
	for _, elm := range s.imports {
		sourceImageName := getSourceImageName(elm)
		srv, err := c.GetImportServer(ctx, s.targetPlatform, sourceImageName, elm.Stage)
		if err != nil {
			return fmt.Errorf("unable to get import server for image %q: %w", sourceImageName, err)
		}

		command := srv.GetCopyCommand(ctx, elm)
		stageImage.Builder.LegacyStapelStageBuilder().Container().AddServiceRunCommands(command)

		checksumLabelKey := imagePkg.WerfImportChecksumLabelPrefix + getImportID(elm)
		sourceStageIDLabelKey := imagePkg.WerfImportSourceStageIDLabelPrefix + getImportID(elm)

		importSourceID := getImportSourceID(c, s.targetPlatform, elm)
		importMetadata, err := c.GetImportMetadata(ctx, s.projectName, importSourceID)
		if err != nil {
			return fmt.Errorf("unable to get import source checksum: %w", err)
		} else if importMetadata == nil {
			panic(fmt.Sprintf("import metadata %s not found", importSourceID))
		}

		imageServiceCommitChangeOptions.AddLabel(map[string]string{
			checksumLabelKey:      importMetadata.Checksum,
			sourceStageIDLabelKey: importMetadata.SourceStageID,
		})
	}

	for _, dep := range s.dependencies {
		depImageServiceOptions := stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions()

		depImageName := c.GetImageNameForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageID := c.GetImageIDForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageDigest := c.GetImageDigestForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageRepo, depImageTag := imagePkg.ParseRepositoryAndTag(depImageName)

		for _, img := range dep.Imports {
			switch img.Type {
			case config.ImageRepoImport:
				depImageServiceOptions.AddEnv(map[string]string{
					img.TargetEnv: depImageRepo,
				})
			case config.ImageTagImport:
				depImageServiceOptions.AddEnv(map[string]string{
					img.TargetEnv: depImageTag,
				})
			case config.ImageNameImport:
				depImageServiceOptions.AddEnv(map[string]string{
					img.TargetEnv: depImageName,
				})
			case config.ImageIDImport:
				depImageServiceOptions.AddEnv(map[string]string{
					img.TargetEnv: depImageID,
				})
			case config.ImageDigestImport:
				depImageServiceOptions.AddEnv(map[string]string{
					img.TargetEnv: depImageDigest,
				})
			}
		}

		depStageID := c.GetStageIDForLastImageStage(s.targetPlatform, dep.ImageName)
		imageServiceCommitChangeOptions.AddLabel(map[string]string{
			dependencyLabelKey(depStageID): depStageID,
		})
	}

	return nil
}

func (s *DependenciesStage) prepareImage(ctx context.Context, c Conveyor, cr container_backend.ContainerBackend, _, stageImage *StageImage) error {
	for _, elm := range s.imports {
		sourceImageConfigName := getSourceImageName(elm)
		var sourceImageName string
		if elm.Stage == "" {
			sourceImageName = c.GetImageNameForLastImageStage(s.targetPlatform, sourceImageConfigName)
		} else {
			sourceImageName = c.GetImageNameForImageStage(s.targetPlatform, sourceImageConfigName, elm.Stage)
		}

		checksumLabelKey := imagePkg.WerfImportChecksumLabelPrefix + getImportID(elm)
		sourceStageIDLabelKey := imagePkg.WerfImportSourceStageIDLabelPrefix + getImportID(elm)

		importSourceID := getImportSourceID(c, s.targetPlatform, elm)
		importMetadata, err := c.GetImportMetadata(ctx, s.projectName, importSourceID)
		if err != nil {
			return fmt.Errorf("unable to get import source checksum: %w", err)
		} else if importMetadata == nil {
			panic(fmt.Sprintf("import metadata %s not found", importSourceID))
		}

		stageImage.Builder.StapelStageBuilder().AddLabels(map[string]string{
			checksumLabelKey:      importMetadata.Checksum,
			sourceStageIDLabelKey: importMetadata.SourceStageID,
		})
		stageImage.Builder.StapelStageBuilder().AddDependencyImport(sourceImageName, elm.Add, elm.To, elm.IncludePaths, elm.ExcludePaths, elm.Owner, elm.Group)
	}

	for _, dep := range s.dependencies {
		depImageName := c.GetImageNameForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageID := c.GetImageIDForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageDigest := c.GetImageDigestForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageRepo, depImageTag := imagePkg.ParseRepositoryAndTag(depImageName)

		for _, img := range dep.Imports {
			switch img.Type {
			case config.ImageRepoImport:
				stageImage.Builder.StapelStageBuilder().AddEnvs(map[string]string{
					img.TargetEnv: depImageRepo,
				})
			case config.ImageTagImport:
				stageImage.Builder.StapelStageBuilder().AddEnvs(map[string]string{
					img.TargetEnv: depImageTag,
				})
			case config.ImageNameImport:
				stageImage.Builder.StapelStageBuilder().AddEnvs(map[string]string{
					img.TargetEnv: depImageName,
				})
			case config.ImageIDImport:
				stageImage.Builder.StapelStageBuilder().AddEnvs(map[string]string{
					img.TargetEnv: depImageID,
				})
			case config.ImageDigestImport:
				stageImage.Builder.StapelStageBuilder().AddEnvs(map[string]string{
					img.TargetEnv: depImageDigest,
				})
			}
		}
		depStageID := c.GetStageIDForLastImageStage(s.targetPlatform, dep.ImageName)
		stageImage.Builder.StapelStageBuilder().AddLabels(map[string]string{
			dependencyLabelKey(depStageID): depStageID,
		})
	}

	return nil
}

func (s *DependenciesStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	if c.UseLegacyStapelBuilder(cb) {
		return s.prepareImageWithLegacyStapelBuilder(ctx, c, cb, prevBuiltImage, stageImage)
	} else {
		return s.prepareImage(ctx, c, cb, prevBuiltImage, stageImage)
	}
}

func (s *DependenciesStage) getImportSourceChecksum(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, importElm *config.Import) (string, error) {
	importSourceID := getImportSourceID(c, s.targetPlatform, importElm)
	importMetadata, err := c.GetImportMetadata(ctx, s.projectName, importSourceID)
	if err != nil {
		return "", fmt.Errorf("unable to get import metadata: %w", err)
	}

	if importMetadata == nil {
		checksum, err := s.generateImportChecksum(ctx, c, cb, importElm)
		if err != nil {
			return "", fmt.Errorf("unable to generate import source checksum: %w", err)
		}

		sourceStageID := getSourceStageID(c, s.targetPlatform, importElm)
		importMetadata = &storage.ImportMetadata{
			ImportSourceID: importSourceID,
			SourceStageID:  sourceStageID,
			Checksum:       checksum,
		}

		if err := c.PutImportMetadata(ctx, s.projectName, importMetadata); err != nil {
			return "", fmt.Errorf("unable to put import metadata: %w", err)
		}
	}

	return importMetadata.Checksum, nil
}

func (s *DependenciesStage) generateImportChecksum(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, importElm *config.Import) (string, error) {
	if err := fetchSourceImageDockerImage(ctx, c, s.targetPlatform, importElm); err != nil {
		return "", fmt.Errorf("unable to fetch source image: %w", err)
	}

	sourceImageDockerImageName := getSourceImageDockerImageName(c, s.targetPlatform, importElm)

	if c.UseLegacyStapelBuilder(cb) {
		importSourceID := getImportSourceID(c, s.targetPlatform, importElm)

		stapelContainerName, err := stapel.GetOrCreateContainer(ctx)
		if err != nil {
			return "", err
		}

		importHostTmpDir := filepath.Join(s.imageTmpDir, string(s.Name()), "imports", importSourceID)
		importContainerDir := s.containerWerfDir

		importScriptHostTmpPath := filepath.Join(importHostTmpDir, "script.sh")
		resultChecksumHostTmpPath := filepath.Join(importHostTmpDir, "checksum")
		importScriptContainerPath := path.Join(importContainerDir, "script.sh")
		resultChecksumContainerPath := path.Join(importContainerDir, "checksum")

		checksumScript := generateChecksumScript(importElm.Add, importElm.IncludePaths, importElm.ExcludePaths, resultChecksumContainerPath)
		if err := stapel.CreateScript(importScriptHostTmpPath, checksumScript); err != nil {
			return "", fmt.Errorf("unable to create script: %w", err)
		}

		runArgs := []string{
			"--rm",
			"--user=0:0",
			"--workdir=/",
			fmt.Sprintf("--volumes-from=%s", stapelContainerName),
			fmt.Sprintf("--volume=%s:%s", importHostTmpDir, importContainerDir),
			fmt.Sprintf("--entrypoint=%s", stapel.BashBinPath()),
			sourceImageDockerImageName,
			importScriptContainerPath,
		}

		if debugImportSourceChecksum() {
			fmt.Println(runArgs)
		}

		if output, err := docker.CliRun_RecordedOutput(ctx, runArgs...); err != nil {
			logboek.Context(ctx).Error().LogF("%s", output)
			return "", err
		}

		data, err := ioutil.ReadFile(resultChecksumHostTmpPath)
		if err != nil {
			return "", fmt.Errorf("unable to read file with import source checksum: %w", err)
		}

		checksum := strings.TrimSpace(string(data))

		return checksum, nil
	} else {
		var checksum string
		var err error

		logboek.Context(ctx).Debug().LogProcess("Calculating dependency import checksum").Do(func() {
			checksum, err = cb.CalculateDependencyImportChecksum(ctx, container_backend.DependencyImportSpec{
				ImageName:    sourceImageDockerImageName,
				FromPath:     importElm.Add,
				ToPath:       importElm.To,
				IncludePaths: importElm.IncludePaths,
				ExcludePaths: importElm.ExcludePaths,
				Owner:        importElm.Owner,
				Group:        importElm.Group,
			}, container_backend.CalculateDependencyImportChecksum{TargetPlatform: s.targetPlatform})
		})

		if err != nil {
			return "", fmt.Errorf("unable to calculate dependency import checksum in %s: %w", sourceImageDockerImageName, err)
		}
		return checksum, nil
	}
}

func generateChecksumScript(from string, includePaths, excludePaths []string, resultChecksumPath string) []string {
	findCommandParts := append([]string{}, stapel.FindBinPath(), "-H", from, "-type", "f")

	var nameIncludeArgs []string
	for _, includePath := range includePaths {
		formattedPath := util.SafeTrimGlobsAndSlashesFromPath(includePath)
		nameIncludeArgs = append(
			nameIncludeArgs,
			fmt.Sprintf("-wholename \"%s\"", path.Join(from, formattedPath)),
			fmt.Sprintf("-wholename \"%s\"", path.Join(from, formattedPath, "**")),
		)
	}

	if len(nameIncludeArgs) != 0 {
		findCommandParts = append(findCommandParts, fmt.Sprintf("\\( %s \\)", strings.Join(nameIncludeArgs, " -or ")))
	}

	excludePaths = append(excludePaths, stapel.CONTAINER_MOUNT_ROOT)

	var nameExcludeArgs []string
	for _, excludePath := range excludePaths {
		formattedPath := util.SafeTrimGlobsAndSlashesFromPath(excludePath)
		nameExcludeArgs = append(
			nameExcludeArgs,
			fmt.Sprintf("! -wholename \"%s\"", path.Join(from, formattedPath)),
			fmt.Sprintf("! -wholename \"%s\"", path.Join(from, formattedPath, "**")),
		)
	}

	if len(nameExcludeArgs) != 0 {
		if len(nameIncludeArgs) != 0 {
			findCommandParts = append(findCommandParts, fmt.Sprintf("-and"))
		}

		findCommandParts = append(findCommandParts, fmt.Sprintf("\\( %s \\)", strings.Join(nameExcludeArgs, " -and ")))
	}

	findCommand := strings.Join(findCommandParts, " ")

	sortCommandParts := append([]string{}, stapel.SortBinPath(), "-n")
	sortCommand := strings.Join(sortCommandParts, " ")

	md5SumCommand := stapel.Md5sumBinPath()

	cutCommandParts := append([]string{}, stapel.CutBinPath(), "-c", "1-32")
	cutCommand := strings.Join(cutCommandParts, " ")

	commands := append([]string{}, findCommand, sortCommand, "checksum", md5SumCommand, cutCommand)

	script := generateChecksumBashFunction()
	script = append(script, fmt.Sprintf("%s > %s", strings.Join(commands, " | "), resultChecksumPath))

	return script
}

func generateChecksumBashFunction() []string {
	var calculateChecksum string

	// TODO: remove in v3 (WERF_EXPERIMENTAL_STAPEL_IMPORT_PERMISSIONS=1 as default)
	switch util.GetBoolEnvironmentDefaultFalse("WERF_EXPERIMENTAL_STAPEL_IMPORT_PERMISSIONS") {
	case true:
		calculateChecksum = fmt.Sprintf(`printf '%%s\t%%s\t%%s\n' "$(%[1]s "${line}" | %[2]s -c 1-32)" "$(%[3]s --format=%%A "${line}")" "${line}"`,
			stapel.Md5sumBinPath(), stapel.CutBinPath(), stapel.StatBinPath())
	default:
		calculateChecksum = fmt.Sprintf(`%[1]s "${line}"`,
			stapel.Md5sumBinPath())
	}

	return []string{
		`checksum() {`,
		`  while read -r line; do`,
		`    ` + calculateChecksum,
		`  done`,
		`}`,
		``,
	}
}

func getDependencyImportID(dependencyImport *config.DependencyImport) string {
	return util.Sha256Hash(
		"Type", string(dependencyImport.Type),
		"TargetEnv", dependencyImport.TargetEnv,
	)
}

func getImportID(importElm *config.Import) string {
	return util.Sha256Hash(
		"ImageName", importElm.ImageName,
		"ArtifactName", importElm.ArtifactName,
		"Stage", importElm.Stage,
		"After", importElm.After,
		"Before", importElm.Before,
		"Add", importElm.Add,
		"To", importElm.To,
		"Group", importElm.Group,
		"Owner", importElm.Owner,
		"IncludePaths", strings.Join(importElm.IncludePaths, "///"),
		"ExcludePaths", strings.Join(importElm.ExcludePaths, "///"),
	)
}

func getImportSourceID(c Conveyor, targetPlatform string, importElm *config.Import) string {
	args := []string{
		"SourceImageContentDigest", getSourceImageContentDigest(c, targetPlatform, importElm),
		"Add", importElm.Add,
		"IncludePaths", strings.Join(importElm.IncludePaths, "///"),
		"ExcludePaths", strings.Join(importElm.ExcludePaths, "///"),
	}

	// TODO: remove in v3 (WERF_EXPERIMENTAL_STAPEL_IMPORT_PERMISSIONS=1 as default)
	if util.GetBoolEnvironmentDefaultFalse("WERF_EXPERIMENTAL_STAPEL_IMPORT_PERMISSIONS") {
		args = append(args,
			"CacheVersion", "true",
		)
	}

	return util.Sha256Hash(args...)
}

func fetchSourceImageDockerImage(ctx context.Context, c Conveyor, targetPlatform string, importElm *config.Import) error {
	sourceImageName := getSourceImageName(importElm)
	if importElm.Stage == "" {
		return c.FetchLastNonEmptyImageStage(ctx, targetPlatform, sourceImageName)
	} else {
		return c.FetchImageStage(ctx, targetPlatform, sourceImageName, importElm.Stage)
	}
}

func getSourceImageDockerImageName(c Conveyor, targetPlatform string, importElm *config.Import) string {
	sourceImageName := getSourceImageName(importElm)

	var sourceImageDockerImageName string
	if importElm.Stage == "" {
		sourceImageDockerImageName = c.GetImageNameForLastImageStage(targetPlatform, sourceImageName)
	} else {
		sourceImageDockerImageName = c.GetImageNameForImageStage(targetPlatform, sourceImageName, importElm.Stage)
	}

	return sourceImageDockerImageName
}

func getSourceStageID(c Conveyor, targetPlatform string, importElm *config.Import) string {
	sourceImageName := getSourceImageName(importElm)

	var sourceStageID string
	if importElm.Stage == "" {
		sourceStageID = c.GetStageIDForLastImageStage(targetPlatform, sourceImageName)
	} else {
		sourceStageID = c.GetStageIDForImageStage(targetPlatform, sourceImageName, importElm.Stage)
	}

	return sourceStageID
}

func getSourceImageContentDigest(c Conveyor, targetPlatform string, importElm *config.Import) string {
	sourceImageName := getSourceImageName(importElm)

	var sourceImageContentDigest string
	if importElm.Stage == "" {
		sourceImageContentDigest = c.GetImageContentDigest(targetPlatform, sourceImageName)
	} else {
		sourceImageContentDigest = c.GetImageStageContentDigest(targetPlatform, sourceImageName, importElm.Stage)
	}

	return sourceImageContentDigest
}

func getSourceImageName(importElm *config.Import) string {
	var sourceImageName string
	if importElm.ImageName != "" {
		sourceImageName = importElm.ImageName
	} else {
		sourceImageName = importElm.ArtifactName
	}

	return sourceImageName
}

func debugImportSourceChecksum() bool {
	return os.Getenv("WERF_DEBUG_IMPORT_SOURCE_CHECKSUM") == "1"
}
