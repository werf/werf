package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
)

func getBuilder(dimgConfig config.DimgInterface, extra *builder.Extra) builder.Builder {
	var b builder.Builder
	switch dimgConfig.(type) {
	case config.Dimg:
		d := dimgConfig.(config.Dimg)
		if d.Shell != nil {
			b = builder.NewShellBuilder(d.Shell)
		} else if d.Ansible != nil {
			b = builder.NewAnsibleBuilder(d.Ansible, extra)
		}
	case config.DimgArtifact:
		d := dimgConfig.(config.DimgArtifact)
		if d.Shell != nil {
			b = builder.NewShellBuilder(d.Shell)
		} else if d.Ansible != nil {
			b = builder.NewAnsibleBuilder(d.Ansible, extra)
		}
	}

	return b
}

func newUserStage(builder builder.Builder) *UserStage {
	s := &UserStage{}
	s.builder = builder
	s.BaseStage = newBaseStage()

	return s
}

type UserStage struct {
	*BaseStage

	builder builder.Builder
}

func (s *UserStage) GetDependencies(_ Cache) string {
	return ""
}

func (s *UserStage) GetContext(_ Cache) string {
	panic("method must be implemented!")
}

func (s *UserStage) GetStageDependenciesChecksum(name StageName) string {
	var args []string
	for _, ga := range s.gitArtifacts {
		checksum, err := ga.StageDependenciesChecksum(string(Install))
		if err != nil {
			panic(err)
		}

		args = append(args, checksum)
	}

	return util.Sha256Hash(args...)
}
