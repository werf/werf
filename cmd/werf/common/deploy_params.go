package common

import (
	"fmt"
	"strings"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/image"
)

func GetUserExtraAnnotations(cmdData *CmdData) (map[string]string, error) {
	extraAnnotationMap := map[string]string{}
	var addAnnotations []string

	addAnnotations = append(addAnnotations, GetAddAnnotations(cmdData)...)

	for _, addAnnotation := range addAnnotations {
		parts := strings.Split(addAnnotation, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad --add-annotation value %s", addAnnotation)
		}

		extraAnnotationMap[parts[0]] = parts[1]
	}

	return extraAnnotationMap, nil
}

func GetUserExtraLabels(cmdData *CmdData) (map[string]string, error) {
	addLabelArray := append([]string{}, GetAddLabels(cmdData)...)
	addLabelMap, err := KeyValueArrayToMap(addLabelArray, "=")
	if err != nil {
		return nil, fmt.Errorf("unsupported --add-label value: %w", err)
	}

	return addLabelMap, nil
}

func KeyValueArrayToMap(pairs []string, sep string) (map[string]string, error) {
	keyValueMap := map[string]string{}
	for _, pair := range pairs {
		parts := strings.SplitN(pair, sep, 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key=value pair %q", pair)
		}

		keyValueMap[parts[0]] = parts[1]
	}

	return keyValueMap, nil
}

func StubImageInfoGetters(werfConfig *config.WerfConfig) (list []*image.InfoGetter) {
	var imagesNames []string
	for _, imageConfig := range werfConfig.StapelImages {
		imagesNames = append(imagesNames, imageConfig.Name)
	}
	for _, imageConfig := range werfConfig.ImagesFromDockerfile {
		imagesNames = append(imagesNames, imageConfig.Name)
	}

	for _, imageName := range imagesNames {
		list = append(list, image.NewInfoGetter(imageName, fmt.Sprintf("%s:%s", StubRepoAddress, StubTag), image.InfoGetterOptions{}))
	}

	return list
}
