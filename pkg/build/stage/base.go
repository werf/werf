package stage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"

	"github.com/flant/logboek"

	"github.com/werf/werf/pkg/config"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

type StageName string

const (
	From                 StageName = "from"
	BeforeInstall        StageName = "beforeInstall"
	ImportsBeforeInstall StageName = "importsBeforeInstall"
	GitArchive           StageName = "gitArchive"
	Install              StageName = "install"
	ImportsAfterInstall  StageName = "importsAfterInstall"
	BeforeSetup          StageName = "beforeSetup"
	ImportsBeforeSetup   StageName = "importsBeforeSetup"
	Setup                StageName = "setup"
	ImportsAfterSetup    StageName = "importsAfterSetup"
	GitCache             StageName = "gitCache"
	GitLatestPatch       StageName = "gitLatestPatch"
	DockerInstructions   StageName = "dockerInstructions"

	Dockerfile StageName = "dockerfile"
)

var (
	AllStages = []StageName{
		From,
		BeforeInstall,
		ImportsBeforeInstall,
		GitArchive,
		Install,
		ImportsAfterInstall,
		BeforeSetup,
		ImportsBeforeSetup,
		Setup,
		ImportsAfterSetup,
		GitCache,
		GitLatestPatch,
		DockerInstructions,

		Dockerfile,
	}
)

type NewBaseStageOptions struct {
	ImageName        string
	ConfigMounts     []*config.Mount
	ImageTmpDir      string
	ContainerWerfDir string
	ProjectName      string
}

func newBaseStage(name StageName, options *NewBaseStageOptions) *BaseStage {
	s := &BaseStage{}
	s.name = name
	s.imageName = options.ImageName
	s.configMounts = options.ConfigMounts
	s.imageTmpDir = options.ImageTmpDir
	s.containerWerfDir = options.ContainerWerfDir
	s.projectName = options.ProjectName
	return s
}

type BaseStage struct {
	name             StageName
	imageName        string
	signature        string
	contentSignature string
	image            container_runtime.ImageInterface
	gitMappings      []*GitMapping
	imageTmpDir      string
	containerWerfDir string
	configMounts     []*config.Mount
	projectName      string
}

func (s *BaseStage) LogDetailedName() string {
	imageName := s.imageName
	if imageName == "" {
		imageName = "~"
	}

	return fmt.Sprintf("%s/%s", imageName, s.Name())
}

func (s *BaseStage) Name() StageName {
	if s.name != "" {
		return s.name
	}

	panic("name must be defined!")
}

func (s *BaseStage) GetDependencies(_ Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	panic("method must be implemented!")
}

func (s *BaseStage) GetNextStageDependencies(_ Conveyor) (string, error) {
	return "", nil
}

func (s *BaseStage) getNextStageGitDependencies(c Conveyor) (string, error) {
	var args []string
	for _, gitMapping := range s.gitMappings {
		if s.image.GetStageDescription() != nil {
			if commitInfo, err := gitMapping.GetBuiltImageCommitInfo(s.image.GetStageDescription().Info.Labels); err != nil {
				return "", fmt.Errorf("unable to get built image commit info from image %s: %s", s.image.Name(), err)
			} else {
				args = append(args, commitInfo.Commit)
			}
		} else {
			latestCommitInfo, err := gitMapping.GetLatestCommitInfo(c)
			if err != nil {
				return "", fmt.Errorf("unable to get latest commit of git mapping %s: %s", gitMapping.Name, err)
			}
			args = append(args, latestCommitInfo.Commit)
		}
	}

	logboek.Debug.LogF("Stage %q next stage dependencies: %#v\n", s.Name(), args)
	sort.Strings(args)

	return util.Sha256Hash(args...), nil
}

func (s *BaseStage) IsEmpty(_ Conveyor, _ container_runtime.ImageInterface) (bool, error) {
	return false, nil
}

func (s *BaseStage) selectStageByOldestCreationTimestamp(stages []*image.StageDescription) (*image.StageDescription, error) {
	var oldestStage *image.StageDescription
	for _, stageDesc := range stages {
		if oldestStage == nil {
			oldestStage = stageDesc
		} else if stageDesc.StageID.UniqueIDAsTime().Before(oldestStage.StageID.UniqueIDAsTime()) {
			oldestStage = stageDesc
		}
	}
	return oldestStage, nil
}

func (s *BaseStage) selectStagesAncestorsByGitMappings(c Conveyor, stages []*image.StageDescription) ([]*image.StageDescription, error) {
	var suitableStages []*image.StageDescription
	var currentCommitsByIndex []string

	for _, gitMapping := range s.gitMappings {
		currentCommitInfo, err := gitMapping.GetLatestCommitInfo(c)
		if err != nil {
			return nil, fmt.Errorf("error getting latest commit of git mapping %s: %s", gitMapping.Name, err)
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
				logboek.LogErrorF("Ignore stage %s: unable to get image commit info for git repo %s: %s", stageDesc.Info.Name, gitMapping.GitRepo().String(), err)
				continue ScanImages
			}

			var commitToCheckAncestry string
			if imageCommitInfo.VirtualMerge {
				commitToCheckAncestry = imageCommitInfo.VirtualMergeFromCommit
			} else {
				commitToCheckAncestry = imageCommitInfo.Commit
			}

			isOurAncestor, err := gitMapping.GitRepo().IsAncestor(commitToCheckAncestry, currentCommit)
			if err != nil {
				return nil, fmt.Errorf("error checking commits ancestry %s<-%s: %s", commitToCheckAncestry, currentCommit, err)
			}

			if !isOurAncestor {
				logboek.Debug.LogF("%s is not ancestor of %s for git repo %s: ignore image %s\n", commitToCheckAncestry, currentCommit, gitMapping.GitRepo().String(), stageDesc.Info.Name)
				continue ScanImages
			}

			logboek.Debug.LogF(
				"%s is ancestor of %s for git repo %s: image %s is suitable for git archive stage\n",
				commitToCheckAncestry, currentCommit, gitMapping.GitRepo().String(), stageDesc.Info.Name,
			)
		}

		suitableStages = append(suitableStages, stageDesc)
	}

	return suitableStages, nil
}

func (s *BaseStage) SelectSuitableStage(c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error) {
	return s.selectStageByOldestCreationTimestamp(stages)
}

func (s *BaseStage) PrepareImage(_ Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	/*
	 * NOTE: BaseStage.PrepareImage does not called in From.PrepareImage.
	 * NOTE: Take into account when adding new base PrepareImage steps.
	 */

	serviceMounts := s.getServiceMounts(prevBuiltImage)
	s.addServiceMountsLabels(serviceMounts, image)
	if err := s.addServiceMountsVolumes(serviceMounts, image); err != nil {
		return fmt.Errorf("error adding mounts volumes: %s", err)
	}

	customMounts := s.getCustomMounts(prevBuiltImage)
	s.addCustomMountLabels(customMounts, image)
	if err := s.addCustomMountVolumes(customMounts, image); err != nil {
		return fmt.Errorf("error adding mounts volumes: %s", err)
	}

	return nil
}

func (s *BaseStage) PreRunHook(_ Conveyor) error {
	return nil
}

func (s *BaseStage) getServiceMounts(prevBuiltImage container_runtime.ImageInterface) map[string][]string {
	return mergeMounts(s.getServiceMountsFromLabels(prevBuiltImage), s.getServiceMountsFromConfig())
}

func (s *BaseStage) getServiceMountsFromLabels(prevBuiltImage container_runtime.ImageInterface) map[string][]string {
	mountpointsByType := map[string][]string{}

	var labels map[string]string
	if prevBuiltImage != nil {
		labels = prevBuiltImage.GetStageDescription().Info.Labels
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

func (s *BaseStage) addServiceMountsVolumes(mountpointsByType map[string][]string, image container_runtime.ImageInterface) error {
	for mountType, mountpoints := range mountpointsByType {
		for _, mountpoint := range mountpoints {
			absoluteMountpoint := path.Join("/", mountpoint)

			var absoluteFrom string
			switch mountType {
			case "tmp_dir":
				absoluteFrom = filepath.Join(s.imageTmpDir, "mount", slug.Slug(absoluteMountpoint))
			case "build_dir":
				absoluteFrom = filepath.Join(werf.GetSharedContextDir(), "mounts", "projects", s.projectName, slug.Slug(absoluteMountpoint))
			default:
				panic(fmt.Sprintf("unknown mount type %s", mountType))
			}

			err := os.MkdirAll(absoluteFrom, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating tmp path %s for mount: %s", absoluteFrom, err)
			}

			image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s", absoluteFrom, absoluteMountpoint))
		}
	}

	return nil
}

func (s *BaseStage) addServiceMountsLabels(mountpointsByType map[string][]string, image container_runtime.ImageInterface) {
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

		image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{labelName: labelValue})
	}
}

func (s *BaseStage) getCustomMounts(prevBuiltImage container_runtime.ImageInterface) map[string][]string {
	return mergeMounts(s.getCustomMountsFromLabels(prevBuiltImage), s.getCustomMountsFromConfig())
}

func (s *BaseStage) getCustomMountsFromLabels(prevBuiltImage container_runtime.ImageInterface) map[string][]string {
	mountpointsByFrom := map[string][]string{}

	var labels map[string]string
	if prevBuiltImage != nil {
		labels = prevBuiltImage.GetStageDescription().Info.Labels
	}
	for k, v := range labels {
		if !strings.HasPrefix(k, imagePkg.WerfMountCustomDirLabelPrefix) {
			continue
		}

		parts := strings.SplitN(k, imagePkg.WerfMountCustomDirLabelPrefix, 2)
		fromPath := strings.Replace(parts[1], "--", "/", -1)
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

func (s *BaseStage) addCustomMountVolumes(mountpointsByFrom map[string][]string, image container_runtime.ImageInterface) error {
	for from, mountpoints := range mountpointsByFrom {
		absoluteFrom := util.ExpandPath(from)

		exist, err := util.FileExists(absoluteFrom)
		if err != nil {
			return err
		}

		if !exist {
			err := os.MkdirAll(absoluteFrom, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating %s: %s", absoluteFrom, err)
			}
		}

		for _, mountpoint := range mountpoints {
			absoluteMountpoint := path.Join("/", mountpoint)
			image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s", absoluteFrom, absoluteMountpoint))
		}
	}

	return nil
}

func (s *BaseStage) addCustomMountLabels(mountpointsByFrom map[string][]string, image container_runtime.ImageInterface) {
	for from, mountpoints := range mountpointsByFrom {
		labelName := fmt.Sprintf("%s%s", imagePkg.WerfMountCustomDirLabelPrefix, strings.Replace(filepath.ToSlash(from), "/", "--", -1))
		labelValue := strings.Join(mountpoints, ";")
		image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{labelName: labelValue})
	}
}

func (s *BaseStage) SetSignature(signature string) {
	s.signature = signature
}

func (s *BaseStage) GetSignature() string {
	return s.signature
}

func (s *BaseStage) SetContentSignature(contentSignature string) {
	s.contentSignature = contentSignature
}

func (s *BaseStage) GetContentSignature() string {
	return s.contentSignature
}

func (s *BaseStage) SetImage(image container_runtime.ImageInterface) {
	s.image = image
}

func (s *BaseStage) GetImage() container_runtime.ImageInterface {
	return s.image
}

func (s *BaseStage) SetGitMappings(gitMappings []*GitMapping) {
	s.gitMappings = gitMappings
}

func (s *BaseStage) GetGitMappings() []*GitMapping {
	return s.gitMappings
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
