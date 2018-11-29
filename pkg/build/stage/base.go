package stage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/slug"
	"github.com/flant/dapp/pkg/util"
)

type StageName string

const (
	From                        StageName = "from"
	BeforeInstall               StageName = "before_install"
	ArtifactImportBeforeInstall StageName = "before_install_artifact"
	GAArchive                   StageName = "g_a_archive"
	GAPreInstallPatch           StageName = "g_a_pre_install_patch"
	Install                     StageName = "install"
	ArtifactImportAfterInstall  StageName = "after_install_artifact"
	GAPostInstallPatch          StageName = "g_a_post_install_patch"
	BeforeSetup                 StageName = "before_setup"
	ArtifactImportBeforeSetup   StageName = "before_setup_artifact"
	GAPreSetupPatch             StageName = "g_a_pre_setup_patch"
	Setup                       StageName = "setup"
	ArtifactImportAfterSetup    StageName = "after_setup_artifact"
	GAPostSetupPatch            StageName = "g_a_post_setup_patch"
	GALatestPatch               StageName = "g_a_latest_patch"
	DockerInstructions          StageName = "docker_instructions"
	GAArtifactPatch             StageName = "g_a_artifact_patch"
	BuildArtifact               StageName = "build_artifact"
)

func newBaseStage() *BaseStage {
	return &BaseStage{}
}

type BaseStage struct {
	signature    string
	image        Image
	gitArtifacts []*GitArtifact
	dimgConfig   *config.Dimg
	tmpDir       string
}

func (s *BaseStage) Name() StageName {
	panic("method must be implemented!")
}

func (s *BaseStage) GetDependencies(_ Conveyor, _ Image) (string, error) {
	panic("method must be implemented!")
}

func (s *BaseStage) IsEmpty(_ Conveyor, _ Image) (bool, error) {
	panic("method must be implemented!")
}

func (s *BaseStage) GetContext(_ Conveyor) (string, error) {
	return "", nil
}

func (s *BaseStage) GetRelatedStageName() StageName {
	return ""
}

func (s *BaseStage) PrepareImage(image Image, prevImage Image) error {
	var err error

	err = s.addServiceMounts(image, prevImage)
	if err != nil {
		return fmt.Errorf("error adding service mounts: %s", err)
	}

	err = s.addCustomMounts(image, prevImage)
	if err != nil {
		return fmt.Errorf("error adding custom mounts: %s", err)
	}

	return nil
}

func (s *BaseStage) addServiceMounts(image Image, prevImage Image) error {
	mountpointsByType := map[string][]string{}

	for _, mountCfg := range s.dimgConfig.Mount {
		mountpointsByType[mountCfg.Type] = append(mountpointsByType[mountCfg.Type], mountCfg.To)
	}

	var labels map[string]string
	if prevImage != nil {
		labels = prevImage.GetLabels()
	}

	for _, labelMountType := range []struct{ Label, MountType string }{
		struct{ Label, MountType string }{"dapp-mount-tmp-dir", "tmp_dir"},
		struct{ Label, MountType string }{"dapp-mount-build-dir", "build_dir"},
	} {
		value, hasKey := labels[labelMountType.Label]
		if !hasKey {
			continue
		}

		mountpoints := strings.Split(value, ";")
		for _, mountpoint := range mountpoints {
			if mountpoint == "" {
				continue
			}

			mountpointsByType[labelMountType.MountType] = append(mountpointsByType[labelMountType.MountType], mountpoint)
		}
	}

	for mountType, mountpoints := range mountpointsByType {
		_ = mountType // todo
		for _, mountpoint := range mountpoints {
			absolutePath := util.ExpandPath(filepath.Join("/", mountpoint))
			tmpPath := filepath.Join(s.tmpDir, slug.Slug(absolutePath))

			err := os.MkdirAll(tmpPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating tmp path %s for mount: %s", tmpPath, err)
			}

			image.AddVolume(fmt.Sprintf("%s:%s", tmpPath, absolutePath))
		}
	}

	return nil
}

func (s *BaseStage) addCustomMounts(image Image, prevImage Image) error {
	mountpointsByFrom := map[string][]string{}

	for _, mountCfg := range s.dimgConfig.Mount {
		if mountCfg.Type != "custom_dir" {
			continue
		}
		fromPath := util.ExpandPath(mountCfg.From)
		mountpointsByFrom[fromPath] = util.UniqAppendString(mountpointsByFrom[fromPath], mountCfg.To)
	}

	var labels map[string]string
	if prevImage != nil {
		labels = prevImage.GetLabels()
	}

	for k, v := range labels {
		if !strings.HasPrefix(k, "dapp-mount-custom-dir-") {
			continue
		}

		parts := strings.SplitN(k, "dapp-mount-custom-dir-", 2)
		fromPath := util.ExpandPath(strings.Replace(parts[1], "--", "/", -1))
		mountpoints := util.UniqStrings(strings.Split(v, ";"))

		mountpointsByFrom[fromPath] = mountpoints
	}

	for from, mountpoints := range mountpointsByFrom {
		err := os.MkdirAll(from, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating %s: %s", from)
		}

		for _, mountpoint := range mountpoints {
			image.AddVolume(fmt.Sprintf("%s:%s", from, mountpoint))
		}

		labelName := fmt.Sprintf("dapp-mount-custom-dir-%s", strings.Replace(from, "/", "--", -1))
		labelValue := strings.Join(mountpoints, ";")

		image.AddServiceChangeLabel(labelName, labelValue)
	}

	return nil
}

func addMountsLabels(image Image, prevImage Image) error {
	/*
	   def image_add_mounts_labels
	     [:tmp_dir, :build_dir].each do |type|
	       next if (mounts = adding_mounts_by_type(type)).empty?
	       image.add_service_change_label :"dapp-mount-#{type.to_s.tr('_', '-')}" => mounts.join(';')
	     end

	     adding_custom_dir_mounts.each do |from, to_pathes|
	       image.add_service_change_label :"dapp-mount-custom-dir-#{from.gsub('/', '--')}" => to_pathes.join(';')
	     end
	   end
	*/
	return nil
}

func (s *BaseStage) SetSignature(signature string) {
	s.signature = signature
}

func (s *BaseStage) GetSignature() string {
	return s.signature
}

func (s *BaseStage) SetImage(image Image) {
	s.image = image
}

func (s *BaseStage) GetImage() Image {
	return s.image
}

func (s *BaseStage) SetGitArtifacts(gitArtifacts []*GitArtifact) {
	s.gitArtifacts = gitArtifacts
}

func (s *BaseStage) GetGitArtifacts() []*GitArtifact {
	return s.gitArtifacts
}
