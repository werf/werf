package config

import (
	"fmt"
	"path/filepath"
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

func NewImagesToProcess(werfConfig *WerfConfig, imageNameList []string, onlyFinal, withoutImages bool) (ImagesToProcess, error) {
	if withoutImages {
		return ImagesToProcess{WithoutImages: true}, nil
	}

	allImageNames := werfConfig.GetImageNameList(false)
	includedImages := make(map[string]bool)
	excludePatterns := make([]string, 0)

	if len(imageNameList) == 0 {
		for _, name := range allImageNames {
			includedImages[name] = true
		}
	} else {
		for _, pattern := range imageNameList {
			include, exclude, isExclusion, err := parsePattern(pattern)
			if err != nil {
				return ImagesToProcess{}, err
			}
			if isExclusion {
				excludePatterns = append(excludePatterns, exclude)
			} else {
				found := false
				for _, name := range allImageNames {
					match, err := filepath.Match(include, name)
					if err != nil {
						return ImagesToProcess{}, fmt.Errorf("invalid pattern %q: %v", include, err)
					}
					if match {
						includedImages[name] = true
						found = true
					}
				}
				if !found {
					return ImagesToProcess{}, fmt.Errorf("no image matches pattern %q", include)
				}
			}
		}
	}

	finalImages := make(map[string]bool)
	for name := range includedImages {
		excluded := false
		for _, exclPattern := range excludePatterns {
			match, _ := filepath.Match(exclPattern, name)
			if match {
				excluded = true
				break
			}
		}
		if !excluded {
			finalImages[name] = true
		}
	}

	if len(includedImages) == 0 && len(excludePatterns) > 0 {
		for _, name := range allImageNames {
			excluded := false
			for _, exclPattern := range excludePatterns {
				match, _ := filepath.Match(exclPattern, name)
				if match {
					excluded = true
					break
				}
			}
			if !excluded {
				finalImages[name] = true
			}
		}
	}

	resolvedImageNames := make([]string, 0, len(finalImages))
	finalImageNameList := make([]string, 0, len(finalImages))
	for name := range finalImages {
		resolvedImageNames = append(resolvedImageNames, name)
		if image := werfConfig.GetImage(name); image != nil && image.IsFinal() {
			finalImageNameList = append(finalImageNameList, name)
		}
	}

	if onlyFinal {
		resolvedImageNames = finalImageNameList
	}

	return ImagesToProcess{
		ImageNameList:      resolvedImageNames,
		FinalImageNameList: finalImageNameList,
		WithoutImages:      len(resolvedImageNames) == 0,
	}, nil
}
