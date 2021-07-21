package docker_registry

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
)

func TestApi_ParseReferenceParts_Valid(t *testing.T) {
	for _, test := range []struct {
		reference              string
		expectedReferenceParts referenceParts
	}{
		{
			reference: "account/project",
			expectedReferenceParts: referenceParts{
				registry:   name.DefaultRegistry,
				repository: "account/project",
				tag:        name.DefaultTag,
			},
		},
		{
			reference: "repo",
			expectedReferenceParts: referenceParts{
				registry:   name.DefaultRegistry,
				repository: "library/repo",
				tag:        name.DefaultTag,
			},
		},
		{
			reference: "registry.com/repo",
			expectedReferenceParts: referenceParts{
				registry:   "registry.com",
				repository: "repo",
				tag:        name.DefaultTag,
			},
		},
		{
			reference: "registry.com/repo:tag",
			expectedReferenceParts: referenceParts{
				registry:   "registry.com",
				repository: "repo",
				tag:        "tag",
			},
		},
		{
			reference: "registry.com/repo:tag@sha256:db6697a61d5679b7ca69dbde3dad6be0d17064d5b6b0e9f7be8d456ebb337209",
			expectedReferenceParts: referenceParts{
				registry:   "registry.com",
				repository: "repo",
				tag:        "tag",
				digest:     "sha256:db6697a61d5679b7ca69dbde3dad6be0d17064d5b6b0e9f7be8d456ebb337209",
			},
		},
	} {
		t.Run(test.reference, func(t *testing.T) {
			parts, err := (&api{}).ParseReferenceParts(test.reference)
			assert.Nil(t, err)
			assert.Equal(t, test.expectedReferenceParts, parts)
		})
	}
}
