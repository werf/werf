package externalref

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/logging"
)

var _ = Describe("ExternalRef Integration", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = logging.WithLogger(context.Background())
	})

	It("should enrich BOM components in parallel and handle many requests", func() {
		var callCount int32
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			purl := r.URL.Query().Get("purl")
			res := &ResolveResult{
				PURL: purl,
				URL:  "https://github.com/example/repo",
				Kind: "vcs",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(res)
		}))
		defer ts.Close()

		svc := NewService(ServiceConfig{ServerURL: ts.URL})
		enricher := NewEnricher(svc.Resolve)

		const componentCount = 50
		components := make([]cdx.Component, componentCount)
		for i := 0; i < componentCount; i++ {
			components[i] = cdx.Component{
				Name:       fmt.Sprintf("pkg-%d", i),
				PackageURL: fmt.Sprintf("pkg:npm/pkg-%d@1.0.0", i),
			}
		}
		bom := &cdx.BOM{Components: &components}

		err := enricher.Enrich(ctx, bom)
		Expect(err).NotTo(HaveOccurred())
		Expect(atomic.LoadInt32(&callCount)).To(Equal(int32(componentCount)))

		for _, comp := range *bom.Components {
			Expect(comp.ExternalReferences).NotTo(BeNil())
			Expect(*comp.ExternalReferences).To(HaveLen(1))
		}
		Expect(bom.ExternalReferences).NotTo(BeNil())
		Expect(*bom.ExternalReferences).To(HaveLen(1))
	})

	It("should retry and eventually succeed on temporary server errors", func() {
		var callCount int32
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&callCount, 1)
			if count <= 2 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			res := &ResolveResult{
				PURL: r.URL.Query().Get("purl"),
				URL:  "https://github.com/example/repo",
				Kind: "vcs",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(res)
		}))
		defer ts.Close()

		svc := NewService(ServiceConfig{ServerURL: ts.URL})
		enricher := NewEnricher(svc.Resolve)

		bom := &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "retry-pkg", PackageURL: "pkg:npm/retry-pkg@1.0.0"},
			},
		}

		err := enricher.Enrich(ctx, bom)
		Expect(err).NotTo(HaveOccurred())
		Expect(atomic.LoadInt32(&callCount)).To(Equal(int32(3)))
	})

	It("should fail if retries are exhausted", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer ts.Close()

		timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		svc := NewService(ServiceConfig{ServerURL: ts.URL})
		enricher := NewEnricher(svc.Resolve)

		bom := &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "fail-pkg", PackageURL: "pkg:npm/fail-pkg@1.0.0"},
			},
		}

		err := enricher.Enrich(timeoutCtx, bom)
		Expect(err).To(HaveOccurred())
	})

	It("should respect context timeout during retries", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer ts.Close()

		svc := NewService(ServiceConfig{ServerURL: ts.URL})
		enricher := NewEnricher(svc.Resolve)

		timeoutCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		bom := &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "timeout-pkg", PackageURL: "pkg:npm/timeout-pkg@1.0.0"},
			},
		}

		start := time.Now()
		err := enricher.Enrich(timeoutCtx, bom)
		Expect(err).To(HaveOccurred())
		Expect(time.Since(start)).To(BeNumerically("<", 2*time.Second))
	})
})
