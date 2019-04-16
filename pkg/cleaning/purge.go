package cleaning

type PurgeOptions struct {
	ImagesPurgeOptions
	StagesPurgeOptions
}

func Purge(options PurgeOptions) error {
	if err := ImagesPurge(options.ImagesPurgeOptions); err != nil {
		return err
	}

	if err := StagesPurge(options.StagesPurgeOptions); err != nil {
		return err
	}

	return nil
}
