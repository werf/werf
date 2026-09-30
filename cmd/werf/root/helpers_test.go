package root

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/moby/moby/client/pkg/versions"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/nelm/v2/pkg/helm/intern/resolver"
	helmchart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	helmregistry "github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/werf/v3/cmd/werf/common"
	"github.com/werf/werf/v3/pkg/logging"
)

const dockerCommandArgsEnv = "WERF_TEST_DOCKER_COMMAND_ARGS"

func TestMain(m *testing.M) {
	if args := os.Getenv(dockerCommandArgsEnv); args != "" {
		os.Exit(runDockerCommand(strings.Split(args, "\n")))
	}
	os.Exit(m.Run())
}

func runDockerCommand(args []string) int {
	ctx := graceful.WithTermination(context.Background())
	defer graceful.Shutdown(ctx, func(_ context.Context, _ graceful.TerminationDescriptor) {})
	ctx = logging.WithLogger(ctx)
	cmd, err := ConstructRootCmd(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if args[0] == "components" {
		cmd, _, err = cmd.Find([]string{"cleanup"})
		if err == nil {
			err = cmd.ParseFlags(args[1:])
		}
		if err == nil {
			var manager *common.ComponentsManager
			manager, _, err = common.InitCommonComponents(cmd.Context(), common.InitCommonComponentsOptions{
				Cmd: common.GetCmdDataFromContext(cmd.Context()), InitWerf: true,
				InitProcessContainerBackend: true, InitDockerRegistry: true,
			})
			if err == nil {
				fmt.Printf("mirrors: %v\n", manager.RegistryMirrors())
			}
		}
	} else {
		cmd.SetArgs(args)
		err = cmd.Execute()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func dockerCommand(dir, dockerHost string, args ...string) (string, string, int) {
	ginkgo.GinkgoHelper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0])
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if name == "HOME" || strings.HasPrefix(name, "XDG_") || strings.HasPrefix(name, "WERF_") || strings.HasPrefix(name, "DOCKER_") || strings.HasPrefix(name, "HELM_") || strings.HasPrefix(name, "KUBE") || strings.HasPrefix(name, "SSH_") {
			continue
		}
		cmd.Env = append(cmd.Env, entry)
	}
	cmd.Env = append(cmd.Env, dockerCommandArgsEnv+"="+strings.Join(append(args, "--dir", dir, "--docker-config", filepath.Join(dir, ".docker")), "\n"),
		"HOME="+ginkgo.GinkgoT().TempDir(), "XDG_CONFIG_HOME="+ginkgo.GinkgoT().TempDir(), "XDG_CACHE_HOME="+ginkgo.GinkgoT().TempDir(), "WERF_HOME="+ginkgo.GinkgoT().TempDir(), "WERF_TELEMETRY=0", "WERF_DISABLE_AUTO_HOST_CLEANUP=1", "WERF_BUILDAH_MODE=docker", "DOCKER_HOST="+dockerHost)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	gomega.Expect(ctx.Err()).NotTo(gomega.HaveOccurred(), "command exceeded 15s: %v\n%s\n%s", args, stdout.String(), stderr.String())
	var exitErr *exec.ExitError
	exitCode := 0
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	} else {
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
	return stdout.String(), stderr.String(), exitCode
}

func dockerProject(withImage bool) string {
	ginkgo.GinkgoHelper()
	dir := ginkgo.GinkgoT().TempDir()
	gomega.Expect(os.CopyFS(dir, os.DirFS("testdata/optional-docker"))).To(gomega.Succeed())
	if withImage {
		f, err := os.OpenFile(filepath.Join(dir, "werf.yaml"), os.O_APPEND|os.O_WRONLY, 0)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		_, err = f.WriteString("---\nimage: app\ndockerfile: Dockerfile\n")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(f.Close()).To(gomega.Succeed())
	}
	for _, args := range [][]string{{"init"}, {"add", "."}, {"-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "-m", "fixture"}} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		output, err := cmd.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "%s", output)
	}
	return dir
}

func settingsDaemon(apiVersion string, pingStatus, infoStatus int) (string, *atomic.Int32) {
	ginkgo.GinkgoHelper()
	requests := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		requests.Add(1)
		w.Header().Set("API-Version", apiVersion)
		w.Header().Set("OSType", "linux")
		w.Header().Set("Content-Type", "application/json")
		status, body := pingStatus, any(map[string]string{"message": "ping denied"})
		if r.URL.Path != "/_ping" {
			status = infoStatus
			body = map[string]string{"message": "settings denied"}
			requested := strings.TrimPrefix(strings.Split(r.URL.Path, "/")[1], "v")
			if versions.GreaterThan(requested, apiVersion) {
				status = http.StatusBadRequest
				body = map[string]string{"message": "client API exceeds advertised daemon API " + apiVersion}
			} else if status == http.StatusOK {
				body = map[string]any{"OSType": "linux", "Architecture": "amd64", "RegistryConfig": map[string]any{"Mirrors": []string{"https://mirror.example"}}}
			}
		}
		w.WriteHeader(status)
		gomega.Expect(json.NewEncoder(w).Encode(body)).To(gomega.Succeed())
	}))
	ginkgo.DeferCleanup(server.Close)
	return "tcp://" + server.Listener.Addr().String(), requests
}

func muteDaemon() (string, *atomic.Int32) {
	requests := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		<-r.Context().Done()
	}))
	ginkgo.DeferCleanup(server.Close)
	return "tcp://" + server.Listener.Addr().String(), requests
}

func planKubeConfig() string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		w.Header().Set("Content-Type", "application/json")
		var body any
		switch r.URL.Path {
		case "/version":
			body = map[string]string{"major": "1", "minor": "35", "gitVersion": "v1.35.0"}
		case "/api":
			body = map[string]any{"kind": "APIVersions", "apiVersion": "v1", "versions": []string{"v1"}}
		case "/apis":
			body = map[string]any{"kind": "APIGroupList", "apiVersion": "v1", "groups": []any{}}
		case "/api/v1":
			body = map[string]any{"kind": "APIResourceList", "apiVersion": "v1", "groupVersion": "v1", "resources": []any{
				map[string]any{"name": "secrets", "kind": "Secret", "namespaced": true, "verbs": []string{"get", "list"}},
				map[string]any{"name": "namespaces", "kind": "Namespace", "namespaced": false, "verbs": []string{"get", "list"}},
			}}
		default:
			switch {
			case strings.HasSuffix(r.URL.Path, "/secrets"):
				body = map[string]any{"kind": "SecretList", "apiVersion": "v1", "items": []any{}}
			case strings.HasPrefix(r.URL.Path, "/api/v1/namespaces/") && strings.Count(r.URL.Path, "/") == 4:
				body = map[string]any{"kind": "Namespace", "apiVersion": "v1", "metadata": map[string]string{"name": strings.TrimPrefix(r.URL.Path, "/api/v1/namespaces/")}}
			default:
				w.WriteHeader(http.StatusNotFound)
				body = map[string]any{"kind": "Status", "apiVersion": "v1", "status": "Failure", "reason": "NotFound", "code": 404}
			}
		}
		gomega.Expect(json.NewEncoder(w).Encode(body)).To(gomega.Succeed())
	}))
	ginkgo.DeferCleanup(server.Close)
	path := filepath.Join(ginkgo.GinkgoT().TempDir(), "kubeconfig")
	gomega.Expect(os.WriteFile(path, fmt.Appendf(nil, `apiVersion: v1
kind: Config
clusters:
- name: fake
  cluster:
    server: %s
contexts:
- name: fake
  context:
    cluster: fake
    user: fake
current-context: fake
users:
- name: fake
  user: {}
`, server.URL), 0o600)).To(gomega.Succeed())
	return path
}

func authenticatedRegistry(dir string) (*httptest.Server, *atomic.Int32) {
	requests := &atomic.Int32{}
	handler := registry.New(registry.Logger(log.New(io.Discard, "", 0)))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "fixture" || password != "fixture-password" {
			w.Header().Set("WWW-Authenticate", `Basic realm="fixture"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		requests.Add(1)
		handler.ServeHTTP(w, r)
	}))
	ginkgo.DeferCleanup(server.Close)
	configDir := filepath.Join(dir, ".docker")
	gomega.Expect(os.MkdirAll(configDir, 0o700)).To(gomega.Succeed())
	data, err := json.Marshal(map[string]any{"auths": map[string]any{
		strings.TrimPrefix(server.URL, "http://"): map[string]string{"auth": base64.StdEncoding.EncodeToString([]byte("fixture:fixture-password"))},
	}})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), data, 0o600)).To(gomega.Succeed())
	return server, requests
}

func addOCIDependency(dir, registryURL string) {
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	data := []byte("apiVersion: v2\nname: dependency\nversion: 1.0.0\n")
	gomega.Expect(tw.WriteHeader(&tar.Header{Name: "dependency/Chart.yaml", Mode: 0o600, Size: int64(len(data))})).To(gomega.Succeed())
	_, err := tw.Write(data)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(tw.Close()).To(gomega.Succeed())
	gomega.Expect(gz.Close()).To(gomega.Succeed())
	cli, err := helmregistry.NewClient(helmregistry.ClientOptPlainHTTP(), helmregistry.ClientOptCredentialsFile(filepath.Join(dir, ".docker", "config.json")), helmregistry.ClientOptWriter(io.Discard))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	host := strings.TrimPrefix(registryURL, "http://")
	_, err = cli.Push(archive.Bytes(), host+"/charts/dependency:1.0.0")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	chartPath := filepath.Join(dir, ".helm", "Chart.yaml")
	chart, err := os.ReadFile(chartPath)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	chart = fmt.Appendf(chart, "dependencies:\n- name: dependency\n  version: 1.0.0\n  repository: oci://%s/charts\n", host)
	gomega.Expect(os.WriteFile(chartPath, chart, 0o600)).To(gomega.Succeed())
	deps := []*helmchart.Dependency{{Name: "dependency", Version: "1.0.0", Repository: "oci://" + host + "/charts"}}
	digest, err := resolver.HashReq(deps, deps)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	lock, err := yaml.Marshal(helmchart.Lock{Dependencies: deps, Digest: digest, Generated: time.Now()})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(os.WriteFile(filepath.Join(dir, ".helm", "Chart.lock"), lock, 0o600)).To(gomega.Succeed())
	for _, args := range [][]string{{"add", ".helm/Chart.yaml", ".helm/Chart.lock"}, {"-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "-m", "OCI fixture"}} {
		output, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "%s", output)
	}
}
