package opstats

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestZZDumpRealSummary(t *testing.T) {
	base := time.Now()
	c := NewCollector()
	ctx := NewContext(context.Background(), c)
	c.add(OperationRegistryTagsList, base, base.Add(2*time.Second))
	c.add(OperationRegistryTagsList, base.Add(time.Second), base.Add(3*time.Second))
	c.add(OperationStageBuild, base, base.Add(5*time.Second))
	c.add(OperationImagePush, base, base.Add(7*time.Second))
	c.add(Operation("registry: image mutate and push"), base, base.Add(2*time.Second))
	c.add(OperationGitClone, base, base.Add(time.Second))
	c.add(OperationStageDigestLockWait, base, base.Add(time.Second))
	c.add(OperationImageSaveLoad, base, base.Add(time.Second))
	c.add(OperationImageInspect, base, base.Add(time.Second))
	c.add(OperationStapelContainer, base, base.Add(time.Second))
	c.add(OperationGiterminismInit, base, base.Add(time.Second))
	c.add(OperationConfigRender, base, base.Add(time.Second))
	c.add(OperationContextAddFiles, base, base.Add(time.Second))
	c.add(OperationStageLockWait, base, base.Add(time.Second))
	c.add(OperationBuildahImageList, base, base.Add(time.Second))
	CountCacheLookup(ctx, OperationRegistryTagsList, CacheLayerMemory, CacheOutcomeHit, false)
	CountCacheLookup(ctx, OperationRegistryTagsList, CacheLayerMemory, CacheOutcomeMiss, true)
	CountCacheLookup(ctx, OperationRegistryTagsList, CacheLayerDisk, CacheOutcomeBypass, false)
	CountCacheLookup(ctx, OperationDockerImageList, CacheLayerMemory, CacheOutcomeHit, false)
	CountCacheLookup(ctx, Operation("registry: image try get"), CacheLayerDisk, CacheOutcomeBypass, false)
	for range 6 { CountEvent(ctx, EventStageBuilt) }
	for range 17 { CountEvent(ctx, EventStageCacheHitRepo) }
	for range 2 { CountEvent(ctx, EventStageDiscarded) }
	for range 1 { CountEvent(ctx, EventStageBroken) }
	for range 2 { CountEvent(ctx, EventConveyorRestart) }
	out := logSummaryOutput(c)
	os.WriteFile("/tmp/werf-new-summary.txt", []byte(out), 0o644)
}
