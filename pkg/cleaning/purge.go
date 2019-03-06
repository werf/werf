package cleaning

type PurgeOptions struct {
	CommonRepoOptions    CommonRepoOptions
	CommonProjectOptions CommonProjectOptions
}

func Purge(options PurgeOptions) error {
	options.CommonProjectOptions.CommonOptions.SkipUsedImages = false
	options.CommonProjectOptions.CommonOptions.RmiForce = true
	options.CommonProjectOptions.CommonOptions.RmForce = false

	if err := ImagesPurge(options.CommonRepoOptions); err != nil {
		return err
	}

	if err := StagesPurge(options.CommonProjectOptions); err != nil {
		return err
	}

	return nil
}
