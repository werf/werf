package build

import (
	"fmt"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/storage"
	"gopkg.in/yaml.v3"
)

func fetchStage(stagesStorage storage.StagesStorage, stg stage.Interface) error {
	if shouldFetch, err := stagesStorage.ShouldFetchImage(&container_runtime.DockerImage{Image: stg.GetImage()}); err == nil && shouldFetch {
		if err := logboek.Default.LogProcess(
			fmt.Sprintf("Fetching stage %s from stages storage", stg.LogDetailedName()),
			logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
			func() error {
				logboek.Info.LogF("Image name: %s\n", stg.GetImage().Name())
				if err := stagesStorage.FetchImage(&container_runtime.DockerImage{Image: stg.GetImage()}); err != nil {
					return fmt.Errorf("unable to fetch stage %s image %s from stages storage %s: %s", stg.LogDetailedName(), stg.GetImage().Name(), stagesStorage.String(), err)
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

func selectSuitableStagesStorageImage(stg stage.Interface, imagesDescs []*image.Info) (*image.Info, error) {
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

func atomicGetImagesBySignatureFromStagesStorageWithCacheReset(conveyor *Conveyor, stageName, stageSig string) ([]*image.Info, error) {
	if err := conveyor.StorageLockManager.LockStageCache(conveyor.projectName(), stageSig); err != nil {
		return nil, fmt.Errorf("error locking project %s stage %s cache: %s", conveyor.projectName(), stageSig, err)
	}
	defer conveyor.StorageLockManager.UnlockStageCache(conveyor.projectName(), stageSig)

	var originImagesDescs []*image.Info
	var err error
	if err := logboek.Default.LogProcess(
		fmt.Sprintf("Get stage %s images by signature %s from stages storage", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			originImagesDescs, err = conveyor.StagesStorage.GetRepoImagesBySignature(conveyor.projectName(), stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage: %s", conveyor.StagesStorage.String(), stageSig, err)
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
			if err := conveyor.StagesStorageCache.StoreImagesBySignature(conveyor.projectName(), stageSig, originImagesDescs); err != nil {
				return fmt.Errorf("error storing stage %s images by signature %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	return originImagesDescs, nil
}

func atomicStoreStageCache(conveyor *Conveyor, stageName, stageSig string, imagesDescs []*image.Info) error {
	if err := conveyor.StorageLockManager.LockStageCache(conveyor.projectName(), stageSig); err != nil {
		return fmt.Errorf("error locking stage %s cache by signature %s: %s", stageName, stageSig, err)
	}
	defer conveyor.StorageLockManager.UnlockStageCache(conveyor.projectName(), stageSig)

	return logboek.Info.LogProcess(
		fmt.Sprintf("Storing stage %s images by signature %s into stages storage cache", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			if err := conveyor.StagesStorageCache.StoreImagesBySignature(conveyor.projectName(), stageSig, imagesDescs); err != nil {
				return fmt.Errorf("error storing stage %s images by signature %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		},
	)
}

func getImagesBySignatureFromCache(conveyor *Conveyor, stageName, stageSig string) (bool, []*image.Info, error) {
	var cacheExists bool
	var cacheImagesDescs []*image.Info

	err := logboek.Info.LogProcess(
		fmt.Sprintf("Getting stage %s images by signature %s from stages storage cache", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			var err error
			cacheExists, cacheImagesDescs, err = conveyor.StagesStorageCache.GetImagesBySignature(conveyor.projectName(), stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage cache: %s", conveyor.projectName(), stageSig, err)
			}

			return nil
		},
	)

	return cacheExists, cacheImagesDescs, err
}
