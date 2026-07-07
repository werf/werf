package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/storage"
)

func TestAI_GetOptionalImagesStorageDefaultLocal(t *testing.T) {
	cmdData := cmdDataFor("", nil, nil, "", "", "", "")
	falseValue := false
	cmdData.InsecureRegistry = &falseValue
	cmdData.SkipTlsVerifyRegistry = &falseValue

	imagesStorage, err := GetOptionalImagesStorage(context.Background(), nil, cmdData)

	require.NoError(t, err)
	require.NotNil(t, imagesStorage)
	assert.Equal(t, storage.LocalStorageAddress, imagesStorage.Address())
}

func TestAI_GetOptionalImagesStorageExplicitAddress(t *testing.T) {
	cmdData := cmdDataFor("", nil, nil, "registry.example.com/project/images", "", "", "")
	falseValue := false
	cmdData.InsecureRegistry = &falseValue
	cmdData.SkipTlsVerifyRegistry = &falseValue

	imagesStorage, err := GetOptionalImagesStorage(context.Background(), nil, cmdData)

	require.NoError(t, err)
	require.NotNil(t, imagesStorage)
	assert.Equal(t, "registry.example.com/project/images", imagesStorage.Address())
}
