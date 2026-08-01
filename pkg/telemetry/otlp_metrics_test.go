package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitMetrics_Disabled(t *testing.T) {
	h, err := InitMetrics(context.Background(), false, "")
	require.NoError(t, err)
	require.NotNil(t, h)

	// No-Op path must be safe to drive through the full lifecycle.
	h.MetricsStart(context.Background(), "werf build")
	h.MetricsEnd(context.Background(), 0)
	require.NoError(t, h.MetricsShutdown(context.Background()))
}

func TestInitMetrics_EnabledWithoutEndpoint(t *testing.T) {
	_, err := InitMetrics(context.Background(), true, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "endpoint is not set")
}

func TestInitMetrics_RejectsUnsupportedScheme(t *testing.T) {
	_, err := InitMetrics(context.Background(), true, "ftp://collector:4318/v1/metrics")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported scheme")
}

func TestInitMetrics_RejectsMissingHost(t *testing.T) {
	_, err := InitMetrics(context.Background(), true, "http:///v1/metrics")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing host")
}

func TestInitMetrics_EnabledLifecycle(t *testing.T) {
	// A fake local OTLP collector so the shutdown flush has somewhere to
	// succeed, instead of depending on a real network endpoint.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	h, err := InitMetrics(context.Background(), true, srv.URL+"/v1/metrics")
	require.NoError(t, err)
	require.NotNil(t, h)

	h.MetricsStart(context.Background(), "werf build")
	h.MetricsEnd(context.Background(), 0)
	require.NoError(t, h.MetricsShutdown(context.Background()))
}

func TestMetricsHandle_NilIsSafe(t *testing.T) {
	var h *MetricsHandle
	h.MetricsStart(context.Background(), "werf build")
	h.MetricsEnd(context.Background(), 0)
	require.NoError(t, h.MetricsShutdown(context.Background()))
}
