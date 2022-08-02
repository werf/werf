package git_repo

type ArchiveFile struct {
	FilePath string
}

func (a *ArchiveFile) GetFilePath() string {
	return a.FilePath
}
