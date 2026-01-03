package image

import (
	"fmt"
	"strings"
	"time"
)

type Summary struct {
	ID          string
	RepoTags    []string
	RepoDigests []string
	Labels      map[string]string
	Created     time.Time
	Size        int64
}

type ImagesList []Summary

// FIXME(multiarch): take into account multiarch stages, which does not use creationTs
func (list ImagesList) ConvertToStages() ([]StageID, error) {
	var stagesList []StageID

	for _, summary := range list {
		repoTags := summary.RepoTags
		if len(repoTags) == 0 {
			repoTags = append(repoTags, "<none>:<none>")
		}

		for _, repoTag := range repoTags {
			_, tag := ParseRepositoryAndTag(repoTag)

			if len(tag) != 70 || len(strings.Split(tag, "-")) != 2 { // 2604b86b2c7a1c6d19c62601aadb19e7d5c6bb8f17bc2bf26a390ea7-1611836746968
				continue
			}

			if digest, creationTs, err := GetDigestAndCreationTsFromLocalStageImageTag(tag); err != nil {
				return nil, err
			} else {
				stagesList = append(stagesList, *NewStageID(digest, creationTs))
			}
		}
	}

	return stagesList, nil
}

func GetDigestAndCreationTsFromLocalStageImageTag(repoStageImageTag string) (string, int64, error) {
	parts := strings.SplitN(repoStageImageTag, "-", 2)
	if creationTs, err := ParseCreationTs(parts[1]); err != nil {
		return "", 0, fmt.Errorf("unable to parse value %q as creation timestamp: %w", parts[1], err)
	} else {
		return parts[0], creationTs, nil
	}
}
