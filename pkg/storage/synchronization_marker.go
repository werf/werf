package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/image"
)

const synchronizationMarkerFingerprintLabel = "werf.synchronization-config-sha256"

func synchronizationMarkerName(repoAddress, projectName string) string {
	return fmt.Sprintf("%s:werf-synchronization-config-%x", repoAddress, sha256.Sum256([]byte(projectName)))
}

func (storage *RepoStagesStorage) SynchronizationMarkerReference(_ context.Context, projectName string) string {
	return synchronizationMarkerName(storage.RepoAddress, projectName)
}

func validSynchronizationFingerprint(fingerprint string) bool {
	decoded, err := hex.DecodeString(fingerprint)
	return err == nil && len(decoded) == sha256.Size
}

func (storage *RepoStagesStorage) GetSynchronizationMarker(ctx context.Context, projectName string) (string, bool, error) {
	reference := synchronizationMarkerName(storage.RepoAddress, projectName)
	img, err := storage.DockerRegistry.TryGetRepoImage(ctx, reference)
	if err != nil {
		return "", false, fmt.Errorf("read synchronization marker: %w", err)
	}
	if img == nil {
		return "", false, nil
	}
	if isStageImageOrAlias(img) {
		return "", false, fmt.Errorf("reserved synchronization marker tag %q contains a built image or custom tag alias", reference)
	}
	fingerprint := img.Labels[synchronizationMarkerFingerprintLabel]
	if img.Labels[image.WerfLabel] != projectName || !validSynchronizationFingerprint(fingerprint) {
		return "", false, fmt.Errorf("synchronization marker %q is malformed", reference)
	}
	return fingerprint, true, nil
}

func (storage *RepoStagesStorage) PutSynchronizationMarker(ctx context.Context, projectName, fingerprint string) error {
	if !validSynchronizationFingerprint(fingerprint) {
		return fmt.Errorf("invalid synchronization configuration fingerprint")
	}
	existing, found, err := storage.GetSynchronizationMarker(ctx, projectName)
	if err != nil {
		return err
	}
	if found {
		if existing != fingerprint {
			return fmt.Errorf("synchronization configuration differs from marker %q", synchronizationMarkerName(storage.RepoAddress, projectName))
		}
		return nil
	}
	reference := synchronizationMarkerName(storage.RepoAddress, projectName)
	if err := storage.DockerRegistry.PushImage(ctx, reference, &docker_registry.PushImageOptions{Labels: map[string]string{
		image.WerfLabel:                       projectName,
		synchronizationMarkerFingerprintLabel: fingerprint,
	}}); err != nil {
		return fmt.Errorf("write synchronization marker: %w", err)
	}
	existing, found, err = storage.GetSynchronizationMarker(ctx, projectName)
	if err != nil {
		return err
	}
	if !found || existing != fingerprint {
		return fmt.Errorf("synchronization marker changed while configuring project %q", projectName)
	}
	return nil
}
