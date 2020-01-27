package stage

import (
	"fmt"
	"sort"

	"github.com/flant/werf/pkg/stages_storage"

	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/util"
)

const GitArchiveResetCommitRegex = "(\\[werf reset\\])|(\\[reset werf\\])"

type NewGitArchiveStageOptions struct {
	ArchivesDir          string
	ScriptsDir           string
	ContainerArchivesDir string
	ContainerScriptsDir  string
}

func NewGitArchiveStage(gitArchiveStageOptions *NewGitArchiveStageOptions, baseStageOptions *NewBaseStageOptions) *GitArchiveStage {
	s := &GitArchiveStage{
		ArchivesDir:          gitArchiveStageOptions.ArchivesDir,
		ScriptsDir:           gitArchiveStageOptions.ScriptsDir,
		ContainerArchivesDir: gitArchiveStageOptions.ContainerArchivesDir,
		ContainerScriptsDir:  gitArchiveStageOptions.ContainerScriptsDir,
	}
	s.GitStage = newGitStage(GitArchive, baseStageOptions)
	return s
}

type GitArchiveStage struct {
	*GitStage

	ArchivesDir          string
	ScriptsDir           string
	ContainerArchivesDir string
	ContainerScriptsDir  string
}

func (s *GitArchiveStage) SelectCacheImage(images []*stages_storage.ImageInfo) (*stages_storage.ImageInfo, error) {
	suitableImages := []*stages_storage.ImageInfo{}

ScanImages:
	for _, img := range images {
		for _, gitMapping := range s.gitMappings {
			currentCommit, err := gitMapping.LatestCommit()
			if err != nil {
				return nil, fmt.Errorf("error getting latest commit of git mapping %s: %s")
			}

			commit := gitMapping.GetGitCommitFromImageLabels(img.Labels)
			if commit != "" {
				isOurAncestor, err := gitMapping.GitRepo().IsAncestor(commit, currentCommit)
				if err != nil {
					return nil, fmt.Errorf("error checking commits ancestry %s<-%s: %s", commit, currentCommit, err)
				}

				if !isOurAncestor {
					fmt.Printf("%s is not ancestor of %s for git repo %s: ignore image %s\n", commit, currentCommit, gitMapping.GitRepo().String(), img.ImageName)
					continue ScanImages
				}
				fmt.Printf("%s is ancestor of %s for git repo %s: image %s is suitable for git archive stage\n", commit, currentCommit, gitMapping.GitRepo().String(), img.ImageName)
			} else {
				fmt.Printf("WARNING: No git commit found in image %s, skipping\n", img.ImageName)
				continue ScanImages
			}
		}

		suitableImages = append(suitableImages, img)
	}

	return s.BaseStage.SelectCacheImage(suitableImages)
}

func (s *GitArchiveStage) GetDependencies(_ Conveyor, prevImage, prevBuiltImage image.ImageInterface) (string, error) {
	var args []string
	for _, gitMapping := range s.gitMappings {
		commit := gitMapping.GetGitCommitFromImageLabels(prevImage.Labels())
		fmt.Printf("gitMapping.GetGitCommitFromImageLabels %v -> %s\n", prevImage.Name(), commit)
		if commit == "" {
			latestCommit, err := gitMapping.LatestCommit()
			if err != nil {
				return "", err
			}
			commit = latestCommit
			fmt.Printf("gitMapping.GetGitCommitFromImageLabels take latest commit -> %s\n", commit)
		}

		// FIXME
		args = append(args, gitMapping.GetParamshash())
	}

	sort.Strings(args)

	return util.Sha256Hash(args...), nil
}

func (s *GitArchiveStage) PrepareImage(c Conveyor, prevBuiltImage, image image.ImageInterface) error {
	if err := s.GitStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	for _, gitMapping := range s.gitMappings {
		if err := gitMapping.ApplyArchiveCommand(image); err != nil {
			return err
		}
	}

	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ArchivesDir, s.ContainerArchivesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ScriptsDir, s.ContainerScriptsDir))

	return nil
}
