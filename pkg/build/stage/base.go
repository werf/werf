package stage

import "github.com/flant/dapp/pkg/build"

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
	gitArtifacts []*build.GitArtifact
}

func (s *BaseStage) Name() StageName {
	panic("method must be implemented!")
}

func (s *BaseStage) GetDependencies(_ Cache) string {
	panic("method must be implemented!")
}

func (s *BaseStage) GetContext(_ Cache) string {
	return ""
}

func (s *BaseStage) GetRelatedStageName() StageName {
	return ""
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

func (s *BaseStage) SetGitArtifacts(gitArtifacts []*build.GitArtifact) {
	s.gitArtifacts = gitArtifacts
}

func (s *BaseStage) GetGitArtifacts() []*build.GitArtifact {
	return s.gitArtifacts
}
