package file_reader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/giterminism"
	"github.com/werf/werf/pkg/util"
)

var DefaultWerfConfigTemplatesDirName = ".werf"

func (r FileReader) ReadConfigTemplateFiles(ctx context.Context, customDirRelPath string, tmplFunc func(templatePathInsideDir string, data []byte, err error) error) error {
	processedFiles := map[string]bool{}

	templatesDirRelPath := DefaultWerfConfigTemplatesDirName
	if customDirRelPath != "" {
		templatesDirRelPath = customDirRelPath
	}

	isFileProcessedFunc := func(relPath string) bool {
		templateRelPath := util.GetRelativeToBaseFilepath(templatesDirRelPath, relPath)
		return processedFiles[templateRelPath]
	}

	readFileBeforeHookFunc := func(relPath string) {
		templateRelPath := util.GetRelativeToBaseFilepath(templatesDirRelPath, relPath)
		processedFiles[templateRelPath] = true
	}

	readFileFunc := func(relPath string) ([]byte, error) {
		readFileBeforeHookFunc(relPath)
		return r.readFile(relPath)
	}

	readCommitFileFunc := func(relPath string) ([]byte, error) {
		readFileBeforeHookFunc(relPath)
		return r.readCommitConfigTemplateFile(ctx, relPath)
	}

	templateRelPathListFromFS, err := r.configTemplateFileRelPathListFromFS(templatesDirRelPath)
	if err != nil {
		return err
	}

	if r.manager.LooseGiterminism() {
		for _, templateRelPath := range templateRelPathListFromFS {
			data, err := readFileFunc(templateRelPath)
			if err := tmplFunc(templateRelPath, data, err); err != nil {
				return err
			}
		}

		return nil
	}

	templateRelPathListFromCommit, err := r.configTemplateFilePathListFromCommit(ctx, templatesDirRelPath)
	if err != nil {
		return err
	}

	for _, templateRelPath := range templateRelPathListFromCommit {
		if accepted, err := r.manager.Config().IsUncommittedConfigTemplateFileAccepted(templateRelPath); err != nil {
			return err
		} else if accepted {
			continue
		}

		data, err := readCommitFileFunc(templateRelPath)
		if err := tmplFunc(templateRelPath, data, err); err != nil {
			return err
		}
	}

	for _, templateRelPath := range templateRelPathListFromFS {
		accepted, err := r.manager.Config().IsUncommittedConfigTemplateFileAccepted(templateRelPath)
		if err != nil {
			return err
		}

		if !accepted {
			if !isFileProcessedFunc(templateRelPath) {
				return giterminism.NewUncommittedConfigurationError(fmt.Sprintf("the werf config template '%s' must be committed", templateRelPath))
			}

			continue
		}

		data, err := readFileFunc(templateRelPath)
		if err := tmplFunc(templateRelPath, data, err); err != nil {
			return err
		}
	}

	return nil
}

func (r FileReader) configTemplateFileRelPathListFromFS(templatesDirRelPath string) ([]string, error) {
	absTemplatesDir := filepath.Join(r.manager.ProjectDir(), templatesDirRelPath)
	if exist, err := util.DirExists(absTemplatesDir); err != nil {
		return nil, fmt.Errorf("unable to check existence of directory %s: %s", absTemplatesDir, err)
	} else if !exist {
		return nil, nil
	}

	var templateRelPathList []string
	if err := filepath.Walk(absTemplatesDir, func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		if matched, err := isTemplateFile(fi.Name()); err != nil {
			return err
		} else if matched {
			templatePathInsideProjectDir := util.GetRelativeToBaseFilepath(r.manager.ProjectDir(), fp)
			templateRelPathList = append(templateRelPathList, templatePathInsideProjectDir)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return templateRelPathList, nil
}

func (r FileReader) configTemplateFilePathListFromCommit(ctx context.Context, templatesDirRelPath string) ([]string, error) {
	paths, err := r.manager.LocalGitRepo().GetCommitFilePathList(ctx, r.manager.HeadCommit())
	if err != nil {
		return nil, fmt.Errorf("unable to get file list from the project git repo commit %s: %s", r.manager.HeadCommit(), err)
	}

	var templatesPathList []string
	for _, relPath := range paths {
		if !util.IsSubpathOfBasePath(templatesDirRelPath, relPath) {
			continue
		}

		if matched, err := isTemplateFile(relPath); err != nil {
			return nil, err
		} else if matched {
			templatesPathList = append(templatesPathList, relPath)
		}

	}

	return templatesPathList, nil
}

func (r FileReader) readCommitConfigTemplateFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.readCommitFile(ctx, relPath, func(ctx context.Context, relPath string) error {
		return giterminism.NewUncommittedConfigurationError(fmt.Sprintf("the werf config template '%s' must be committed", relPath))
	})
}

func isTemplateFile(path string) (bool, error) {
	return filepath.Match("*.tmpl", filepath.Base(path))
}
