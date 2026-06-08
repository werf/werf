package externalref

import (
	"context"
	"net/http/httptest"

	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/logging"
)

var _ = Describe("Enricher", func() {
	var ts *httptest.Server
	var enricher *Enricher
	var ctx context.Context

	BeforeEach(func() {
		handler, _ := mockResolver()
		ts = httptest.NewServer(handler)
		service := NewService(ServiceConfig{ServerURL: ts.URL})
		enricher = NewEnricher(service.Resolve)
		ctx = logging.WithLogger(context.Background())
	})

	AfterEach(func() {
		ts.Close()
	})

	Describe("Enrich", func() {
		type enrichCase struct {
			bom   *cdx.BOM
			check func(error)
		}

		DescribeTable("enriches BOM correctly",
			func(c enrichCase) {
				err := enricher.Enrich(ctx, c.bom)
				c.check(err)
			},
			Entry("all components with purl", enrichCase{
				bom: &cdx.BOM{
					Components: &[]cdx.Component{
						{Name: "lodash", Version: "4.17.21", PackageURL: "pkg:npm/lodash@4.17.21", Type: cdx.ComponentTypeLibrary},
						{Name: "express", Version: "4.18.2", PackageURL: "pkg:npm/express@4.18.2", Type: cdx.ComponentTypeLibrary},
					},
				},
				check: func(err error) {
					Expect(err).NotTo(HaveOccurred())
				},
			}),
			Entry("appends to existing external refs", enrichCase{
				bom: &cdx.BOM{
					Components: &[]cdx.Component{
						{
							Name:               "lodash",
							Version:            "4.17.21",
							PackageURL:         "pkg:npm/lodash@4.17.21",
							Type:               cdx.ComponentTypeLibrary,
							ExternalReferences: &[]cdx.ExternalReference{{URL: "https://example.com", Type: cdx.ERTypeWebsite}},
						},
					},
				},
				check: func(err error) {
					Expect(err).NotTo(HaveOccurred())
				},
			}),
			Entry("returns error on component without purl", enrichCase{
				bom: &cdx.BOM{
					Components: &[]cdx.Component{
						{Name: "no-purl-lib", Version: "1.0.0", Type: cdx.ComponentTypeLibrary},
					},
				},
				check: func(err error) {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("no purl"))
				},
			}),
			Entry("returns error on first failed resolve", enrichCase{
				bom: &cdx.BOM{
					Components: &[]cdx.Component{
						{Name: "lodash", Version: "4.17.21", PackageURL: "pkg:npm/lodash@4.17.21", Type: cdx.ComponentTypeLibrary},
						{Name: "unknown", Version: "0.0.0", PackageURL: "pkg:npm/unknown@0.0.0", Type: cdx.ComponentTypeLibrary},
					},
				},
				check: func(err error) {
					Expect(err).To(HaveOccurred())
				},
			}),
			Entry("returns no error on nil components", enrichCase{
				bom: &cdx.BOM{},
				check: func(err error) {
					Expect(err).NotTo(HaveOccurred())
				},
			}),
			Entry("returns no error on empty components", enrichCase{
				bom: &cdx.BOM{Components: &[]cdx.Component{}},
				check: func(err error) {
					Expect(err).NotTo(HaveOccurred())
				},
			}),
		)

		It("sets ExternalReferences on component and BOM level", func() {
			bom := &cdx.BOM{
				Components: &[]cdx.Component{
					{Name: "lodash", Version: "4.17.21", PackageURL: "pkg:npm/lodash@4.17.21", Type: cdx.ComponentTypeLibrary},
				},
			}

			Expect(enricher.Enrich(ctx, bom)).NotTo(HaveOccurred())
			Expect((*bom.Components)[0].ExternalReferences).NotTo(BeNil())

			refs := *(*bom.Components)[0].ExternalReferences
			Expect(refs).To(HaveLen(1))
			Expect(refs[0].URL).To(Equal("https://github.com/lodash/lodash"))
			Expect(refs[0].Type).To(Equal(cdx.ERTypeVCS))

			Expect(bom.ExternalReferences).NotTo(BeNil())
			Expect(*bom.ExternalReferences).To(HaveLen(1))
			Expect((*bom.ExternalReferences)[0].URL).To(Equal("https://github.com/lodash/lodash"))
		})

		It("deduplicates BOM-level external references", func() {
			bom := &cdx.BOM{
				Components: &[]cdx.Component{
					{Name: "lodash-a", Version: "4.17.21", PackageURL: "pkg:npm/lodash@4.17.21", Type: cdx.ComponentTypeLibrary},
					{Name: "lodash-b", Version: "4.17.21", PackageURL: "pkg:npm/lodash@4.17.21", Type: cdx.ComponentTypeLibrary},
					{Name: "express", Version: "4.18.2", PackageURL: "pkg:npm/express@4.18.2", Type: cdx.ComponentTypeLibrary},
				},
			}

			Expect(enricher.Enrich(ctx, bom)).NotTo(HaveOccurred())
			Expect(*bom.ExternalReferences).To(HaveLen(2))
		})

		It("preserves existing ExternalReferences on component", func() {
			bom := &cdx.BOM{
				Components: &[]cdx.Component{
					{
						Name:               "lodash",
						Version:            "4.17.21",
						PackageURL:         "pkg:npm/lodash@4.17.21",
						Type:               cdx.ComponentTypeLibrary,
						ExternalReferences: &[]cdx.ExternalReference{{URL: "https://example.com", Type: cdx.ERTypeWebsite}},
					},
				},
			}

			Expect(enricher.Enrich(ctx, bom)).NotTo(HaveOccurred())

			refs := *(*bom.Components)[0].ExternalReferences
			Expect(refs).To(HaveLen(2))
			Expect(refs[0].URL).To(Equal("https://example.com"))
			Expect(refs[1].URL).To(Equal("https://github.com/lodash/lodash"))
		})
	})

	Describe("ExternalRefPatcher", func() {
		It("calls Enrich via Apply", func() {
			bom := &cdx.BOM{
				Components: &[]cdx.Component{
					{Name: "lodash", Version: "4.17.21", PackageURL: "pkg:npm/lodash@4.17.21", Type: cdx.ComponentTypeLibrary},
				},
			}

			patcher := &ExternalRefPatcher{enricher: enricher}
			result, err := patcher.Apply(ctx, bom)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(bom))
			Expect(result.ExternalReferences).NotTo(BeNil())
		})

		It("returns original BOM on error", func() {
			bom := &cdx.BOM{
				Components: &[]cdx.Component{
					{Name: "unknown", Version: "0.0.0", PackageURL: "pkg:npm/unknown@0.0.0", Type: cdx.ComponentTypeLibrary},
				},
			}

			patcher := &ExternalRefPatcher{enricher: enricher}
			result, err := patcher.Apply(ctx, bom)

			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(bom))
		})
	})
})
