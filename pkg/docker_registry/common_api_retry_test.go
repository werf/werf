package docker_registry

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("registry connection retries", func() {
	ginkgo.DescribeTable("preserves the bounded transport policy through tags discovery", func(failure string, persistent bool, expectedDials int, succeeds bool) {
		fixture := newRegistryRetryFixture()
		fixture.failure = func(ctx context.Context, attempt int) error {
			if !persistent && attempt > 1 {
				return nil
			}
			switch failure {
			case "refused":
				return fixture.refusedDial(ctx)
			case "temporary":
				return &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ETIMEDOUT}
			default:
				return x509.UnknownAuthorityError{}
			}
		}
		tags, err := fixture.api.Tags(context.Background(), fixture.reference)
		if succeeds {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(tags).To(gomega.Equal([]string{"ok"}))
		} else {
			gomega.Expect(err).To(gomega.HaveOccurred())
		}
		gomega.Expect(fixture.tagDials.Load()).To(gomega.Equal(int64(expectedDials)))
	},
		ginkgo.Entry("recovers after a refused connection", "refused", false, 2, true),
		ginkgo.Entry("stops after three refused connections", "refused", true, 3, false),
		ginkgo.Entry("keeps retrying temporary failures", "temporary", false, 2, true),
		ginkgo.Entry("preserves the existing temporary retry budget", "temporary", true, 3, false),
		ginkgo.Entry("does not retry certificate failures", "certificate", true, 1, false),
	)

	ginkgo.DescribeTable("does not dial again after the request ends", func(deadline bool) {
		fixture := newRegistryRetryFixture()
		ctx, cancel := context.WithCancel(context.Background())
		if deadline {
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		}
		defer cancel()
		fixture.failure = func(ctx context.Context, attempt int) error {
			err := fixture.refusedDial(ctx)
			if deadline {
				<-ctx.Done()
			} else {
				cancel()
			}
			return err
		}
		_, err := fixture.api.Tags(ctx, fixture.reference)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(fixture.tagDials.Load()).To(gomega.Equal(int64(1)))
	}, ginkgo.Entry("canceled", false), ginkgo.Entry("deadline exceeded", true))

	ginkgo.It("does not retry an authorization rejection", func() {
		fixture := newRegistryRetryFixture()
		fixture.failure = func(context.Context, int) error { return nil }
		fixture.tagStatus = http.StatusUnauthorized
		_, err := fixture.api.Tags(context.Background(), fixture.reference)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(fixture.tagDials.Load()).To(gomega.Equal(int64(1)))
	})

	ginkgo.DescribeTable("does not classify ended requests as retryable", func(ended error) {
		gomega.Expect(isRetryableRegistryTransportError(errors.Join(ended, syscall.ECONNREFUSED))).To(gomega.BeFalse())
	}, ginkgo.Entry("canceled", context.Canceled), ginkgo.Entry("deadline exceeded", context.DeadlineExceeded))
})
