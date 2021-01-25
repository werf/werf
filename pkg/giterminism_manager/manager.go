package giterminism_manager

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager/config"
	"github.com/werf/werf/pkg/giterminism_manager/file_reader"
	"github.com/werf/werf/pkg/giterminism_manager/inspector"
)

type NewManagerOptions struct {
	LooseGiterminism bool
	Dev              bool
}

func NewManager(ctx context.Context, projectDir string, localGitRepo git_repo.Local, headCommit string, options NewManagerOptions) (Interface, error) {
	sharedOptions := &sharedOptions{
		projectDir:       projectDir,
		localGitRepo:     localGitRepo,
		headCommit:       headCommit,
		looseGiterminism: options.LooseGiterminism,
		dev:              options.Dev,
	}

	fr := file_reader.NewFileReader(sharedOptions)

	c, err := config.NewConfig(ctx, fr)
	if err != nil {
		return nil, err
	}

	fr.SetGiterminismConfig(c)

	i := inspector.NewInspector(c, sharedOptions)

	m := &Manager{
		sharedOptions: sharedOptions,
		fileReader:    fr,
		inspector:     i,
	}

	logboek.Context(ctx).Debug().LogF("-- giterminism_manager.NewManager: projectDir=%q localGitRepo.WorkTreeDir=%q\n", projectDir, localGitRepo.WorkTreeDir)

	return m, nil
}

type Manager struct {
	fileReader FileReader
	inspector  Inspector

	*sharedOptions
}

func (m Manager) FileReader() FileReader {
	return m.fileReader
}

func (m Manager) Inspector() Inspector {
	return m.inspector
}

type sharedOptions struct {
	projectDir       string
	headCommit       string
	localGitRepo     git_repo.Local
	looseGiterminism bool
	dev              bool
}

func (s *sharedOptions) ProjectDir() string {
	return s.projectDir
}

func (s *sharedOptions) HeadCommit() string {
	return s.headCommit
}

func (s *sharedOptions) LocalGitRepo() *git_repo.Local {
	return &s.localGitRepo
}

func (s *sharedOptions) LooseGiterminism() bool {
	return s.looseGiterminism
}

func (s *sharedOptions) Dev() bool {
	return s.dev
}
