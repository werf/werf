package externalref

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = Describe("Service", func() {
	var ts *httptest.Server
	var calls *int
	var service *Service
	var ctx context.Context

	BeforeEach(func() {
		handler, c := mockResolver()
		calls = c
		ts = httptest.NewServer(handler)
		service = NewService(ServiceConfig{ServerURL: ts.URL})
		ctx = logging.WithLogger(context.Background())
	})

	AfterEach(func() {
		ts.Close()
	})

	Describe("Resolve", func() {
		type resolveCase struct {
			purl  string
			check func(*ResolveResult, error)
		}

		DescribeTable("returns expected result",
			func(c resolveCase) {
				result, err := service.Resolve(ctx, c.purl)
				c.check(result, err)
			},
			Entry("resolves lodash to VCS URL", resolveCase{
				purl: "pkg:npm/lodash@4.17.21",
				check: func(r *ResolveResult, err error) {
					Expect(err).NotTo(HaveOccurred())
					Expect(r.URL).To(Equal("https://github.com/lodash/lodash"))
					Expect(r.Kind).To(Equal("vcs"))
					Expect(r.Confirmed).To(BeTrue())
					Expect(r.PURL).To(Equal("pkg:npm/lodash@4.17.21"))
				},
			}),
			Entry("resolves express to VCS URL", resolveCase{
				purl: "pkg:npm/express@4.18.2",
				check: func(r *ResolveResult, err error) {
					Expect(err).NotTo(HaveOccurred())
					Expect(r.URL).To(Equal("https://github.com/expressjs/express"))
					Expect(r.Kind).To(Equal("vcs"))
					Expect(r.Confirmed).To(BeTrue())
				},
			}),
			Entry("returns error on 404", resolveCase{
				purl: "pkg:npm/unknown@0.0.0",
				check: func(r *ResolveResult, err error) {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unexpected status 404"))
				},
			}),
			Entry("returns error on empty URL in response", resolveCase{
				purl: "pkg:npm/empty-url-pkg@1.0.0",
				check: func(r *ResolveResult, err error) {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(ErrEmptyURL.Error()))
				},
			}),
			Entry("returns error on bad JSON", resolveCase{
				purl: "pkg:npm/bad-json@1.0.0",
				check: func(r *ResolveResult, err error) {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("parse response"))
				},
			}),
		)

		It("returns error on server error (without retry)", func() {
			timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			noRetryService := NewService(ServiceConfig{
				ServerURL:  ts.URL,
				HTTPClient: &http.Client{Timeout: 30 * time.Second},
			})
			_, err := noRetryService.Resolve(timeoutCtx, "pkg:npm/server-error@1.0.0")
			Expect(err).To(HaveOccurred())
		})

		It("sends werf User-Agent header", func() {
			var capturedUA string
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedUA = r.UserAgent()
				w.WriteHeader(http.StatusNotFound)
			})
			uaTS := httptest.NewServer(handler)
			defer uaTS.Close()

			uaService := NewService(ServiceConfig{ServerURL: uaTS.URL})
			_, _ = uaService.Resolve(ctx, "pkg:npm/lodash@4.17.21")

			Expect(capturedUA).To(Equal(werf.UserAgent))
		})

		It("counts resolve calls", func() {
			_, _ = service.Resolve(ctx, "pkg:npm/lodash@4.17.21")
			_, _ = service.Resolve(ctx, "pkg:npm/express@4.18.2")
			Expect(*calls).To(Equal(2))
		})
	})
})
