package git_repo

var CommonGitDataManager GitDataManager

func Init(gitDataManager GitDataManager) error {
	CommonGitDataManager = gitDataManager
	return nil
}
