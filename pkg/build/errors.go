package build

import "errors"

var ErrMutableStageLocalStorage = errors.New(`local storage is not supported. Please specify a repo using the --repo flag or the WERF_REPO environment variable.

Building a stage without a repo is not supported due to the excessive overhead caused by build backend limitations.`)
