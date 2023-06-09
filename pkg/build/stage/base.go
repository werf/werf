package stage

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker_registry"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

type StageName string

const (
	From                      StageName = "from"
	BeforeInstall             StageName = "beforeInstall"
	DependenciesBeforeInstall StageName = "dependenciesBeforeInstall"
	GitArchive                StageName = "gitArchive"
	Install                   StageName = "install"
	DependenciesAfterInstall  StageName = "dependenciesAfterInstall"
	BeforeSetup               StageName = "beforeSetup"
	DependenciesBeforeSetup   StageName = "dependenciesBeforeSetup"
	Setup                     StageName = "setup"
	DependenciesAfterSetup    StageName = "dependenciesAfterSetup"
	GitCache                  StageName = "gitCache"
	GitLatestPatch            StageName = "gitLatestPatch"
	DockerInstructions        StageName = "dockerInstructions"

	Dockerfile StageName = "dockerfile"
)

// TODO(compatibility): remove in v1.3
func GetLegacyCompatibleStageName(name StageName) string {
	switch name {
	case DependenciesBeforeInstall:
		return "importsBeforeInstall"
	case DependenciesAfterInstall:
		return "importsAfterInstall"
	case DependenciesBeforeSetup:
		return "importsBeforeSetup"
	case DependenciesAfterSetup:
		return "importsAfterSetup"
	default:
		return string(name)
	}
}

var AllStages = []StageName{
	From,
	BeforeInstall,
	DependenciesBeforeInstall,
	GitArchive,
	Install,
	DependenciesAfterInstall,
	BeforeSetup,
	DependenciesBeforeSetup,
	Setup,
	DependenciesAfterSetup,
	GitCache,
	GitLatestPatch,
	DockerInstructions,

	Dockerfile,
}

type BaseStageOptions struct {
	LogName          string
	TargetPlatform   string
	ImageName        string
	ConfigMounts     []*config.Mount
	ImageTmpDir      string
	ContainerWerfDir string
	ProjectName      string
}

func NewBaseStage(name StageName, options *BaseStageOptions) *BaseStage {
	s := &BaseStage{}
	s.name = name
	s.logName = options.LogName
	s.targetPlatform = options.TargetPlatform
	s.imageName = options.ImageName
	s.configMounts = options.ConfigMounts
	s.imageTmpDir = options.ImageTmpDir
	s.containerWerfDir = options.ContainerWerfDir
	s.projectName = options.ProjectName
	return s
}

type BaseStage struct {
	name             StageName
	logName          string
	targetPlatform   string
	imageName        string
	digest           string
	contentDigest    string
	stageImage       *StageImage
	gitMappings      []*GitMapping
	imageTmpDir      string
	containerWerfDir string
	configMounts     []*config.Mount
	projectName      string
}

func (s *BaseStage) HasPrevStage() bool {
	return true
}

func (s *BaseStage) IsStapelStage() bool {
	return true
}

func (s *BaseStage) LogDetailedName() string {
	imageName := s.imageName
	if imageName == "" {
		imageName = "~"
	}

	return fmt.Sprintf("%s/%s", imageName, s.LogName())
}

func (s *BaseStage) TargetPlatform() string {
	return s.targetPlatform
}

func (s *BaseStage) ImageName() string {
	return s.imageName
}

func (s *BaseStage) LogName() string {
	if s.logName != "" {
		return s.logName
	}
	return string(s.Name())
}

func (s *BaseStage) Name() StageName {
	if s.name != "" {
		return s.name
	}

	panic("name must be defined!")
}

func (s *BaseStage) ExpandDependencies(ctx context.Context, c Conveyor, baseEnv map[string]string) error {
	return nil
}

func (s *BaseStage) FetchDependencies(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _ docker_registry.GenericApiInterface) error {
	return nil
}

func (s *BaseStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	panic("method must be implemented!")
}

func (s *BaseStage) GetNextStageDependencies(_ context.Context, _ Conveyor) (string, error) {
	return "", nil
}

func (s *BaseStage) getNextStageGitDependencies(ctx context.Context, c Conveyor) (string, error) {
	var args []string
	for _, gitMapping := range s.gitMappings {
		if s.stageImage != nil && s.stageImage.Image.GetStageDescription() != nil {
			if commitInfo, err := gitMapping.GetBuiltImageCommitInfo(s.stageImage.Image.GetStageDescription().Info.Labels); err != nil {
				return "", fmt.Errorf("unable to get built image commit info from image %s: %w", s.stageImage.Image.Name(), err)
			} else {
				args = append(args, commitInfo.Commit)
			}
		} else {
			latestCommitInfo, err := gitMapping.GetLatestCommitInfo(ctx, c)
			if err != nil {
				return "", fmt.Errorf("unable to get latest commit of git mapping %s: %w", gitMapping.Name, err)
			}
			args = append(args, latestCommitInfo.Commit)
		}
	}

	logboek.Context(ctx).Debug().LogF("Stage %q next stage dependencies: %#v\n", s.LogName(), args)
	sort.Strings(args)

	return util.Sha256Hash(args...), nil
}

func (s *BaseStage) IsEmpty(_ context.Context, _ Conveyor, _ *StageImage) (bool, error) {
	return false, nil
}

func (s *BaseStage) selectStageByOldestCreationTimestamp(stages []*imagePkg.StageDescription) (*imagePkg.StageDescription, error) {
	var oldestStage *imagePkg.StageDescription
	for _, stageDesc := range stages {
		if oldestStage == nil {
			oldestStage = stageDesc
		} else if stageDesc.StageID.UniqueIDAsTime().Before(oldestStage.StageID.UniqueIDAsTime()) {
			oldestStage = stageDesc
		}
	}
	return oldestStage, nil
}

func (s *BaseStage) selectStagesAncestorsByGitMappings(ctx context.Context, c Conveyor, stages []*imagePkg.StageDescription) ([]*imagePkg.StageDescription, error) {
	var suitableStages []*imagePkg.StageDescription
	var currentCommitsByIndex []string

	for _, gitMapping := range s.gitMappings {
		currentCommitInfo, err := gitMapping.GetLatestCommitInfo(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("error getting latest commit of git mapping %s: %w", gitMapping.Name, err)
		}

		var currentCommit string
		if currentCommitInfo.VirtualMerge {
			currentCommit = currentCommitInfo.VirtualMergeFromCommit
		} else {
			currentCommit = currentCommitInfo.Commit
		}

		currentCommitsByIndex = append(currentCommitsByIndex, currentCommit)
	}

ScanImages:
	for _, stageDesc := range stages {
		for i, gitMapping := range s.gitMappings {
			currentCommit := currentCommitsByIndex[i]

			imageCommitInfo, err := gitMapping.GetBuiltImageCommitInfo(stageDesc.Info.Labels)
			if err != nil {
				logboek.Context(ctx).Warn().LogF("Ignore stage %s: unable to get image commit info for git repo %s: %s", stageDesc.Info.Name, gitMapping.GitRepo().String(), err)
				continue ScanImages
			}

			var commitToCheckAncestry string
			if imageCommitInfo.VirtualMerge {
				commitToCheckAncestry = imageCommitInfo.VirtualMergeFromCommit
			} else {
				commitToCheckAncestry = imageCommitInfo.Commit
			}

			isOurAncestor, err := gitMapping.GitRepo().IsAncestor(ctx, commitToCheckAncestry, currentCommit)
			if err != nil {
				return nil, fmt.Errorf("error checking commits ancestry %s<-%s: %w", commitToCheckAncestry, currentCommit, err)
			}

			if !isOurAncestor {
				logboek.Context(ctx).Debug().LogF("%s is not ancestor of %s for git repo %s: ignore image %s\n", commitToCheckAncestry, currentCommit, gitMapping.GitRepo().String(), stageDesc.Info.Name)
				continue ScanImages
			}

			logboek.Context(ctx).Debug().LogF(
				"%s is ancestor of %s for git repo %s: image %s is suitable for git archive stage\n",
				commitToCheckAncestry, currentCommit, gitMapping.GitRepo().String(), stageDesc.Info.Name,
			)
		}

		suitableStages = append(suitableStages, stageDesc)
	}

	return suitableStages, nil
}

func (s *BaseStage) SelectSuitableStage(_ context.Context, c Conveyor, stages []*imagePkg.StageDescription) (*imagePkg.StageDescription, error) {
	return s.selectStageByOldestCreationTimestamp(stages)
}

func (s *BaseStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	/*
	 * NOTE: BaseStage.PrepareImage does not called in From.PrepareImage.
	 * NOTE: Take into account when adding new base PrepareImage steps.
	 */

	addLabels := map[string]string{imagePkg.WerfProjectRepoCommitLabel: c.GiterminismManager().HeadCommit()}
	if c.UseLegacyStapelBuilder(cb) {
		stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions().AddLabel(addLabels)
	} else {
		stageImage.Builder.StapelStageBuilder().AddLabels(addLabels)
	}

	serviceMounts := s.getServiceMounts(prevBuiltImage)
	s.addServiceMountsLabels(serviceMounts, c, cb, stageImage)
	if err := s.addServiceMountsVolumes(serviceMounts, c, cb, stageImage, false); err != nil {
		return fmt.Errorf("error adding mounts volumes: %w", err)
	}

	customMounts := s.getCustomMounts(prevBuiltImage)
	s.addCustomMountLabels(customMounts, c, cb, stageImage)
	if err := s.addCustomMountVolumes(customMounts, c, cb, stageImage, false); err != nil {
		return fmt.Errorf("error adding mounts volumes: %w", err)
	}

	return nil
}

func (s *BaseStage) PreRun(_ context.Context, _ Conveyor) error {
	return nil
}

func (s *BaseStage) getServiceMounts(prevBuiltImage *StageImage) map[string][]string {
	return mergeMounts(s.getServiceMountsFromLabels(prevBuiltImage), s.getServiceMountsFromConfig())
}

func (s *BaseStage) getServiceMountsFromLabels(prevBuiltImage *StageImage) map[string][]string {
	mountpointsByType := map[string][]string{}

	var labels map[string]string
	if prevBuiltImage != nil {
		labels = prevBuiltImage.Image.GetStageDescription().Info.Labels
	}

	for _, labelMountType := range []struct{ Label, MountType string }{
		{imagePkg.WerfMountTmpDirLabel, "tmp_dir"},
		{imagePkg.WerfMountBuildDirLabel, "build_dir"},
	} {
		v, hasKey := labels[labelMountType.Label]
		if !hasKey {
			continue
		}

		mountpoints := util.RejectEmptyStrings(util.UniqStrings(strings.Split(v, ";")))
		mountpointsByType[labelMountType.MountType] = mountpoints
	}

	return mountpointsByType
}

func (s *BaseStage) getServiceMountsFromConfig() map[string][]string {
	mountpointsByType := map[string][]string{}

	for _, mountCfg := range s.configMounts {
		if !util.IsStringsContainValue([]string{"tmp_dir", "build_dir"}, mountCfg.Type) {
			continue
		}

		mountpoint := path.Clean(mountCfg.To)
		mountpointsByType[mountCfg.Type] = append(mountpointsByType[mountCfg.Type], mountpoint)
	}

	return mountpointsByType
}

func (s *BaseStage) addServiceMountsVolumes(mountpointsByType map[string][]string, c Conveyor, cr container_backend.ContainerBackend, stageImage *StageImage, cleanupMountpoints bool) error {
	for mountType, mountpoints := range mountpointsByType {
		for _, mountpoint := range mountpoints {
			absoluteMountpoint := path.Join("/", mountpoint)

			var absoluteFrom string
			switch mountType {
			case "tmp_dir":
				absoluteFrom = filepath.Join(s.imageTmpDir, "mount", slug.LimitedSlug(absoluteMountpoint, slug.DefaultSlugMaxSize))
			case "build_dir":
				absoluteFrom = filepath.Join(werf.GetSharedContextDir(), "mounts", "projects", s.projectName, slug.LimitedSlug(absoluteMountpoint, slug.DefaultSlugMaxSize))
			default:
				panic(fmt.Sprintf("unknown mount type %s", mountType))
			}

			err := os.MkdirAll(absoluteFrom, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating tmp path %s for mount: %w", absoluteFrom, err)
			}

			volume := fmt.Sprintf("%s:%s", absoluteFrom, absoluteMountpoint)
			if c.UseLegacyStapelBuilder(cr) {
				stageImage.Builder.LegacyStapelStageBuilder().Container().RunOptions().AddVolume(volume)
			} else {
				stageImage.Builder.StapelStageBuilder().AddBuildVolumes(volume)
				if cleanupMountpoints {
					stageImage.Builder.StapelStageBuilder().RemoveData(container_backend.RemoveInsidePath, []string{absoluteMountpoint}, nil)
				}
			}
		}
	}

	return nil
}

func (s *BaseStage) addServiceMountsLabels(mountpointsByType map[string][]string, c Conveyor, cr container_backend.ContainerBackend, stageImage *StageImage) {
	for mountType, mountpoints := range mountpointsByType {
		var labelName string
		switch mountType {
		case "tmp_dir":
			labelName = imagePkg.WerfMountTmpDirLabel
		case "build_dir":
			labelName = imagePkg.WerfMountBuildDirLabel
		default:
			panic(fmt.Sprintf("unknown mount type %s", mountType))
		}

		labelValue := strings.Join(mountpoints, ";")

		addLabels := map[string]string{labelName: labelValue}
		if c.UseLegacyStapelBuilder(cr) {
			stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions().AddLabel(addLabels)
		} else {
			stageImage.Builder.StapelStageBuilder().AddLabels(addLabels)
		}
	}
}

func (s *BaseStage) getCustomMounts(prevBuiltImage *StageImage) map[string][]string {
	return mergeMounts(s.getCustomMountsFromLabels(prevBuiltImage), s.getCustomMountsFromConfig())
}

func (s *BaseStage) getCustomMountsFromLabels(prevBuiltImage *StageImage) map[string][]string {
	mountpointsByFrom := map[string][]string{}

	var labels map[string]string
	if prevBuiltImage != nil {
		labels = prevBuiltImage.Image.GetStageDescription().Info.Labels
	}
	for k, v := range labels {
		if !strings.HasPrefix(k, imagePkg.WerfMountCustomDirLabelPrefix) {
			continue
		}

		parts := strings.SplitN(k, imagePkg.WerfMountCustomDirLabelPrefix, 2)
		fromPath := strings.ReplaceAll(parts[1], "--", "/")
		fromFilepath := filepath.FromSlash(fromPath)

		mountpoints := util.RejectEmptyStrings(util.UniqStrings(strings.Split(v, ";")))
		mountpointsByFrom[fromFilepath] = mountpoints
	}

	return mountpointsByFrom
}

func (s *BaseStage) getCustomMountsFromConfig() map[string][]string {
	mountpointsByFrom := map[string][]string{}
	for _, mountCfg := range s.configMounts {
		if mountCfg.Type != "custom_dir" {
			continue
		}

		from := filepath.Clean(mountCfg.From)
		mountpoint := path.Clean(mountCfg.To)

		mountpointsByFrom[from] = util.UniqAppendString(mountpointsByFrom[from], mountpoint)
	}

	return mountpointsByFrom
}

func (s *BaseStage) addCustomMountVolumes(mountpointsByFrom map[string][]string, c Conveyor, cr container_backend.ContainerBackend, stageImage *StageImage, cleanupMountpoints bool) error {
	for from, mountpoints := range mountpointsByFrom {
		absoluteFrom := util.ExpandPath(from)

		exist, err := util.FileExists(absoluteFrom)
		if err != nil {
			return err
		}

		if !exist {
			err := os.MkdirAll(absoluteFrom, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating %s: %w", absoluteFrom, err)
			}
		}

		for _, mountpoint := range mountpoints {
			absoluteMountpoint := path.Join("/", mountpoint)

			volume := fmt.Sprintf("%s:%s", absoluteFrom, absoluteMountpoint)
			if c.UseLegacyStapelBuilder(cr) {
				stageImage.Builder.LegacyStapelStageBuilder().Container().RunOptions().AddVolume(volume)
			} else {
				stageImage.Builder.StapelStageBuilder().AddBuildVolumes(volume)
				if cleanupMountpoints {
					stageImage.Builder.StapelStageBuilder().RemoveData(container_backend.RemoveInsidePath, []string{absoluteMountpoint}, nil)
				}
			}
		}
	}

	return nil
}

func (s *BaseStage) addCustomMountLabels(mountpointsByFrom map[string][]string, c Conveyor, cr container_backend.ContainerBackend, stageImage *StageImage) {
	for from, mountpoints := range mountpointsByFrom {
		labelName := fmt.Sprintf("%s%s", imagePkg.WerfMountCustomDirLabelPrefix, strings.ReplaceAll(filepath.ToSlash(from), "/", "--"))
		labelValue := strings.Join(mountpoints, ";")

		addLabels := map[string]string{labelName: labelValue}
		if c.UseLegacyStapelBuilder(cr) {
			stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions().AddLabel(addLabels)
		} else {
			stageImage.Builder.StapelStageBuilder().AddLabels(addLabels)
		}
	}
}

func (s *BaseStage) SetDigest(digest string) {
	s.digest = digest
}

func (s *BaseStage) GetDigest() string {
	return s.digest
}

func (s *BaseStage) SetContentDigest(contentDigest string) {
	s.contentDigest = contentDigest
}

func (s *BaseStage) GetContentDigest() string {
	return s.contentDigest
}

func (s *BaseStage) SetStageImage(stageImage *StageImage) {
	s.stageImage = stageImage
}

func (s *BaseStage) GetStageImage() *StageImage {
	return s.stageImage
}

func (s *BaseStage) SetGitMappings(gitMappings []*GitMapping) {
	s.gitMappings = gitMappings
}

func (s *BaseStage) GetGitMappings() []*GitMapping {
	return s.gitMappings
}

func (s *BaseStage) UsesBuildContext() bool {
	return false
}

func mergeMounts(a, b map[string][]string) map[string][]string {
	res := map[string][]string{}

	for k, mountpoints := range a {
		res[k] = mountpoints
	}
	for k, mountpoints := range b {
		res[k] = util.UniqStrings(append(res[k], mountpoints...))
	}

	return res
}
