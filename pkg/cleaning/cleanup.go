package cleaning

type CleanupOptions struct {
	ImagesCleanupOptions ImagesCleanupOptions
	StagesCleanupOptions StagesCleanupOptions
}

func Cleanup(options CleanupOptions) error {
	if err := ImagesCleanup(options.ImagesCleanupOptions); err != nil {
		return err
	}

	if err := StagesCleanup(options.StagesCleanupOptions); err != nil {
		return err
	}

	return nil
}
