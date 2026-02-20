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
	"github.com/werf/werf/v2/pkg/build/import_server"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/image"
	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/stapel"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var ErrNothingToImport = fmt.Errorf("nothing to import")

const nothingChecksum = "d41d8cd98f00b204e9800998ecf8427e"

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

				// TODO: in v3 we should return err instead of warning
				if sourceChecksum == nothingChecksum {
					global_warnings.GlobalWarningLn(ctx, fmt.Sprintf("This import config does nothing: %s", formatImportTitle(elm)))
				}

				logboek.Context(ctx).Default().LogF("%s: %s\n", sourceChecksum, formatImportTitle(elm))

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

func formatImportTitle(elm *config.Import) string {
	title := fmt.Sprintf("image=%s add=%s to=%s", elm.ImageName, elm.Add, elm.To)
	if len(elm.IncludePaths) != 0 {
		title += fmt.Sprintf(" includePaths=%v", elm.IncludePaths)
	}
	if len(elm.ExcludePaths) != 0 {
		title += fmt.Sprintf(" excludePaths=%v", elm.ExcludePaths)
	}
	return fmt.Sprintf("import[%s]", title)
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

		var sourceImageName string

		if elm.ExternalImage {
			sourceImageName = elm.ImageName
		} else {
			sourceImageConfigName := getSourceImageName(elm)
			if elm.Stage == "" {
				sourceImageName = c.GetImageNameForLastImageStage(s.targetPlatform, sourceImageConfigName)
			} else {
				sourceImageName = c.GetImageNameForImageStage(s.targetPlatform, sourceImageConfigName, elm.Stage)
			}
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

// generateChecksumScript prepares a shell script that calculates a checksum over the
// set of files that would be imported from `from` with the given include/exclude rules.
//
// IMPORTANT: the embedded rsync binary is an old 3.1.2 build with limited and
// unreliable support for options like --out-format when combined with --dry-run.
// We cannot simply ask rsync to print "just file paths" in a machine-friendly format.
//
// To keep the checksum aligned with the actual import logic, we reuse rsync and the same
// filters as the import server does (PrepareRsyncFilters).
//
// The script uses two-phase approach matching the import logic:
// Phase 1: Use 'find' to get directories that directly match include globs
//
//	(rsync with --prune-empty-dirs cannot detect empty target directories in --dry-run mode)
//
// Phase 2: Use rsync to get files matching the globs (using PrepareRsyncFilters)
// This ensures empty directories that match globs like "app/**/add-dir" are included in checksum.
func generateChecksumScript(from string, includePaths, excludePaths []string, resultChecksumPath string) []string {
	// As we are running rsync from the container root, we need to make sure that the paths
	// we are passing to rsync are relative to the container root.
	var includePathsCopy []string
	var excludePathsCopy []string
	{
		if len(includePaths) == 0 {
			includePathsCopy = append(includePathsCopy, from)
		} else {
			for _, includePath := range includePaths {
				includePathsCopy = append(includePathsCopy, path.Join(from, includePath))
			}
		}

		for _, excludePath := range excludePaths {
			excludePathsCopy = append(excludePathsCopy, path.Join(from, excludePath))
		}
	}

	// Exclude the stapel container mount root, as in the previous implementation.
	if from == "/" {
		excludePathsCopy = append(excludePathsCopy, stapel.CONTAINER_MOUNT_ROOT)
	}

	// We have an old rsync version, so we can't use --out-format and other options to parse file paths.
	// Example lines:
	// "-rw-r--r--    1 root     root             0 Nov 17 22:57 test-file.a"
	// "drwxr-xr-x    1 root     root             0 Nov 17 22:57 some-directory"
	// "lrw-r--r--    1 root     root             0 Dec 10 10:07 path/to/link -> target/path"
	//
	// NOTE: We include directories (d*) in the checksum calculation. This ensures checksum
	// consistency with what gets copied. For file globs like "**/*.txt", only parent directories
	// of matching files are included. For directory globs like "app/**/add-dir", the directories
	// themselves are included even if empty (via Phase 1 find command).
	parseFilePathCommand := `while read -r mode rest; do
	 case "$mode" in
	   -*) echo "/${rest##* }" ;;
	   d*) echo "/${rest##* }/" ;;
	   l*)
	     name="${rest%% -> *}"
	     name="${name##* }"
	     echo "/$name"
	     ;;
	 esac
	done`

	var commands string

	// phase 1: Use 'find' to get directories that directly match include globs
	// this is necessary because rsync with --prune-empty-dirs removes empty directories
	// from output even when they match the include pattern.
	//
	// optimized approach:
	// - for patterns with **: use ONE find command with multiple -name conditions
	// - for simple paths: batch all existence checks together
	if len(includePaths) > 0 {
		commands = "{ " + generateFirstPhaseFindCommand(from, includePathsCopy) + "; "
	} else {
		commands = "{ "
	}

	// phase 2: Get files matching the globs using rsync (standard behavior)
	// do not follow symlinks when calculating checksums to avoid runner hang-ups (-L)
	rsyncFilesCommand := stapel.RsyncBinPath() + " -r --dry-run"
	// run rsync from the container root to avoid problems with relative paths in the output.
	rsyncFilesCommand += import_server.PrepareRsyncFilters("", includePathsCopy, excludePathsCopy)
	rsyncFilesCommand += " " + "/"
	commands += rsyncFilesCommand + " | " + parseFilePathCommand + "; }"

	sortCommand := fmt.Sprintf("%s -u", stapel.SortBinPath()) // -u for unique (remove duplicates)
	md5SumCommand := stapel.Md5sumBinPath()
	cutCommand := fmt.Sprintf("%s -c 1-32", stapel.CutBinPath())

	allCommands := []string{commands, sortCommand, "checksum", md5SumCommand, cutCommand}

	script := generateChecksumBashFunction()
	script = append(script, fmt.Sprintf("%s > %s", strings.Join(allCommands, " | "), resultChecksumPath))

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

// generateFirstPhaseFindCommand creates an optimized shell command to find directories
// matching the include patterns. Instead of running multiple find commands (one per pattern),
// this function:
// 1. groups patterns with ** by their common prefix to minimize filesystem traversals
// 2. uses a single find with multiple -name conditions where possible
// 3. batches simple path checks together
//
// this is important for performance when the source directory contains hundreds of
// thousands of files.
func generateFirstPhaseFindCommand(from string, includePathsCopy []string) string {
	if len(includePathsCopy) == 0 {
		return "true"
	}

	baseDir := from
	if baseDir == "" {
		baseDir = "/"
	}

	// separate glob patterns (with **) from simple paths
	type globPattern struct {
		prefix      string // path before **
		dirName     string // directory name to find (after **)
		fullPattern string // original pattern for path filtering
	}

	var globPatterns []globPattern
	var simplePaths []string

	for _, p := range includePathsCopy {
		if strings.Contains(p, "**") {
			parts := strings.SplitN(p, "**", 2)
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")
			if suffix == "" {
				continue
			}
			// for patterns like "app/**/sub/dir", we need the last segment
			dirName := path.Base(suffix)
			globPatterns = append(globPatterns, globPattern{
				prefix:      prefix,
				dirName:     dirName,
				fullPattern: p,
			})
		} else {
			simplePaths = append(simplePaths, p)
		}
	}

	var commands []string

	// group glob patterns by prefix to run fewer find commands
	if len(globPatterns) > 0 {
		// group patterns by their prefix
		prefixGroups := make(map[string][]globPattern)
		for _, gp := range globPatterns {
			prefixGroups[gp.prefix] = append(prefixGroups[gp.prefix], gp)
		}

		for prefix, patterns := range prefixGroups {
			// prefix already contains the full path,
			// so use it directly as searchDir
			searchDir := prefix
			if searchDir == "" {
				searchDir = baseDir
			}

			// build find command with multiple -name conditions (OR)
			var nameConditions []string
			var pathPrefixes []string
			for _, gp := range patterns {
				nameConditions = append(nameConditions, fmt.Sprintf("-name '%s'", gp.dirName))
				// calculate path prefix for filtering
				pathPrefix := strings.TrimSuffix(strings.SplitN(gp.fullPattern, "**", 2)[0], "/")
				if pathPrefix == "" {
					pathPrefix = baseDir
				}
				pathPrefixes = append(pathPrefixes, pathPrefix)
			}

			// combine name conditions with -o (OR)
			nameExpr := strings.Join(nameConditions, " -o ")
			if len(nameConditions) > 1 {
				nameExpr = "\\( " + nameExpr + " \\)"
			}

			// build case pattern for path filtering
			var casePatterns []string
			for _, pfx := range pathPrefixes {
				casePatterns = append(casePatterns, pfx+"*")
			}
			casePattern := strings.Join(casePatterns, "|")

			findCmd := fmt.Sprintf(
				"find %s -type d %s 2>/dev/null | while read d; do case \"$d\" in %s) echo \"$d/\" ;; esac; done",
				searchDir, nameExpr, casePattern)
			commands = append(commands, findCmd)
		}
	}

	// handle simple paths - just check if directory exists
	for _, p := range simplePaths {
		commands = append(commands, fmt.Sprintf("[ -d '%s' ] && echo '%s/' || true", p, p))
	}

	if len(commands) == 0 {
		return "true"
	}

	return strings.Join(commands, "; ")
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
	if importElm.ExternalImage {
		return nil
	}

	sourceImageName := getSourceImageName(importElm)
	if importElm.Stage == "" {
		return c.FetchLastNonEmptyImageStage(ctx, targetPlatform, sourceImageName)
	} else {
		return c.FetchImageStage(ctx, targetPlatform, sourceImageName, importElm.Stage)
	}
}

func getSourceImageDockerImageName(c Conveyor, targetPlatform string, importElm *config.Import) string {
	if importElm.ExternalImage {
		return importElm.ImageName
	}
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
	if importElm.ExternalImage {
		return fmt.Sprintf("%s:%s", image.WerfImportSourceExternalImagePrefix, importElm.ImageName)
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
		return fmt.Sprintf("%s:%s", image.WerfImportSourceExternalImagePrefix, importElm.ImageName)
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
