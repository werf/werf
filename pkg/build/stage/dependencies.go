package stage

import (
	"context"
	"fmt"
	"strings"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
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

func (s *DependenciesStage) GetDependencies(ctx context.Context, c Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	var args []string

	for _, elm := range s.imports {
		sourceContentDigest := getSourceImageContentDigest(c, s.targetPlatform, elm)

		args = append(args, sourceContentDigest)
		args = append(args, elm.Add)
		args = append(args, elm.To)
		args = append(args, elm.Group, elm.Owner)
		args = append(args, strings.Join(elm.IncludePaths, "///"))
		args = append(args, strings.Join(elm.ExcludePaths, "///"))
	}

	for _, dep := range s.dependencies {
		args = append(args, "Dependency", c.GetImageNameForLastImageStage(s.targetPlatform, dep.ImageName))
		for _, imp := range dep.Imports {
			args = append(args, "DependencyImport", getDependencyImportID(imp))
		}
	}

	return util.Sha256Hash(args...), nil
}

func (s *DependenciesStage) GetContextDependencies(_ context.Context, c Conveyor) (string, error) {
	var args []string

	for _, elm := range s.imports {
		sourceContextDigest := getSourceImageContextDigest(c, s.targetPlatform, elm)

		args = append(args, sourceContextDigest)
		args = append(args, elm.Add)
		args = append(args, elm.To)
		args = append(args, elm.Group, elm.Owner)
		args = append(args, strings.Join(elm.IncludePaths, "///"))
		args = append(args, strings.Join(elm.ExcludePaths, "///"))
	}

	for _, dep := range s.dependencies {
		args = append(args, "Dependency", c.GetImageContextDigest(s.targetPlatform, dep.ImageName))
		for _, imp := range dep.Imports {
			args = append(args, "DependencyImport", getDependencyImportID(imp))
		}
	}

	return util.Sha256Hash(args...), nil
}

func getSourceImageContextDigest(c Conveyor, targetPlatform string, importElm *config.Import) string {
	if importElm.ExternalImage {
		return fmt.Sprintf("%s:%s", image.WerfImportSourceExternalImagePrefix, importElm.From)
	}

	sourceImageName := getSourceImageName(importElm)

	if importElm.Stage == "" {
		return c.GetImageContextDigest(targetPlatform, sourceImageName)
	}

	return c.GetImageStageContentDigest(targetPlatform, sourceImageName, importElm.Stage)
}

func (s *DependenciesStage) prepareImageWithLegacyStapelBuilder(ctx context.Context, c Conveyor, cr container_backend.ContainerBackend, _, stageImage *StageImage) error {
	imageServiceCommitChangeOptions := stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions()

	for _, elm := range s.imports {
		sourceImageName := getSourceImageName(elm)
		srv, err := c.GetImportServer(ctx, s.targetPlatform, sourceImageName, elm.Stage, elm.ExternalImage)
		if err != nil {
			return fmt.Errorf("unable to get import server for image %q: %w", sourceImageName, err)
		}

		command := srv.GetCopyCommand(ctx, elm)
		stageImage.Builder.LegacyStapelStageBuilder().Container().AddServiceRunCommands(command)

		sourceStageIDLabelKey := image.WerfImportSourceStageIDLabelPrefix + getImportID(elm)
		sourceStageID := getSourceStageID(c, s.targetPlatform, elm)

		imageServiceCommitChangeOptions.AddLabel(map[string]string{
			sourceStageIDLabelKey: sourceStageID,
		})
	}

	for _, dep := range s.dependencies {
		depImageServiceOptions := stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions()

		depImageName := c.GetImageNameForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageID := c.GetImageIDForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageDigest := c.GetImageDigestForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageRepo, depImageTag := image.ParseRepositoryAndTag(depImageName)

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

		var sourceImageName string

		if elm.ExternalImage {
			sourceImageName = elm.From
		} else {
			sourceImageConfigName := getSourceImageName(elm)
			if elm.Stage == "" {
				sourceImageName = c.GetImageNameForLastImageStage(s.targetPlatform, sourceImageConfigName)
			} else {
				sourceImageName = c.GetImageNameForImageStage(s.targetPlatform, sourceImageConfigName, elm.Stage)
			}
		}

		sourceStageIDLabelKey := image.WerfImportSourceStageIDLabelPrefix + getImportID(elm)
		sourceStageID := getSourceStageID(c, s.targetPlatform, elm)

		stageImage.Builder.StapelStageBuilder().AddLabels(map[string]string{
			sourceStageIDLabelKey: sourceStageID,
		})
		stageImage.Builder.StapelStageBuilder().AddDependencyImport(sourceImageName, elm.Add, elm.To, elm.IncludePaths, elm.ExcludePaths, elm.Owner, elm.Group)
	}

	for _, dep := range s.dependencies {
		depImageName := c.GetImageNameForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageID := c.GetImageIDForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageDigest := c.GetImageDigestForLastImageStage(s.targetPlatform, dep.ImageName)
		depImageRepo, depImageTag := image.ParseRepositoryAndTag(depImageName)

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

func getDependencyImportID(dependencyImport *config.DependencyImport) string {
	return util.Sha256Hash(
		"Type", string(dependencyImport.Type),
		"TargetEnv", dependencyImport.TargetEnv,
	)
}

func getImportID(importElm *config.Import) string {
	return util.Sha256Hash(
		"ImageName", importElm.From,
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

func getSourceStageID(c Conveyor, targetPlatform string, importElm *config.Import) string {
	if importElm.ExternalImage {
		return fmt.Sprintf("%s:%s", image.WerfImportSourceExternalImagePrefix, importElm.From)
	}

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
	if importElm.ExternalImage {
		return fmt.Sprintf("%s:%s", image.WerfImportSourceExternalImagePrefix, importElm.From)
	}

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
	return importElm.From
}
