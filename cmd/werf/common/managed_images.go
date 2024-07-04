package common

import (
	"context"
	"fmt"
	"sort"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/util"
)

func GetManagedImageName(userSpecifiedImageName string) string {
	switch userSpecifiedImageName {
	case "~", storage.NamelessImageRecordTag:
		return ""
	}
	return userSpecifiedImageName
}

func GetManagedImagesNames(ctx context.Context, projectName string, stagesStorage storage.StagesStorage, werfConfig *config.WerfConfig) ([]string, error) {
	var names []string
	if publishedNames, err := stagesStorage.GetManagedImages(ctx, projectName); err != nil {
		return nil, fmt.Errorf("unable to get managed images for project %q: %w", projectName, err)
	} else {
		names = append(names, publishedNames...)
	}

	names = append(names, werfConfig.GetImageNameList(false)...)

	uniqNames := util.UniqStrings(names)
	sort.Strings(uniqNames)

	return uniqNames, nil
}
