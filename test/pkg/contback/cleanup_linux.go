//go:build linux

package contback

import (
	"bytes"
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

const (
	cleanupProjectEnv   = "_WERF_TEST_CLEANUP_PROJECT"
	storageSnapshotEnv  = "_WERF_TEST_BUILDAH_STORAGE_SNAPSHOT"
	namespaceJobTimeout = 2 * time.Minute
)

type cleanupProjectRequest struct {
	ProjectName  string
	Repositories []string
}

func init() {
	if requestJSON, requested := os.LookupEnv(cleanupProjectEnv); requested {
		runNamespaceJob(cleanupProjectEnv, func(ctx context.Context) error {
			var request cleanupProjectRequest
			if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
				return err
			}
			return cleanupBuildahProjectInNamespace(ctx, request)
		})
	}
	if projectName, requested := os.LookupEnv(storageSnapshotEnv); requested {
		runNamespaceJob(storageSnapshotEnv, func(ctx context.Context) error {
			snapshot, err := snapshotBuildahStorageInNamespace(ctx, projectName)
			if err != nil {
				return err
			}
			return json.NewEncoder(os.Stdout).Encode(snapshot)
		})
	}
}

// runNamespaceJob is the child side of runInBuildahNamespace: the test binary was re-executed
// inside the buildah user namespace only to run job, so it exits right after it.
func runNamespaceJob(env string, job func(ctx context.Context) error) {
	if err := os.Unsetenv(env); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), namespaceJobTimeout)
	err := job(ctx)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

// runInBuildahNamespace re-executes the test binary inside the buildah user namespace (the
// rootless containers storage is only accessible from there) with env=value set, which the
// child's init recognizes, and returns the child's stdout.
func runInBuildahNamespace(ctx context.Context, env, value string) ([]byte, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("find test executable: %w", err)
	}
	command := exec.CommandContext(ctx, "buildah", "unshare", executable)
	if os.Geteuid() == 0 {
		command = exec.CommandContext(ctx, executable)
	}
	command.Env = append(os.Environ(), env+"="+value)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout, nil
}

func cleanupBuildahProject(ctx context.Context, projectName string, repos []string) error {
	request, err := json.Marshal(cleanupProjectRequest{ProjectName: projectName, Repositories: repos})
	if err != nil {
		return fmt.Errorf("encode cleanup request: %w", err)
	}
	if _, err := runInBuildahNamespace(ctx, cleanupProjectEnv, string(request)); err != nil {
		return fmt.Errorf("clean buildah project images: %w", err)
	}
	return nil
}

func snapshotBuildahStorage(ctx context.Context, projectName string) (BuildahStorageSnapshot, error) {
	output, err := runInBuildahNamespace(ctx, storageSnapshotEnv, projectName)
	if err != nil {
		return BuildahStorageSnapshot{}, fmt.Errorf("snapshot buildah storage: %w", err)
	}
	var snapshot BuildahStorageSnapshot
	if err := json.Unmarshal(output, &snapshot); err != nil {
		return BuildahStorageSnapshot{}, fmt.Errorf("decode buildah storage snapshot: %w", err)
	}
	return snapshot, nil
}

func openBuildahImageRuntime() (storage.Store, *libimage.Runtime, error) {
	options, err := buildah.NewNativeStoreOptions(unshare.GetRootlessUID(), buildah.DefaultStorageDriver)
	if err != nil {
		return nil, nil, fmt.Errorf("get test storage options: %w", err)
	}
	store, err := storage.GetStore(storage.StoreOptions(*options))
	if err != nil {
		return nil, nil, fmt.Errorf("open test image storage: %w", err)
	}
	imageRuntime, err := libimage.RuntimeFromStore(store, &libimage.RuntimeOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("open test image runtime: %w", err)
	}
	return store, imageRuntime, nil
}

func snapshotBuildahStorageInNamespace(ctx context.Context, projectName string) (BuildahStorageSnapshot, error) {
	store, imageRuntime, err := openBuildahImageRuntime()
	if err != nil {
		return BuildahStorageSnapshot{}, err
	}
	var snapshot BuildahStorageSnapshot
	images, err := imageRuntime.ListImages(ctx, &libimage.ListImagesOptions{Filters: []string{"label=werf=" + projectName}})
	if err != nil {
		return BuildahStorageSnapshot{}, fmt.Errorf("list native test images: %w", err)
	}
	for _, img := range images {
		snapshot.ProjectImageIDs = append(snapshot.ProjectImageIDs, img.ID())
	}
	containers, err := store.Containers()
	if err != nil {
		return BuildahStorageSnapshot{}, fmt.Errorf("list native containers: %w", err)
	}
	for _, container := range containers {
		snapshot.Containers = append(snapshot.Containers, BuildahContainer{ID: container.ID, ImageID: container.ImageID, Names: container.Names})
	}
	return snapshot, nil
}

func cleanupBuildahProjectInNamespace(ctx context.Context, request cleanupProjectRequest) error {
	if request.ProjectName == "" {
		return fmt.Errorf("empty test project")
	}
	_, imageRuntime, err := openBuildahImageRuntime()
	if err != nil {
		return err
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
