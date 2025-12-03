package config

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type ImagesToProcess struct {
	ImageNameList      []string
	FinalImageNameList []string
	WithoutImages      bool
}

func parsePattern(pattern string) (include, exclude string, isExclusion bool, err error) {
	if strings.Contains(pattern, "**") {
		return "", "", false, fmt.Errorf("recursive glob (**) not supported")
	}
	if len(pattern) > 100 {
		return "", "", false, fmt.Errorf("pattern too long")
	}

	if strings.ContainsAny(pattern, "&|$") {
		return "", "", false, fmt.Errorf("special characters not allowed")
	}

	if strings.HasPrefix(pattern, "!") {
		return "", strings.TrimPrefix(pattern, "!"), true, nil
	}
	return pattern, "", false, nil
}

func parsePatterns(imageNameList []string) (includePatterns, excludePatterns []string, err error) {
	includePatterns = make([]string, 0, len(imageNameList))
	excludePatterns = make([]string, 0, len(imageNameList))

	for _, pattern := range imageNameList {
		include, exclude, isExclusion, err := parsePattern(pattern)
		if err != nil {
			return nil, nil, err
		}
		if isExclusion {
			excludePatterns = append(excludePatterns, exclude)
		} else {
			includePatterns = append(includePatterns, include)
		}
	}
	return includePatterns, excludePatterns, nil
}

func matchPattern(name, pattern string) (bool, error) {
	match, err := filepath.Match(pattern, name)
	if err != nil {
		return false, fmt.Errorf("invalid pattern %q: %v", pattern, err)
	}
	return match, nil
}

func filterUsingIncludePatterns(allImageNames, includePatterns []string) (map[string]bool, error) {
	finalImages := make(map[string]bool)
	if len(includePatterns) == 0 {
		for _, name := range allImageNames {
			finalImages[name] = true
		}
		return finalImages, nil
	}

	for _, pattern := range includePatterns {
		found := false
		for _, name := range allImageNames {
			match, err := matchPattern(name, pattern)
			if err != nil {
				return nil, err
			}
			if match {
				finalImages[name] = true
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("no image matches pattern %q", pattern)
		}
	}
	return finalImages, nil
}

func filterUsingExcludePatterns(finalImages map[string]bool, excludePatterns, includePatterns []string) (map[string]bool, error) {
	if len(excludePatterns) == 0 {
		return finalImages, nil
	}

	temp := make(map[string]bool)
	for name := range finalImages {
		shouldExclude := false
		for _, excludePattern := range excludePatterns {
			match, err := matchPattern(name, excludePattern)
			if err != nil {
				return nil, err
			}
			if match && !slices.Contains(includePatterns, name) {
				shouldExclude = true
				break
			}
		}
		if !shouldExclude {
			temp[name] = true
		}
	}
	return temp, nil
}

func filterUsingFinalAttribute(werfConfig *WerfConfig, imageNames []string) []string {
	finalImageNameList := make([]string, 0, len(imageNames))
	for _, name := range imageNames {
		if image := werfConfig.GetImage(name); image != nil && image.IsFinal() {
			finalImageNameList = append(finalImageNameList, name)
		}
	}
	return finalImageNameList
}

func NewImagesToProcess(werfConfig *WerfConfig, imageNameList []string, onlyFinal, withoutImages bool) (ImagesToProcess, error) {
	if withoutImages {
		return ImagesToProcess{WithoutImages: true}, nil
	}

	allImageNames := werfConfig.GetImageNameList(false)
	includePatterns, excludePatterns, err := parsePatterns(imageNameList)
	if err != nil {
		return ImagesToProcess{}, err
	}

	finalImages, err := filterUsingIncludePatterns(allImageNames, includePatterns)
	if err != nil {
		return ImagesToProcess{}, err
	}

	finalImages, err = filterUsingExcludePatterns(finalImages, excludePatterns, includePatterns)
	if err != nil {
		return ImagesToProcess{}, err
	}

	resolvedImageNames := make([]string, 0, len(finalImages))
	for name := range finalImages {
		resolvedImageNames = append(resolvedImageNames, name)
	}

	finalImageNameList := filterUsingFinalAttribute(werfConfig, resolvedImageNames)

	if onlyFinal {
		resolvedImageNames = finalImageNameList
	}

	return ImagesToProcess{
		ImageNameList:      resolvedImageNames,
		FinalImageNameList: finalImageNameList,
		WithoutImages:      len(resolvedImageNames) == 0,
	}, nil
}
