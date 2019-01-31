package deploy

import (
	"fmt"
	"os"

	"github.com/flant/werf/pkg/config"
	"github.com/ghodss/yaml"
)

const (
	TemplateEmptyValue = "\"-\""
)

type GitInfoGetter interface {
	IsBranchState() bool
	GetCurrentBranchName() string
	IsTagState() bool
	GetCurrentTagName() string
	GetHeadCommit() string
}

type ImageInfoGetter interface {
	IsNameless() bool
	GetName() string
	GetImageName() string
	GetImageId() (string, error)
}

type ServiceValuesOptions struct {
	ForceTag    string
	ForceBranch string
}

func GetImagesInfoGetters(configImages []*config.Image, imagesRepo, tag string, withoutRegistry bool) []ImageInfoGetter {
	var images []ImageInfoGetter

	for _, image := range configImages {
		d := &ImageInfo{Config: image, WithoutRegistry: withoutRegistry, ImagesRepo: imagesRepo, Tag: tag}
		images = append(images, d)
	}

	return images
}

func GetServiceValues(projectName, repo, namespace, dockerTag string, localGit GitInfoGetter, images []ImageInfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
	if debug() {
		fmt.Printf("GetServiceValues %s %s %s %s %#v\n", projectName, repo, namespace, dockerTag, opts)
	}

	res := make(map[string]interface{})

	ciInfo := map[string]interface{}{
		"is_tag":    false,
		"is_branch": false,
		"branch":    TemplateEmptyValue,
		"tag":       TemplateEmptyValue,
		"ref":       TemplateEmptyValue,
	}

	werfInfo := map[string]interface{}{
		"name":       projectName,
		"repo":       repo,
		"docker_tag": dockerTag,
		"ci":         ciInfo,
	}

	res["global"] = map[string]interface{}{
		"namespace": namespace,
		"werf":      werfInfo,
	}

	if opts.ForceBranch != "" {
		ciInfo["is_branch"] = true
		ciInfo["branch"] = opts.ForceBranch
	} else if opts.ForceTag != "" {
		ciInfo["is_tag"] = true
		ciInfo["tag"] = opts.ForceTag
	} else {
		ciCommitTag := os.Getenv("WERF_AUTOTAG_GIT_TAG")
		ciCommitRefName := os.Getenv("WERF_AUTOTAG_GIT_BRANCH")
		if ciCommitTag != "" {
			ciInfo["tag"] = ciCommitTag
			ciInfo["ref"] = ciCommitTag
			ciInfo["is_tag"] = true
		} else if ciCommitRefName != "" {
			ciInfo["branch"] = ciCommitRefName
			ciInfo["ref"] = ciCommitRefName
			ciInfo["is_branch"] = true
		} else if localGit != nil {
			if localGit.IsBranchState() {
				ciInfo["branch"] = localGit.GetCurrentBranchName()
				ciInfo["ref"] = localGit.GetCurrentBranchName()
				ciInfo["is_branch"] = true
			} else if localGit.IsTagState() {
				ciInfo["tag"] = localGit.GetCurrentTagName()
				ciInfo["ref"] = localGit.GetCurrentTagName()
				ciInfo["is_tag"] = true
			} else {
				ciInfo["tag"] = localGit.GetHeadCommit()
				ciInfo["ref"] = localGit.GetHeadCommit()
				ciInfo["is_tag"] = true
			}
		}
	}

	imagesInfo := make(map[string]interface{})
	werfInfo["image"] = imagesInfo

	for _, image := range images {
		imageData := make(map[string]interface{})

		if image.IsNameless() {
			werfInfo["is_nameless_image"] = true
			werfInfo["image"] = imageData
		} else {
			werfInfo["is_nameless_image"] = false
			imagesInfo[image.GetName()] = imageData
		}

		imageData["docker_image"] = image.GetImageName()
		imageData["docker_image_id"] = TemplateEmptyValue

		imageID, err := image.GetImageId()
		if err != nil {
			return nil, err
		}

		if debug() {
			fmt.Printf("GetServiceValues got image id of %s: %#v", image.GetImageName(), imageID)
		}

		var value string
		if imageID == "" {
			value = TemplateEmptyValue
		} else {
			value = imageID
		}
		imageData["docker_image_id"] = value
	}

	if debug() {
		data, err := yaml.Marshal(res)
		fmt.Printf("GetServiceValues result (err=%s):\n%s\n", err, data)
	}

	return res, nil
}
