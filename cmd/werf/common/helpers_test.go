package common

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/registry"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

type writeRecordingRegistry struct {
	handler http.Handler

	mu     sync.Mutex
	writes []string
}

var _ http.Handler = (*writeRecordingRegistry)(nil)

func (r *writeRecordingRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		r.mu.Lock()
		r.writes = append(r.writes, req.Method+" "+req.URL.Path)
		r.mu.Unlock()
	}
	r.handler.ServeHTTP(w, req)
}

func (r *writeRecordingRegistry) recordedWrites() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.writes...)
}

func startRecordingRegistry(repo string) (string, *writeRecordingRegistry) {
	recorder := &writeRecordingRegistry{handler: registry.New()}
	server := httptest.NewServer(recorder)
	ginkgo.DeferCleanup(server.Close)
	return strings.TrimPrefix(server.URL, "http://") + "/" + repo, recorder
}

func storageManagerCmdData(args ...string) *CmdData {
	command := &cobra.Command{}
	data := &CmdData{}
	SetupSecondaryStagesStorageOptions(data, command)
	SetupCacheStagesStorageOptions(data, command)
	SetupRepoOptions(data, command, RepoDataOptions{OptionalRepo: true})
	SetupFinalRepo(data, command)
	SetupContainerRegistryMirror(data, command)
	SetupSynchronization(data, command)
	SetupCheckBuiltImages(data, command)
	gomega.Expect(command.ParseFlags(args)).To(gomega.Succeed())
	return data
}
