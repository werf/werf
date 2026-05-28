package manager

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
)

func storeStageDescIntoLocalManifestCache(ctx context.Context, projectName string, stageID image.StageID, stg storage.BaseStorage, stageDesc *image.StageDesc) error {
	stageImageName := fmt.Sprintf("%s:%s-%d", stg.Address(), stageID.Digest, stageID.CreationTs)

	logboek.Context(ctx).Debug().LogF("Storing image %s info into manifest cache\n", stageImageName)
	if err := image.CommonManifestCache.StoreImageInfo(ctx, stg.String(), stageDesc.Info); err != nil {
		return fmt.Errorf("error storing image %s info: %w", stageImageName, err)
	}

	return nil
}
