package docker

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/moby/moby/client"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func daemonSettingsContext(handler http.Handler) context.Context {
	server := httptest.NewServer(handler)
	ginkgo.DeferCleanup(server.Close)
	api, err := client.New(client.WithHost(server.URL))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(api.Close)
	ctx := context.WithValue(context.Background(), ctxDockerCliKey, true)
	return context.WithValue(ctx, ctxAPIClientKey, api)
}
