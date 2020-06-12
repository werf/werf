package stage

import (
	"fmt"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/util"
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
		var imgName string
		if elm.ImageName != "" {
			imgName = elm.ImageName
		} else {
			imgName = elm.ArtifactName
		}

		if elm.Stage == "" {
			args = append(args, c.GetImageContentSignature(imgName))
		} else {
			args = append(args, c.GetImageStageContentSignature(imgName, elm.Stage))
		}

		args = append(args, elm.Add, elm.To)
		args = append(args, elm.Group, elm.Owner)
		args = append(args, elm.IncludePaths...)
		args = append(args, elm.ExcludePaths...)

		if elm.Stage != "" {
			args = append(args, elm.Stage)
		}
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

		srv, err := c.GetImportServer(importImage, elm.Stage)
		if err != nil {
			return fmt.Errorf("unable to get import server for image %q: %s", importImage, err)
		}

		command := srv.GetCopyCommand(elm)
		image.Container().AddServiceRunCommands(command)

		imageServiceCommitChangeOptions := image.Container().ServiceCommitChangeOptions()

		labelKey := imagePkg.WerfImportLabelPrefix + slug.Slug(importImage)

		var labelValue string
		if elm.Stage == "" {
			labelValue = c.GetImageIDForLastImageStage(importImage)
		} else {
			labelValue = c.GetImageIDForImageStage(importImage, elm.Stage)
		}

		imageServiceCommitChangeOptions.AddLabel(map[string]string{labelKey: labelValue})
	}

	return nil
}
