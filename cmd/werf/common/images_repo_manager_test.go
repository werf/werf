package common

import (
	"fmt"
	"testing"
)

func TestGetImagesRepoManager(t *testing.T) {
	expectationsByRepoMode := map[string]struct {
		namelessImageRepo        string
		namelessImageRepoWithTag string
		imageRepo                string
		imageRepoWithTag         string
		isMonorep                bool
	}{
		MonorepImagesRepoMode: {
			namelessImageRepo:        "repo",
			namelessImageRepoWithTag: "repo:tag",
			imageRepo:                "repo",
			imageRepoWithTag:         fmt.Sprintf("repo:image%stag", MonorepTagPartsSeparator),
			isMonorep:                true,
		},
		MultirepImagesRepoMode: {
			namelessImageRepo:        "repo",
			namelessImageRepoWithTag: "repo:tag",
			imageRepo:                "repo/image",
			imageRepoWithTag:         "repo/image:tag",
			isMonorep:                false,
		},
	}

	for imagesRepoMode, expected := range expectationsByRepoMode {
		t.Run(imagesRepoMode, func(t *testing.T) {
			m, err := GetImagesRepoManager("repo", imagesRepoMode)
			if err != nil {
				t.Error(err)
			}

			namelessImageRepo := m.ImageRepo("")
			if expected.namelessImageRepo != namelessImageRepo {
				t.Errorf("\n[EXPECTED]: %q\n[GOT]: %q", expected.namelessImageRepo, namelessImageRepo)
			}

			namelessImageRepoWithTag := m.ImageRepoWithTag("", "tag")
			if expected.namelessImageRepoWithTag != namelessImageRepoWithTag {
				t.Errorf("\n[EXPECTED]: %q\n[GOT]: %q", expected.namelessImageRepoWithTag, namelessImageRepoWithTag)
			}

			imageRepo := m.ImageRepo("image")
			if expected.imageRepo != imageRepo {
				t.Errorf("\n[EXPECTED]: %q\n[GOT]: %q", expected.imageRepo, imageRepo)
			}

			imageRepoWithTag := m.ImageRepoWithTag("image", "tag")
			if expected.imageRepoWithTag != imageRepoWithTag {
				t.Errorf("\n[EXPECTED]: %q\n[GOT]: %q", expected.imageRepoWithTag, imageRepoWithTag)
			}

			isMonorep := m.IsMonorep()
			if expected.isMonorep != isMonorep {
				t.Errorf("\n[EXPECTED]: %v\n[GOT]: %v", expected.isMonorep, isMonorep)
			}
		})
	}
}
