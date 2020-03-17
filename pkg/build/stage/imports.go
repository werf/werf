package stage

import (
	"fmt"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/container_runtime"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/slug"
	"github.com/flant/werf/pkg/util"
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

func newImportsStage(imports []*config.Import, name StageName, baseStageOptions *NewBaseStageOptions) *ImportsStage {
	s := &ImportsStage{}
	s.imports = imports
	s.BaseStage = newBaseStage(name, baseStageOptions)
	return s
}

type ImportsStage struct {
	*BaseStage

	imports []*config.Import
}

func (s *ImportsStage) GetDependencies(c Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	var args []string

	for _, elm := range s.imports {
		if elm.ImageName != "" {
			args = append(args, c.GetImageStagesSignature(elm.ImageName))
		} else {
			args = append(args, c.GetImageStagesSignature(elm.ArtifactName))
		}

		args = append(args, elm.Add, elm.To)
		args = append(args, elm.Group, elm.Owner)
		args = append(args, elm.IncludePaths...)
		args = append(args, elm.ExcludePaths...)
	}

	return util.Sha256Hash(args...), nil
}

func (s *ImportsStage) PrepareImage(c Conveyor, _, image container_runtime.ImageInterface) error {
	for _, elm := range s.imports {
		var importImage string
		if elm.ImageName != "" {
			importImage = elm.ImageName
		} else {
			importImage = elm.ArtifactName
		}

		srv, err := c.GetImportServer(importImage)
		if err != nil {
			return fmt.Errorf("unable to get import server for image %q: %s", importImage, err)
		}

		command := srv.GetCopyCommand(elm)
		image.Container().AddServiceRunCommands(command)

		imageServiceCommitChangeOptions := image.Container().ServiceCommitChangeOptions()

		var labelKey, labelValue string
		if elm.ImageName != "" {
			labelKey = imagePkg.WerfImportLabelPrefix + slug.Slug(elm.ImageName)
			labelValue = c.GetImageLastStageImageID(elm.ImageName)
		} else {
			labelKey = imagePkg.WerfImportLabelPrefix + slug.Slug(elm.ArtifactName)
			labelValue = c.GetImageLastStageImageID(elm.ArtifactName)
		}

		imageServiceCommitChangeOptions.AddLabel(map[string]string{labelKey: labelValue})
	}

	return nil
}
