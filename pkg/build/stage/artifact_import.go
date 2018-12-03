package stage

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"gopkg.in/satori/go.uuid.v1"

	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/slug"
	"github.com/flant/dapp/pkg/util"
)

type getImportsOptions struct {
	Before StageName
	After  StageName
}

func getImports(dimgBaseConfig *config.DimgBase, options *getImportsOptions) []*config.ArtifactImport {
	var imports []*config.ArtifactImport
	for _, elm := range dimgBaseConfig.Import {
		if options.Before != "" && elm.Before != "" && elm.Before == string(options.Before) {
			imports = append(imports, elm)
		} else if options.After != "" && elm.After != "" && elm.After == string(options.After) {
			imports = append(imports, elm)
		}
	}

	return imports
}

func newArtifactImportStage(imports []*config.ArtifactImport, baseStageOptions *NewBaseStageOptions) *ArtifactImportStage {
	s := &ArtifactImportStage{}
	s.imports = imports
	s.BaseStage = newBaseStage(baseStageOptions)
	return s
}

type ArtifactImportStage struct {
	*BaseStage

	imports []*config.ArtifactImport
}

func (s *ArtifactImportStage) GetDependencies(c Conveyor, _ image.Image) (string, error) {
	var args []string

	for _, elm := range s.imports {
		args = append(args, c.GetDimgSignature(elm.ArtifactName))
		args = append(args, elm.Add, elm.To)
		args = append(args, elm.Group, elm.Owner)
		args = append(args, elm.IncludePaths...)
		args = append(args, elm.ExcludePaths...)
	}

	return util.Sha256Hash(args...), nil
}

func (s *ArtifactImportStage) PrepareImage(c Conveyor, _, image image.Image) error {
	for _, elm := range s.imports {
		command, volume, err := s.generateImportData(c, elm)
		if err != nil {
			return err
		}

		image.Container().AddRunCommands(command)
		image.Container().RunOptions().AddVolume(volume)

		imageServiceCommitChangeOptions := image.Container().ServiceCommitChangeOptions()
		imageServiceCommitChangeOptions.AddLabel(map[string]string{
			fmt.Sprintf("dapp-artifact-%s", slug.Slug(elm.ArtifactName)): c.GetDimgSignature(elm.ArtifactName),
		})
	}

	return nil
}

func (s *ArtifactImportStage) generateImportData(c Conveyor, i *config.ArtifactImport) (string, string, error) {
	exportFolderName := uuid.NewV4().String()
	dimgArtifactTmpPath := path.Join(s.dimgTmpDir, i.ArtifactName, exportFolderName)
	dimgArtifactContainerTmpPath := path.Join(s.dimgContainerTmpDir, i.ArtifactName, exportFolderName)

	artifactCommand := generateSafeCp(i.Add, dimgArtifactContainerTmpPath, "", "", []string{}, []string{})

	toolchainContainer, err := dappdeps.ToolchainContainer()
	if err != nil {
		return "", "", err
	}

	baseContainer, err := dappdeps.BaseContainer()
	if err != nil {
		return "", "", err
	}

	args := []string{
		"--rm",
		fmt.Sprintf("--volumes-from=%s", toolchainContainer),
		fmt.Sprintf("--volumes-from=%s", baseContainer),
		fmt.Sprintf("--entrypoint=%s", dappdeps.BaseBinPath("bash")),
		fmt.Sprintf("--volume=%s:%s", dimgArtifactTmpPath, dimgArtifactContainerTmpPath),
		c.GetDimgSignature(i.ArtifactName),
		"-ec",
		image.ShelloutPack(artifactCommand),
	}

	err = docker.CliRun(args...)
	if err != nil {
		return "", "", err
	}

	command := generateSafeCp(dimgArtifactContainerTmpPath, i.To, i.Owner, i.Group, i.IncludePaths, i.ExcludePaths)

	volume := fmt.Sprintf("%s:%s:ro", dimgArtifactTmpPath, dimgArtifactContainerTmpPath)

	return command, volume, nil
}

func generateSafeCp(from, to, owner, group string, includePaths, excludePaths []string) string {
	var args []string

	mkdirBin := dappdeps.BaseBinPath("mkdir")
	mkdirPath := path.Dir(to)
	mkdirCommand := fmt.Sprintf("%s -p %s", mkdirBin, mkdirPath)

	rsyncBin := dappdeps.BaseBinPath("rsync")
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
	var prevPart string

	currentPart := filePath
	for {
		parts = append(parts, currentPart)
		currentPart = path.Dir(prevPart)

		if prevPart == currentPart {
			break
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(parts[:])))

	return parts
}
