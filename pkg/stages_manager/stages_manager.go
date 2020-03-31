package stages_manager

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/werf"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/storage"
	"gopkg.in/yaml.v2"
)

type StagesManager struct {
	ProjectName string

	StorageLockManager storage.LockManager
	StagesStorage      storage.StagesStorage
	StagesStorageCache storage.StagesStorageCache
}

func NewStagesManager(projectName string, storageLockManager storage.LockManager, stagesStorage storage.StagesStorage, stagesStorageCache storage.StagesStorageCache) *StagesManager {
	return &StagesManager{
		ProjectName:        projectName,
		StorageLockManager: storageLockManager,
		StagesStorage:      stagesStorage,
		StagesStorageCache: stagesStorageCache,
	}
}

func (m *StagesManager) GetAllStages() ([]*image.Info, error) {
	// TODO: optimize
	//if cacheExists, images, err := m.StagesStorageCache.GetAllStages(m.ProjectName); err != nil {
	//	return nil, err
	//} else if cacheExists {
	//	return images, nil
	//} else {
	return m.StagesStorage.GetAllStages(m.ProjectName)
	//}
}

func (m *StagesManager) DeleteStages(options storage.DeleteImageOptions, imageList ...*image.Info) error {
	for _, imgInfo := range imageList {
		if err := m.StagesStorageCache.DeleteStagesBySignature(m.ProjectName, imgInfo.Signature); err != nil {
			return fmt.Errorf("unable to delete %s %s stages storage cache record: %s", err)
		}
	}
	return m.StagesStorage.DeleteStages(options, imageList...)
}

func (m *StagesManager) FetchStage(stg stage.Interface) error {
	if freshImgInfo, err := m.StagesStorage.GetImageInfo(m.ProjectName, stg.GetImage().GetStagesStorageImageInfo().Signature, stg.GetImage().GetStagesStorageImageInfo().UniqueID); err != nil {
		return err
	} else if freshImgInfo == nil {
		// TODO: stages manager should report to conveyor that conveoyor should be reset
		return fmt.Errorf("Invalid stage %s image %q! Stage is no longer available in the %s. Remove cache directory %s and retry!", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), filepath.Join(werf.GetLocalCacheDir(), "stages_storage_v4", m.ProjectName, stg.GetSignature()))
	}

	if shouldFetch, err := m.StagesStorage.ShouldFetchImage(&container_runtime.DockerImage{Image: stg.GetImage()}); err == nil && shouldFetch {
		if err := logboek.Default.LogProcess(
			fmt.Sprintf("Fetching stage %s from stages storage", stg.LogDetailedName()),
			logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
			func() error {
				logboek.Info.LogF("Image name: %s\n", stg.GetImage().Name())
				if err := m.StagesStorage.FetchImage(&container_runtime.DockerImage{Image: stg.GetImage()}); err != nil {
					return fmt.Errorf("unable to fetch stage %s image %s from stages storage %s: %s", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), err)
				}
				return nil
			},
		); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (m *StagesManager) SelectSuitableStagesStorageImage(stg stage.Interface, imagesDescs []*image.Info) (*image.Info, error) {
	if len(imagesDescs) == 0 {
		return nil, nil
	}

	var imgInfo *image.Info
	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Selecting suitable image for stage %s by signature %s", stg.Name(), stg.GetSignature()),
		logboek.LevelLogProcessOptions{},
		func() error {
			var err error
			imgInfo, err = stg.SelectCacheImage(imagesDescs)
			return err
		},
	); err != nil {
		return nil, err
	}
	if imgInfo == nil {
		return nil, nil
	}

	imgInfoData, err := yaml.Marshal(imgInfo)
	if err != nil {
		panic(err)
	}

	_ = logboek.Debug.LogBlock("Selected cache image", logboek.LevelLogBlockOptions{Style: logboek.HighlightStyle()}, func() error {
		logboek.Debug.LogF(string(imgInfoData))
		return nil
	})

	return imgInfo, nil
}

func (m *StagesManager) AtomicGetImagesBySignatureFromStagesStorageWithCacheReset(stageName, stageSig string) ([]*image.Info, error) {
	if err := m.StorageLockManager.LockStageCache(m.ProjectName, stageSig); err != nil {
		return nil, fmt.Errorf("error locking project %s stage %s cache: %s", m.ProjectName, stageSig, err)
	}
	defer m.StorageLockManager.UnlockStageCache(m.ProjectName, stageSig)

	var originImagesDescs []*image.Info
	var err error
	if err := logboek.Default.LogProcess(
		fmt.Sprintf("Get stage %s images by signature %s from stages storage", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			originImagesDescs, err = m.StagesStorage.GetRepoImagesBySignature(m.ProjectName, stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage: %s", m.StagesStorage.String(), stageSig, err)
			}

			logboek.Debug.LogF("Images: %#v\n", originImagesDescs)

			return nil
		},
	); err != nil {
		return nil, err
	}

	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Storing stage %s images by signature %s into stages storage cache", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			if err := m.StagesStorageCache.StoreStagesBySignature(m.ProjectName, stageSig, originImagesDescs); err != nil {
				return fmt.Errorf("error storing stage %s images by signature %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	return originImagesDescs, nil
}

func (m *StagesManager) AtomicStoreStageCache(stageName, stageSig string, imagesDescs []*image.Info) error {
	if err := m.StorageLockManager.LockStageCache(m.ProjectName, stageSig); err != nil {
		return fmt.Errorf("error locking stage %s cache by signature %s: %s", stageName, stageSig, err)
	}
	defer m.StorageLockManager.UnlockStageCache(m.ProjectName, stageSig)

	return logboek.Info.LogProcess(
		fmt.Sprintf("Storing stage %s images by signature %s into stages storage cache", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			if err := m.StagesStorageCache.StoreStagesBySignature(m.ProjectName, stageSig, imagesDescs); err != nil {
				return fmt.Errorf("error storing stage %s images by signature %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		},
	)
}

func (m *StagesManager) GetImagesBySignatureFromCache(stageName, stageSig string) (bool, []*image.Info, error) {
	var cacheExists bool
	var cacheImagesDescs []*image.Info

	err := logboek.Info.LogProcess(
		fmt.Sprintf("Getting stage %s images by signature %s from stages storage cache", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			var err error
			cacheExists, cacheImagesDescs, err = m.StagesStorageCache.GetStagesBySignature(m.ProjectName, stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage cache: %s", m.ProjectName, stageSig, err)
			}

			return nil
		},
	)

	return cacheExists, cacheImagesDescs, err
}
