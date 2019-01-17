package stage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/slug"
	"github.com/flant/werf/pkg/util"
)

type StageName string

const (
	From                        StageName = "from"
	BeforeInstall               StageName = "before_install"
	ArtifactImportBeforeInstall StageName = "imports_before_install"
	GitArchive                  StageName = "git_archive"
	Install                     StageName = "install"
	ArtifactImportAfterInstall  StageName = "imports_after_install"
	BeforeSetup                 StageName = "before_setup"
	ArtifactImportBeforeSetup   StageName = "imports_before_setup"
	Setup                       StageName = "setup"
	ArtifactImportAfterSetup    StageName = "imports_after_setup"
	GitCache                    StageName = "git_cache"
	GitLatestPatch              StageName = "git_latest_patch"
	DockerInstructions          StageName = "docker_instructions"
)

const (
	mountTmpDirLabel          = "werf-mount-type-tmp-dir"
	mountBuildDIrLabel        = "werf-mount-type-build-dir"
	mountCustomDirLabelPrefix = "werf-mount-type-custom-dir-"
)

type NewBaseStageOptions struct {
	DimgName         string
	ConfigMounts     []*config.Mount
	DimgTmpDir       string
	ContainerWerfDir string
	ProjectBuildDir  string
}

func newBaseStage(name StageName, options *NewBaseStageOptions) *BaseStage {
	s := &BaseStage{}
	s.name = name
	s.dimgName = options.DimgName
	s.configMounts = options.ConfigMounts
	s.projectBuildDir = options.ProjectBuildDir
	s.dimgTmpDir = options.DimgTmpDir
	s.containerWerfDir = options.ContainerWerfDir
	return s
}

type BaseStage struct {
	name             StageName
	dimgName         string
	signature        string
	image            image.Image
	gitPaths         []*GitPath
	dimgTmpDir       string
	containerWerfDir string
	projectBuildDir  string
	configMounts     []*config.Mount
}

func (s *BaseStage) Name() StageName {
	if s.name != "" {
		return s.name
	}

	panic("name must be defined!")
}

func (s *BaseStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	panic("method must be implemented!")
}

func (s *BaseStage) IsEmpty(_ Conveyor, _ image.Image) (bool, error) {
	return false, nil
}

func (s *BaseStage) ShouldBeReset(_ image.Image) (bool, error) {
	return false, nil
}

func (s *BaseStage) PrepareImage(_ Conveyor, prevBuiltImage, image image.Image) error {
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

func (s *BaseStage) AfterImageSyncDockerStateHook(_ Conveyor) error {
	return nil
}

func (s *BaseStage) PreRunHook(_ Conveyor) error {
	return nil
}

func (s *BaseStage) getServiceMounts(prevBuiltImage image.Image) map[string][]string {
	return mergeMounts(s.getServiceMountsFromLabels(prevBuiltImage), s.getServiceMountsFromConfig())
}

func (s *BaseStage) getServiceMountsFromLabels(prevBuiltImage image.Image) map[string][]string {
	mountpointsByType := map[string][]string{}

	var labels map[string]string
	if prevBuiltImage != nil {
		labels = prevBuiltImage.Labels()
	}

	for _, labelMountType := range []struct{ Label, MountType string }{
		{mountTmpDirLabel, "tmp_dir"},
		{mountBuildDIrLabel, "build_dir"},
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

		mountpoint := filepath.Clean(mountCfg.To)
		mountpointsByType[mountCfg.Type] = append(mountpointsByType[mountCfg.Type], mountpoint)
	}

	return mountpointsByType
}

func (s *BaseStage) addServiceMountsVolumes(mountpointsByType map[string][]string, image image.Image) error {
	for mountType, mountpoints := range mountpointsByType {
		for _, mountpoint := range mountpoints {
			absoluteMountpoint := filepath.Join("/", mountpoint)

			var absoluteFrom string
			switch mountType {
			case "tmp_dir":
				absoluteFrom = filepath.Join(s.dimgTmpDir, "mount", slug.Slug(absoluteMountpoint))
			case "build_dir":
				absoluteFrom = filepath.Join(s.projectBuildDir, "mount", slug.Slug(absoluteMountpoint))
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

func (s *BaseStage) addServiceMountsLabels(mountpointsByType map[string][]string, image image.Image) {
	for mountType, mountpoints := range mountpointsByType {
		var labelName string
		switch mountType {
		case "tmp_dir":
			labelName = mountTmpDirLabel
		case "build_dir":
			labelName = mountBuildDIrLabel
		default:
			panic(fmt.Sprintf("unknown mount type %s", mountType))
		}

		labelValue := strings.Join(mountpoints, ";")

		image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{labelName: labelValue})
	}
}

func (s *BaseStage) getCustomMounts(prevBuiltImage image.Image) map[string][]string {
	return mergeMounts(s.getCustomMountsFromLabels(prevBuiltImage), s.getCustomMountsFromConfig())
}

func (s *BaseStage) getCustomMountsFromLabels(prevBuiltImage image.Image) map[string][]string {
	mountpointsByFrom := map[string][]string{}

	var labels map[string]string
	if prevBuiltImage != nil {
		labels = prevBuiltImage.Labels()
	}
	for k, v := range labels {
		if !strings.HasPrefix(k, mountCustomDirLabelPrefix) {
			continue
		}

		parts := strings.SplitN(k, mountCustomDirLabelPrefix, 2)
		from := strings.Replace(parts[1], "--", "/", -1)

		mountpoints := util.RejectEmptyStrings(util.UniqStrings(strings.Split(v, ";")))
		mountpointsByFrom[from] = mountpoints
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
		mountpoint := filepath.Clean(mountCfg.To)

		mountpointsByFrom[from] = util.UniqAppendString(mountpointsByFrom[from], mountpoint)
	}

	return mountpointsByFrom
}

func (s *BaseStage) addCustomMountVolumes(mountpointsByFrom map[string][]string, image image.Image) error {
	for from, mountpoints := range mountpointsByFrom {
		absoluteFrom := util.ExpandPath(from)

		err := os.MkdirAll(absoluteFrom, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating %s: %s", absoluteFrom, err)
		}

		for _, mountpoint := range mountpoints {
			absoluteMountpoint := filepath.Join("/", mountpoint)
			image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s", absoluteFrom, absoluteMountpoint))
		}
	}

	return nil
}

func (s *BaseStage) addCustomMountLabels(mountpointsByFrom map[string][]string, image image.Image) {
	for from, mountpoints := range mountpointsByFrom {
		labelName := fmt.Sprintf("%s%s", mountCustomDirLabelPrefix, strings.Replace(from, "/", "--", -1))
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

func (s *BaseStage) SetImage(image image.Image) {
	s.image = image
}

func (s *BaseStage) GetImage() image.Image {
	return s.image
}

func (s *BaseStage) SetGitPaths(gitPaths []*GitPath) {
	s.gitPaths = gitPaths
}

func (s *BaseStage) GetGitPaths() []*GitPath {
	return s.gitPaths
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
