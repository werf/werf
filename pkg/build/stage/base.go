package stage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/slug"
	"github.com/flant/dapp/pkg/util"
)

type StageName string

const (
	From                        StageName = "from"
	BeforeInstall               StageName = "before_install"
	ArtifactImportBeforeInstall StageName = "before_install_artifact"
	GAArchive                   StageName = "g_a_archive"
	Install                     StageName = "install"
	ArtifactImportAfterInstall  StageName = "after_install_artifact"
	BeforeSetup                 StageName = "before_setup"
	ArtifactImportBeforeSetup   StageName = "before_setup_artifact"
	Setup                       StageName = "setup"
	ArtifactImportAfterSetup    StageName = "after_setup_artifact"
	GAPostSetupPatch            StageName = "g_a_post_setup_patch"
	GALatestPatch               StageName = "g_a_latest_patch"
	DockerInstructions          StageName = "docker_instructions"
)

type NewBaseStageOptions struct {
	DimgTmpDir       string
	ContainerDappDir string
	ProjectBuildDir  string
}

func newBaseStage(options *NewBaseStageOptions) *BaseStage {
	s := &BaseStage{}
	s.projectBuildDir = options.ProjectBuildDir
	s.dimgTmpDir = options.DimgTmpDir
	s.containerDappDir = options.ContainerDappDir
	return s
}

type BaseStage struct {
	signature        string
	image            image.Image
	gitArtifacts     []*GitArtifact
	dimgTmpDir       string
	containerDappDir string
	projectBuildDir  string
}

func (s *BaseStage) Name() StageName {
	panic("method must be implemented!")
}

func (s *BaseStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	panic("method must be implemented!")
}

func (s *BaseStage) IsEmpty(_ Conveyor, _ image.Image) (bool, error) {
	return false, nil
}

func (s *BaseStage) PrepareImage(_ Conveyor, prevBuiltImage, image image.Image) error {
	var err error

	/*
	 * NOTE: BaseStage.PrepareImage does not called in From.PrepareImage.
	 * NOTE: Take into account when adding new base PrepareImage steps.
	 */

	err = s.addServiceMounts(prevBuiltImage, image)
	if err != nil {
		return fmt.Errorf("error adding service mounts: %s", err)
	}

	err = s.addCustomMounts(prevBuiltImage, image)
	if err != nil {
		return fmt.Errorf("error adding custom mounts: %s", err)
	}

	return nil
}

func (s *BaseStage) PreRunHook(_ Conveyor) error {
	return nil
}

func (s *BaseStage) addServiceMounts(prevBuiltImage, image image.Image) error {
	mountpointsByType := s.getServiceMountsFromLabels(prevBuiltImage)

	s.addServiceMountsLabels(mountpointsByType, image)

	if err := s.addServiceMountsVolumes(mountpointsByType, image); err != nil {
		return err
	}

	return nil
}

const (
	dappMountTypeTmpDirLabel          = "dapp-mount-type-tmp-dir"
	dappMountTypeBuildDirLabel        = "dapp-mount-type-build-dir"
	dappMountTypeCustomDirLabelPrefix = "dapp-mount-type-custom-dir-"
)

func (s *BaseStage) getServiceMountsFromLabels(prevBuiltImage image.Image) map[string][]string {
	mountpointsByType := map[string][]string{}

	var labels map[string]string
	if prevBuiltImage != nil {
		labels = prevBuiltImage.Labels()
	}

	for _, labelMountType := range []struct{ Label, MountType string }{
		{dappMountTypeTmpDirLabel, "tmp_dir"},
		{dappMountTypeBuildDirLabel, "build_dir"},
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
			labelName = dappMountTypeTmpDirLabel
		case "build_dir":
			labelName = dappMountTypeBuildDirLabel
		default:
			panic(fmt.Sprintf("unknown mount type %s", mountType))
		}

		labelValue := strings.Join(mountpoints, ";")

		image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{labelName: labelValue})
	}
}

func (s *BaseStage) addCustomMounts(prevBuiltImage, image image.Image) error {
	mountpointsByFrom := s.getCustomMountsFromLabels(prevBuiltImage)

	s.addCustomMountLabels(mountpointsByFrom, image)

	if err := s.addCustomMountVolumes(mountpointsByFrom, image); err != nil {
		return err
	}

	return nil
}

func (s *BaseStage) getCustomMountsFromLabels(prevBuiltImage image.Image) map[string][]string {
	mountpointsByFrom := map[string][]string{}

	var labels map[string]string
	if prevBuiltImage != nil {
		labels = prevBuiltImage.Labels()
	}
	for k, v := range labels {
		if !strings.HasPrefix(k, dappMountTypeCustomDirLabelPrefix) {
			continue
		}

		parts := strings.SplitN(k, dappMountTypeCustomDirLabelPrefix, 2)
		from := strings.Replace(parts[1], "--", "/", -1)

		mountpoints := util.RejectEmptyStrings(util.UniqStrings(strings.Split(v, ";")))
		mountpointsByFrom[from] = mountpoints
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
		labelName := fmt.Sprintf("%s%s", dappMountTypeCustomDirLabelPrefix, strings.Replace(from, "/", "--", -1))
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

func (s *BaseStage) SetGitArtifacts(gitArtifacts []*GitArtifact) {
	s.gitArtifacts = gitArtifacts
}

func (s *BaseStage) GetGitArtifacts() []*GitArtifact {
	return s.gitArtifacts
}
