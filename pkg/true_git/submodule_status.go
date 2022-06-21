package true_git

import "context"

type SubmodulesStatus []*SubmoduleStatus

type SubmoduleStatus struct {
	GitDir string `json:"-"`

	Path       string           `json:"path"`
	Commit     string           `json:"commit"`
	Submodules SubmodulesStatus `json:"submodules"`
}

func GetSubmodulesStatus(ctx context.Context, gitDir, workTreeCacheDir string) (status SubmodulesStatus, resErr error) {
	resErr = withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		var err error
		status, err = getSubmodulesStatus(ctx, gitDir, workTreeCacheDir)
		return err
	})

	return
}

func getSubmodulesStatus(ctx context.Context, gitDir, workTreeCacheDir string) (SubmodulesStatus, error) {

}
