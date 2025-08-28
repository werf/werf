package main

import (
	"context"
	"log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

type RepoStagesStorage struct {
	repo string
}

type LocalStagesStorage struct {
	repo             string
	containerBackend string
}

func main() {
	// it needs to prebuild base images
	// docker build -t localhost:5001/playground:latest .
	// docker push localhost:5001/playground:latest
	prevBuiltTag := "latest"
	gitArchivePath := "files.tar"
	project := "playground"
	repo := "localhost:5001/" + project
	repoStagesStorage := &RepoStagesStorage{repo: repo}

	if err := repoStagesStorage.ApplyGitArchiveLayer(context.Background(), prevBuiltTag, gitArchivePath); err != nil {
		log.Fatalf("Error adding layer to repo stages storage: %v", err)
	}

	// docker build -t playground:latest .
	localStagesStorage := &LocalStagesStorage{repo: project, containerBackend: "docker"}

	if err := localStagesStorage.ApplyGitArchiveLayer(context.Background(), prevBuiltTag, gitArchivePath); err != nil {
		log.Fatalf("Error adding layer to local stages storage: %v", err)
	}
}

func (s *RepoStagesStorage) ApplyGitArchiveLayer(ctx context.Context, tag string, gitArchivePath string) error {
	oldImage := s.repo + ":" + tag
	ref, err := name.ParseReference(oldImage)
	if err != nil {
		return err
	}

	baseImg, err := remote.Image(ref)
	if err != nil {
		return err
	}

	layer, err := tarball.LayerFromFile(gitArchivePath)
	if err != nil {
		return err
	}

	newImg, err := mutate.AppendLayers(baseImg, layer)
	if err != nil {
		return err
	}

	newImage := s.repo + ":with-files"
	newRef, err := name.ParseReference(newImage)
	if err != nil {
		return err
	}

	if err := remote.Write(newRef, newImg); err != nil {
		return err
	}
	return nil
}

func (s *LocalStagesStorage) ApplyGitArchiveLayer(ctx context.Context, tag string, gitArchivePath string) error {
	if s.containerBackend == "buildah" {
		// left current implementation as is
		return nil
	}
	oldImage := s.repo + ":" + tag
	ref, _ := name.ParseReference(oldImage)
	baseImg, err := daemon.Image(ref)
	if err != nil {
		return err
	}

	layer, err := tarball.LayerFromFile(gitArchivePath)
	if err != nil {
		return err
	}

	newImg, err := mutate.AppendLayers(baseImg, layer)
	if err != nil {
		return err
	}

	newImage := s.repo + ":with-files"
	newRef, err := name.NewTag(newImage)
	if err != nil {
		return err
	}

	if _, err := daemon.Write(newRef, newImg); err != nil {
		return err
	}
	return nil
}
