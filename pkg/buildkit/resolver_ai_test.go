package buildkit

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_ImageMetaResolver_InsecureReferenceParsing(t *testing.T) {
	secure := NewImageMetaResolver(nil, ImageMetaResolverOptions{})
	assert.Empty(t, secure.parseReferenceOptions())

	insecure := NewImageMetaResolver(nil, ImageMetaResolverOptions{InsecureRegistry: true})
	require.Len(t, insecure.parseReferenceOptions(), 1)

	ref, err := name.ParseReference("localhost:5000/project:tag", insecure.parseReferenceOptions()...)
	require.NoError(t, err)
	assert.Equal(t, "http", ref.Context().Registry.Scheme())
}
