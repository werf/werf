package file_manager

func CreateWerfConfigRender(path string) (string, error) {
	newFile, err := newFileWithPath(path)
	if err != nil {
		return "", err
	}

	return newFile, nil
}
