package stage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flant/werf/pkg/stapel"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/docker"
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

func (s *ImportsStage) GetDependencies(c Conveyor, _, _ imagePkg.ImageInterface) (string, error) {
	var args []string

	for _, elm := range s.imports {
		if elm.ImageName != "" {
			args = append(args, c.GetImageLatestStageSignature(elm.ImageName))
		} else {
			args = append(args, c.GetImageLatestStageSignature(elm.ArtifactName))
		}

		args = append(args, elm.Add, elm.To)
		args = append(args, elm.Group, elm.Owner)
		args = append(args, elm.IncludePaths...)
		args = append(args, elm.ExcludePaths...)
	}

	return util.Sha256Hash(args...), nil
}

func (s *ImportsStage) PrepareImage(c Conveyor, _, image imagePkg.ImageInterface) error {
	for _, elm := range s.imports {
		importContainerTmpPath := s.importContainerTmpPath(elm)

		artifactTmpPath := path.Join(importContainerTmpPath, path.Base(elm.Add))

		command := generateSafeCp(artifactTmpPath, elm.To, elm.Owner, elm.Group, elm.IncludePaths, elm.ExcludePaths)

		importImageTmpDir, importImageContainerTmpDir := s.importImageTmpDirs(elm)
		volume := fmt.Sprintf("%s:%s:ro", importImageTmpDir, importImageContainerTmpDir)

		image.Container().AddServiceRunCommands(command)
		image.Container().RunOptions().AddVolume(volume)

		imageServiceCommitChangeOptions := image.Container().ServiceCommitChangeOptions()

		var labelKey, labelValue string
		if elm.ImageName != "" {
			labelKey = imagePkg.WerfImportLabelPrefix + slug.Slug(elm.ImageName)
			labelValue = c.GetImageLatestStageSignature(elm.ImageName)
		} else {
			labelKey = imagePkg.WerfImportLabelPrefix + slug.Slug(elm.ArtifactName)
			labelValue = c.GetImageLatestStageSignature(elm.ArtifactName)
		}

		imageServiceCommitChangeOptions.AddLabel(map[string]string{labelKey: labelValue})
	}

	return nil
}

func (s *ImportsStage) PreRunHook(c Conveyor) error {
	for _, elm := range s.imports {
		if err := s.prepareImportData(c, elm); err != nil {
			return err
		}
	}

	return nil
}

func (s *ImportsStage) prepareImportData(c Conveyor, i *config.Import) error {
	importContainerTmpPath := s.importContainerTmpPath(i)

	artifactTmpPath := path.Join(importContainerTmpPath, path.Base(i.Add))

	imageCommand := generateSafeCp(i.Add, artifactTmpPath, "", "", []string{}, []string{})

	stapelContainerName, err := stapel.GetOrCreateContainer()
	if err != nil {
		return err
	}

	importImageTmp, importImageContainerTmp := s.importImageTmpDirs(i)

	var dockerImageName string
	if i.ImageName != "" {
		dockerImageName = c.GetImageLatestStageImageName(i.ImageName)
	} else {
		dockerImageName = c.GetImageLatestStageImageName(i.ArtifactName)
	}

	if err := os.MkdirAll(importImageTmp, os.ModePerm); err != nil {
		return err
	}

	args := []string{
		"--rm",
		"--user=0:0",
		"--workdir=/",
		fmt.Sprintf("--volumes-from=%s", stapelContainerName),
		fmt.Sprintf("--entrypoint=%s", stapel.BashBinPath()),
		fmt.Sprintf("--volume=%s:%s", importImageTmp, importImageContainerTmp),
		dockerImageName,
		"-ec",
		imagePkg.ShelloutPack(imageCommand),
	}

	err = docker.CliRun(args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *ImportsStage) importContainerTmpPath(i *config.Import) string {
	importID := util.Sha256Hash(fmt.Sprintf("%+v", i))
	_, importImageContainerTmpPath := s.importImageTmpDirs(i)
	importContainerTmpPath := path.Join(importImageContainerTmpPath, importID)

	return importContainerTmpPath
}

func (s *ImportsStage) importImageTmpDirs(i *config.Import) (string, string) {
	var importNamePathPart string
	if i.ImageName != "" {
		importNamePathPart = slug.Slug(i.ImageName)
	} else {
		importNamePathPart = slug.Slug(i.ArtifactName)
	}

	importImageTmpDir := filepath.Join(s.imageTmpDir, "import", importNamePathPart)
	importImageContainerTmpDir := path.Join(s.containerWerfDir, "import", importNamePathPart)

	return importImageTmpDir, importImageContainerTmpDir
}

func generateSafeCp(from, to, owner, group string, includePaths, excludePaths []string) string {
	var args []string

	mkdirBin := stapel.MkdirBinPath()
	mkdirPath := path.Dir(to)
	mkdirCommand := fmt.Sprintf("%s -p %s", mkdirBin, mkdirPath)

	rsyncBin := stapel.RsyncBinPath()
	var rsyncChownOption string
	if owner != "" || group != "" {
		rsyncChownOption = fmt.Sprintf("--chown=%s:%s", owner, group)
	}
	rsyncCommand := fmt.Sprintf("%s --archive --links --inplace %s", rsyncBin, rsyncChownOption)

	if len(includePaths) != 0 {
		/**
				Если указали include_paths — это означает, что надо копировать
				только указанные пути. Поэтому exclude_paths в приоритете, т.к. в данном режиме
		        exclude_paths может относится только к путям, указанным в include_paths.
		        При этом случай, когда в include_paths указали более специальный путь, чем в exclude_paths,
		        будет обрабатываться в пользу exclude, этот путь не скопируется.
		*/
		for _, p := range excludePaths {
			rsyncCommand += fmt.Sprintf(" --filter='-/ %s'", path.Join(from, p))
		}

		for _, p := range includePaths {
			targetPath := path.Join(from, p)

			// Генерируем разрешающее правило для каждого элемента пути
			for _, pathPart := range descentPath(targetPath) {
				rsyncCommand += fmt.Sprintf(" --filter='+/ %s'", pathPart)
			}

			/**
					На данный момент не знаем директорию или файл имел в виду пользователь,
			        поэтому подставляем фильтры для обоих возможных случаев.

					Автоматом подставляем паттерн ** для включения файлов, содержащихся в
			        директории, которую пользователь указал в include_paths.
			*/
			rsyncCommand += fmt.Sprintf(" --filter='+/ %s'", targetPath)
			rsyncCommand += fmt.Sprintf(" --filter='+/ %s'", path.Join(targetPath, "**"))
		}

		// Все что не подошло по include — исключается
		rsyncCommand += fmt.Sprintf(" --filter='-/ %s'", path.Join(from, "**"))
	} else {
		for _, p := range excludePaths {
			rsyncCommand += fmt.Sprintf(" --filter='-/ %s'", path.Join(from, p))
		}
	}

	/**
		Слэш после from — это инструкция rsync'у для копирования
	    содержимого директории from, а не самой директории.
	*/
	rsyncCommand += fmt.Sprintf(" $(if [ -d %[1]s ] ; then echo %[1]s/ ; else echo %[1]s ; fi) %[2]s", from, to)

	args = append(args, mkdirCommand, rsyncCommand)
	command := strings.Join(args, " && ")

	return command
}

func descentPath(filePath string) []string {
	var parts []string

	part := filePath
	for {
		parts = append(parts, part)
		part = path.Dir(part)

		if part == path.Dir(part) {
			break
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(parts[:])))

	return parts
}
