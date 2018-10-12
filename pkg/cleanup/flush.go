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
	Mode                 FlushModeOptions     `json:"mode"`
}

type FlushModeOptions struct {
	WithStages bool `json:"with_stages"`
	WithDimgs  bool `json:"with_dimgs"`
	OnlyRepo   bool `json:"only_repo"`
}

func Flush(options FlushOptions) error {
	if options.CommonRepoOptions.Repository != "" {
		if err := repoImagesFlush(options); err != nil {
			return err
		}
	}

	if !options.Mode.OnlyRepo {
		if err := projectImagesFlush(options); err != nil {
			return err
		}
	}

	return nil
}

func repoImagesFlush(options FlushOptions) error {
	err := lock.WithLock(options.CommonRepoOptions.Repository, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		if options.Mode.WithDimgs {
			if err := repoDimgsFlush(options.CommonRepoOptions); err != nil {
				return err
			}
		}

		if options.Mode.WithStages {
			if err := repoDimgstagesFlush(options.CommonRepoOptions); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func projectImagesFlush(options FlushOptions) error {
	projectImagesLockName := fmt.Sprintf("%s.images", options.CommonProjectOptions.ProjectName)
	err := lock.WithLock(projectImagesLockName, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		if options.Mode.WithDimgs {
			if err := projectDimgsFlush(options.CommonProjectOptions); err != nil {
				return err
			}
		}

		if options.Mode.WithStages {
			if err := projectDimgstagesFlush(options.CommonProjectOptions); err != nil {
				return err
			}
		}

		if err := projectCleanup(options.CommonProjectOptions); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func repoDimgsFlush(options CommonRepoOptions) error {
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

func repoDimgstagesFlush(options CommonRepoOptions) error {
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

func projectDimgsFlush(options CommonProjectOptions) error {
	filterSet := filters.NewArgs()
	filterSet.Add("label", "dapp-dimg=true")
	if err := dappImagesFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	return nil
}

func projectDimgstagesFlush(options CommonProjectOptions) error {
	if err := dappImagesFlushByFilterSet(projectDimgstageFilterSet(options), options.CommonOptions); err != nil {
		return err
	}

	return nil
}
