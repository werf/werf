package opstats

import (
	"context"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collector cache counters", func() {
	It("keeps lookups identical to the sum of the three outcomes", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)

		for range 5 {
			CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeHit, false)
		}
		for range 3 {
			CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeMiss, true)
		}
		CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeBypass, false)

		summary := collector.CacheSummary(ctx)
		Expect(summary).To(Equal([]CacheSummary{{
			Operation: OperationRegistryTagsList,
			Hit:       5,
			Miss:      3,
			Bypass:    1,
			Shared:    3,
		}}))
		Expect(summary[0].lookups()).To(Equal(9))
	})

	It("never counts a hit as shared", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)

		CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeHit, true)

		Expect(collector.CacheSummary(ctx)).To(Equal([]CacheSummary{{
			Operation: OperationRegistryTagsList,
			Hit:       1,
		}}))
	})

	It("keeps a row with lookups of a single kind, zeros included", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)

		CountCacheLookup(ctx, OperationDockerImageList, CacheOutcomeBypass, false)

		Expect(collector.CacheSummary(ctx)).To(Equal([]CacheSummary{{
			Operation: OperationDockerImageList,
			Bypass:    1,
		}}))
	})

	It("sorts rows by lookups and records nothing for an unnamed operation", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)

		CountCacheLookup(ctx, OperationDockerImageList, CacheOutcomeHit, false)
		CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeHit, false)
		CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeMiss, false)
		CountCacheLookup(ctx, "", CacheOutcomeHit, false)

		Expect(collector.CacheSummary(ctx)).To(Equal([]CacheSummary{
			{Operation: OperationRegistryTagsList, Hit: 1, Miss: 1},
			{Operation: OperationDockerImageList, Hit: 1},
		}))
	})

	It("is a no-op without a collector in context", func() {
		CountCacheLookup(context.Background(), OperationRegistryTagsList, CacheOutcomeHit, false)
	})

	It("counts concurrent lookups exactly once each", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)

		var wg sync.WaitGroup
		for range 50 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeMiss, true)
			}()
		}
		wg.Wait()

		Expect(collector.CacheSummary(ctx)).To(Equal([]CacheSummary{{
			Operation: OperationRegistryTagsList,
			Miss:      50,
			Shared:    50,
		}}))
	})
})

var _ = Describe("Collector cache pending/commit flush", func() {
	ctx := context.Background()

	It("reports only the counters recorded since the last commit", func() {
		collector := NewCollector()
		counting := NewContext(ctx, collector)

		CountCacheLookup(counting, OperationRegistryTagsList, CacheOutcomeHit, false)
		Expect(collector.PendingCacheSummary(ctx)).To(Equal([]CacheSummary{{
			Operation: OperationRegistryTagsList,
			Hit:       1,
		}}))
		collector.CommitFlush(ctx)

		CountCacheLookup(counting, OperationRegistryTagsList, CacheOutcomeMiss, true)
		Expect(collector.PendingCacheSummary(ctx)).To(Equal([]CacheSummary{{
			Operation: OperationRegistryTagsList,
			Miss:      1,
			Shared:    1,
		}}))
		collector.CommitFlush(ctx)

		Expect(collector.PendingCacheSummary(ctx)).To(BeEmpty())
		Expect(collector.CacheSummary(ctx)).To(Equal([]CacheSummary{{
			Operation: OperationRegistryTagsList,
			Hit:       1,
			Miss:      1,
			Shared:    1,
		}}))
	})

	It("keeps the observations made while the report was written", func() {
		collector := NewCollector()
		counting := NewContext(ctx, collector)

		CountCacheLookup(counting, OperationRegistryTagsList, CacheOutcomeHit, false)
		Expect(collector.PendingCacheSummary(ctx)).To(HaveLen(1))

		// Recorded after the snapshot the report was built from, while the file was being
		// written: committing the flush must not swallow it.
		CountCacheLookup(counting, OperationRegistryTagsList, CacheOutcomeMiss, false)
		collector.CommitFlush(ctx)

		Expect(collector.PendingCacheSummary(ctx)).To(Equal([]CacheSummary{{
			Operation: OperationRegistryTagsList,
			Miss:      1,
		}}))
	})

	It("retains the pending counters when the report was not committed", func() {
		collector := NewCollector()
		counting := NewContext(ctx, collector)

		CountCacheLookup(counting, OperationRegistryTagsList, CacheOutcomeHit, false)
		Expect(collector.PendingCacheSummary(ctx)).To(HaveLen(1))

		CountCacheLookup(counting, OperationRegistryTagsList, CacheOutcomeHit, false)
		Expect(collector.PendingCacheSummary(ctx)).To(Equal([]CacheSummary{{
			Operation: OperationRegistryTagsList,
			Hit:       2,
		}}))
	})

	It("never reports negative counters across three snapshots taken while lookups keep coming", func() {
		collector := NewCollector()
		counting := NewContext(ctx, collector)

		var reported int
		for range 3 {
			CountCacheLookup(counting, OperationRegistryTagsList, CacheOutcomeHit, false)
			pending := collector.PendingCacheSummary(ctx)
			Expect(pending).To(HaveLen(1))
			Expect(pending[0].Hit).To(BeNumerically(">=", 1))
			reported += pending[0].Hit
			CountCacheLookup(counting, OperationRegistryTagsList, CacheOutcomeHit, false)
			collector.CommitFlush(ctx)
		}
		reported += collector.PendingCacheSummary(ctx)[0].Hit

		Expect(reported).To(Equal(collector.CacheSummary(ctx)[0].Hit))
	})

	It("advances nothing when the report took no cache snapshot", func() {
		collector := NewCollector()
		counting := NewContext(ctx, collector)

		CountCacheLookup(counting, OperationRegistryTagsList, CacheOutcomeHit, false)
		collector.CommitFlush(ctx)

		Expect(collector.PendingCacheSummary(ctx)).To(HaveLen(1))
	})
})
