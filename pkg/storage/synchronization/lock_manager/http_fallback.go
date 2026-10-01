package lock_manager

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/werf/logboek"
)

type synchronizationUnavailableError struct{ err error }

func (e *synchronizationUnavailableError) Error() string { return e.err.Error() }
func (e *synchronizationUnavailableError) Unwrap() error { return e.err }

func classifySynchronizationRequestError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.Canceled) {
		return err
	}
	var certVerification *tls.CertificateVerificationError
	var unknownAuthority x509.UnknownAuthorityError
	var invalidCertificate x509.CertificateInvalidError
	var hostname x509.HostnameError
	var tlsRecord tls.RecordHeaderError
	if errors.As(err, &certVerification) || errors.As(err, &unknownAuthority) || errors.As(err, &invalidCertificate) || errors.As(err, &hostname) || errors.As(err, &tlsRecord) {
		return err
	}
	var dns *net.DNSError
	var operation *net.OpError
	var network net.Error
	if errors.As(err, &dns) || (errors.As(err, &operation) && (operation.Op == "dial" || operation.Op == "read" || operation.Op == "write")) || (errors.As(err, &network) && network.Timeout()) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return &synchronizationUnavailableError{err: err}
	}
	return err
}

type httpFallbackState struct {
	disabled atomic.Bool
	warning  sync.Once
}

func (s *httpFallbackState) disable(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	var unavailable *synchronizationUnavailableError
	if !errors.As(err, &unavailable) {
		return false
	}
	s.warning.Do(func() {
		logboek.Context(ctx).Warn().LogF("Public synchronization service is unavailable: %s. Continuing without distributed publication locks for the rest of this command. Concurrent builds may publish duplicate images; if their outputs differ, a retry may select different contents. Configure --synchronization with a shared server to require synchronization.\n", err)
		s.disabled.Store(true)
	})
	return true
}
