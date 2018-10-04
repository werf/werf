package cleanup

import (
	"fmt"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/flant/dapp/pkg/lock"
)

type FlushOptions struct {
	CommonRepoOptions    CommonRepoOptions    `json:"common_repo_options"`
	CommonProjectOptions CommonProjectOptions `json:"common_project_options"`
	WithStages           bool                 `json:"with_stages"`
	WithDimgs            bool                 `json:"with_dimgs"`
	OnlyRepo             bool                 `json:"only_repo"`
}

func Flush(options FlushOptions) error {
	var err error

	if options.CommonRepoOptions.Repository != "" {
		err = lock.WithLock(options.CommonRepoOptions.Repository, lock.LockOptions{Timeout: time.Second * 600}, func() error {
			if options.WithDimgs {
				if err := repoDappDimgsFlush(options.CommonRepoOptions); err != nil {
					return err
				}
			}

			if options.WithStages {
				if err := repoDappDimgstagesFlush(options.CommonRepoOptions); err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	if !options.OnlyRepo {
		projectImagesLockName := fmt.Sprintf("%s.images", options.CommonProjectOptions.ProjectName)
		err = lock.WithLock(projectImagesLockName, lock.LockOptions{Timeout: time.Second * 600}, func() error {
			if options.WithDimgs {
				if err := dappProjectDimgsFlush(options.CommonProjectOptions); err != nil {
					return err
				}
			}

			if options.WithStages {
				if err := dappProjectDimgstagesFlush(options.CommonProjectOptions); err != nil {
					return err
				}
			}

			if err := dappProjectCleanup(options.CommonProjectOptions); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func repoDappDimgsFlush(options CommonRepoOptions) error {
	dimgImages, err := repoDimgImages(options)
	if err != nil {
		return err
	}

	err = repoImagesRemove(dimgImages, options)
	if err != nil {
		return err
	}

	return nil
}

func repoDappDimgstagesFlush(options CommonRepoOptions) error {
	dimgstageImages, err := repoDimgstageImages(options)
	if err != nil {
		return err
	}

	err = repoImagesRemove(dimgstageImages, options)
	if err != nil {
		return err
	}

	return nil
}

func dappProjectDimgsFlush(options CommonProjectOptions) error {
	filterSet := filters.NewArgs()
	filterSet.Add("label", "dapp-dimg=true")
	if err := dappImagesFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	return nil
}

func dappProjectDimgstagesFlush(options CommonProjectOptions) error {
	filterSet := filters.NewArgs()
	filterSet.Add("label", dappLabel(options))
	filterSet.Add("reference", stageCacheReference(options))
	if err := dappImagesFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	return nil
}
