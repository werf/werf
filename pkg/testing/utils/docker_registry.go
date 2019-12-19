package utils

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/flant/go-containerregistry/pkg/authn"
	"github.com/flant/go-containerregistry/pkg/name"
	"github.com/flant/go-containerregistry/pkg/v1/remote"
)

func RegistryRepositoryList(reference string) []string {
	repo, err := name.NewRepository(reference, name.WeakValidation)
	Ω(err).ShouldNot(HaveOccurred(), fmt.Sprintf("parsing repo %q: %v", reference, err))

	tags, err := remote.List(repo, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil && strings.Contains(err.Error(), "NAME_UNKNOWN") {
		return []string{}
	}

	Ω(err).ShouldNot(HaveOccurred(), fmt.Sprintf("reading tags for %q: %v", repo, err))
	return tags
}
