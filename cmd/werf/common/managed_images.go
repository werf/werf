package common

import (
	"context"
	"fmt"
	"sort"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/storage"
)

func GetManagedImageName(userSpecifiedImageName string) string {
	return userSpecifiedImageName
}

func GetManagedImagesNames(ctx context.Context, projectName string, metaStorage storage.RegistryStorage, werfConfig *config.WerfConfig) ([]string, error) {
	var names []string
	if publishedNames, err := metaStorage.GetManagedImages(ctx, projectName); err != nil {
		return nil, fmt.Errorf("unable to get managed images for project %q: %w", projectName, err)
	} else {
		names = append(names, publishedNames...)
	}

	names = append(names, werfConfig.GetImageNameList(false)...)

	uniqNames := util.UniqStrings(names)
	sort.Strings(uniqNames)

	return uniqNames, nil
}
