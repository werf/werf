package deploy

import (
	"fmt"
	"os"

	"github.com/ghodss/yaml"
)

const (
	TemplateEmptyValue = "-"
)

type ServiceValuesOptions struct {
	Fake bool
}

type GitInfoGetter interface {
	IsBranchState() bool
	GetCurrentBranchName() string
	IsTagState() bool
	GetCurrentTagName() string
	GetHeadCommit() string
}

type DimgInfoGetter interface {
	IsNameless() bool
	GetName() string
	GetImageName() string
	GetImageId() (string, error)
}

func GetServiceValues(projectName, repo, namespace, dockerTag string, localGit GitInfoGetter, images []DimgInfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
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

	dappInfo := map[string]interface{}{
		"name":       projectName,
		"repo":       repo,
		"docker_tag": dockerTag,
		"ci":         ciInfo,
	}

	res["global"] = map[string]interface{}{
		"namespace": namespace,
		"dapp":      dappInfo,
	}

	if !opts.Fake {
		ciCommitTag := os.Getenv("CI_COMMIT_TAG")
		if ciCommitTag == "" {
			ciCommitTag = os.Getenv("CI_BUILD_TAG")
		}
		ciCommitRefName := os.Getenv("CI_COMMIT_REF_NAME")
		if ciCommitRefName == "" {
			ciCommitRefName = os.Getenv("CI_BUILD_REF_NAME")
		}

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

	dimgsInfo := make(map[string]interface{})
	dappInfo["dimg"] = dimgsInfo

	for _, image := range images {
		imageData := make(map[string]interface{})

		if image.IsNameless() {
			dappInfo["is_nameless_dimg"] = true
			dappInfo["dimg"] = imageData
		} else {
			dappInfo["is_nameless_dimg"] = false
			dimgsInfo[image.GetName()] = imageData
		}

		imageData["docker_image"] = image.GetImageName()
		imageData["docker_image_id"] = TemplateEmptyValue

		if !opts.Fake {
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
	}

	if debug() {
		data, err := yaml.Marshal(res)
		fmt.Printf("GetServiceValues result (err=%s):\n%s\n", err, data)
	}

	return res, nil
}
