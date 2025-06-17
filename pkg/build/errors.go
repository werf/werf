package build

import "errors"

var (
	ErrMutableStageLocalStorage = errors.New(`local storage is not supported. Please specify a repo using the --repo flag or the WERF_REPO environment variable.

Building a stage without a repo is not supported due to the excessive overhead caused by build backend limitations.`)

	ErrMutableStageLocalStorageImageSpec = errors.New(ErrMutableStageLocalStorage.Error() + `

To debug the build locally, consider running a local registry or skipping the imageSpec stage using the option --skip-image-spec-stage.`)
)
