//go:build linux

package contback

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"go.podman.io/common/libimage"
	"go.podman.io/storage"
	"go.podman.io/storage/pkg/unshare"

	"github.com/werf/werf/v3/pkg/buildah"
)

const cleanupProjectEnv = "_WERF_TEST_CLEANUP_PROJECT"

type cleanupProjectRequest struct {
	ProjectName  string
	Repositories []string
}

func init() {
	requestJSON, requested := os.LookupEnv(cleanupProjectEnv)
	if !requested {
		return
	}
	if err := os.Unsetenv(cleanupProjectEnv); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var request cleanupProjectRequest
	if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	err := cleanupBuildahProjectInNamespace(ctx, request)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func cleanupBuildahProject(ctx context.Context, projectName string, repos []string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find test executable: %w", err)
	}
	request, err := json.Marshal(cleanupProjectRequest{ProjectName: projectName, Repositories: repos})
	if err != nil {
		return fmt.Errorf("encode cleanup request: %w", err)
	}
	command := exec.CommandContext(ctx, "buildah", "unshare", executable)
	if os.Geteuid() == 0 {
		command = exec.CommandContext(ctx, executable)
	}
	command.Env = append(os.Environ(), cleanupProjectEnv+"="+string(request))
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clean buildah project images: %w: %s", err, output)
	}
	return nil
}

func cleanupBuildahProjectInNamespace(ctx context.Context, request cleanupProjectRequest) error {
	if request.ProjectName == "" {
		return fmt.Errorf("empty test project")
	}
	options, err := buildah.NewNativeStoreOptions(unshare.GetRootlessUID(), buildah.DefaultStorageDriver)
	if err != nil {
		return fmt.Errorf("get test storage options: %w", err)
	}
	store, err := storage.GetStore(storage.StoreOptions(*options))
	if err != nil {
		return fmt.Errorf("open test image storage: %w", err)
	}
	imageRuntime, err := libimage.RuntimeFromStore(store, &libimage.RuntimeOptions{})
	if err != nil {
		return fmt.Errorf("open test image runtime: %w", err)
	}
	list := func() ([]string, error) {
		images, err := imageRuntime.ListImages(ctx, &libimage.ListImagesOptions{Filters: []string{"label=werf=" + request.ProjectName}})
		if err != nil {
			return nil, fmt.Errorf("list native test images: %w", err)
		}
		var selected []cleanupImage
		for _, img := range images {
			selected = append(selected, cleanupImage{ID: img.ID(), Names: img.Names()})
		}
		return projectImageReferences(selected, request.ProjectName, request.Repositories), nil
	}
	removeImage := func(ctx context.Context, ref string) error {
		_, removeErrors := imageRuntime.RemoveImages(ctx, []string{ref}, &libimage.RemoveImagesOptions{NoPrune: true})
		err := errors.Join(removeErrors...)
		onlyInUseErrors := err != nil
		for _, removeErr := range removeErrors {
			if removeErr != nil && !errors.Is(removeErr, storage.ErrImageUsedByContainer) {
				onlyInUseErrors = false
			}
		}
		if onlyInUseErrors {
			return errors.Join(errCleanupImageInUse, err)
		}
		return err
	}
	return cleanupImageReferences(ctx, "buildah", list, removeImage)
}
