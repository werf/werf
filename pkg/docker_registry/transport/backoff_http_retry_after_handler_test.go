package transport

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("backoff_http_retry_after_handler", func() {
	var rec *httptest.ResponseRecorder
	BeforeEach(func() {
		rec = httptest.NewRecorder()
	})
	It("should do nothing if http.status not in [301, 429, 503]", func() {
		rec.WriteHeader(http.StatusOK)

		_, err := backoffHttpRetryAfterHandler(rec.Result())
		Expect(err).To(Succeed())
	})
	It("should return wrapped tmp_err if http response has status 301 but Retry-After header is not present", func() {
		rec.WriteHeader(http.StatusMovedPermanently)

		_, err := backoffHttpRetryAfterHandler(rec.Result())
		Expect(errors.Is(err, ErrRetryAfterHeaderNotPresent)).To(BeTrue())
	})
	It("should return backoff.PermanentError if http response has status 429 but Retry-After header has value in invalid time format", func() {
		rec.Header().Set(retryAfterHeaderKey, "some string with five spaces")
		rec.WriteHeader(http.StatusTooManyRequests)

		_, err := backoffHttpRetryAfterHandler(rec.Result())
		var errT *backoff.PermanentError
		Expect(errors.As(err, &errT)).To(BeTrue())
	})
	It("should return backoff.RetryAfterError if http response has status 429 and header has valid value time", func() {
		retryAfterSeconds := 10
		retryAfterTime := time.Now().Add(time.Second * time.Duration(retryAfterSeconds))
		rec.Header().Set(retryAfterHeaderKey, retryAfterTime.Format(time.RFC1123))
		rec.WriteHeader(http.StatusTooManyRequests)

		_, err := backoffHttpRetryAfterHandler(rec.Result())
		var errT *backoff.RetryAfterError
		Expect(errors.As(err, &errT)).To(BeTrue())
		Expect(int(errT.Duration.Seconds())).To(Or(Equal(retryAfterSeconds), Equal(retryAfterSeconds-1)))
	})
	It("should return backoff.PermanentError if http response has status 503 and header has invalid seconds value", func() {
		rec.Header().Set(retryAfterHeaderKey, "test")
		rec.WriteHeader(http.StatusServiceUnavailable)

		_, err := backoffHttpRetryAfterHandler(rec.Result())
		var errT *backoff.PermanentError
		Expect(errors.As(err, &errT)).To(BeTrue())
	})
	It("should return backoff.RetryAfterError if http response has status 503 and header has valid seconds value", func() {
		retryAfterSeconds := 10
		rec.Header().Set(retryAfterHeaderKey, strconv.Itoa(retryAfterSeconds))
		rec.WriteHeader(http.StatusTooManyRequests)

		_, err := backoffHttpRetryAfterHandler(rec.Result())
		var errT *backoff.RetryAfterError
		Expect(errors.As(err, &errT)).To(BeTrue())
		Expect(int(errT.Duration.Seconds())).To(Or(Equal(retryAfterSeconds), Equal(retryAfterSeconds-1)))
	})
})
