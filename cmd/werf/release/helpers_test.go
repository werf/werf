package release_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	releasecommon "github.com/werf/nelm/v2/pkg/helm/pkg/release/common"
	releasev1 "github.com/werf/nelm/v2/pkg/helm/pkg/release/v1"
	"github.com/werf/werf/v3/cmd/werf/release/get"
	"github.com/werf/werf/v3/cmd/werf/release/history"
	"github.com/werf/werf/v3/cmd/werf/release/list"
)

const (
	subCommandEnv     = "WERF_TEST_RELEASE_SUBCOMMAND"
	subCommandArgsEnv = "WERF_TEST_RELEASE_SUBCOMMAND_ARGS"

	testReleaseName      = "mytest"
	testReleaseNamespace = "mytest-ns"
)

func runSubCommand(subCommand string, args []string) int {
	ctx := context.Background()

	var cmd *cobra.Command
	switch subCommand {
	case "list":
		cmd = list.NewCmd(ctx)
	case "get":
		cmd = get.NewCmd(ctx)
	case "history":
		cmd = history.NewCmd(ctx)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", subCommand)
		return 2
	}

	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "execute: %s\n", err)
		return 1
	}

	return 0
}

type commandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func runReleaseCommand(subCommand, ciEnvVar string, args []string) commandResult {
	ginkgo.GinkgoHelper()

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(cleanEnviron(),
		subCommandEnv+"="+subCommand,
		subCommandArgsEnv+"="+strings.Join(args, "\n"),
		ciEnvVar+"=true",
		"WERF_HOME="+ginkgo.GinkgoT().TempDir(),
		"WERF_TELEMETRY=0",
		"WERF_DISABLE_AUTO_HOST_CLEANUP=1",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	result := commandResult{}
	err := cmd.Run()

	var exitErr *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exitErr):
		result.ExitCode = exitErr.ExitCode()
	default:
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	return result
}

func cleanEnviron() []string {
	return lo.Filter(os.Environ(), func(entry string, _ int) bool {
		name, _, _ := strings.Cut(entry, "=")

		switch name {
		case "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "CI":
			return false
		}

		return !strings.HasPrefix(name, "WERF_")
	})
}

func startKubeStub() string {
	ginkgo.GinkgoHelper()

	secret := releaseSecret(1)
	secretList, err := json.Marshal(corev1.SecretList{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "SecretList"},
		ListMeta: metav1.ListMeta{ResourceVersion: "1"},
		Items:    []corev1.Secret{secret},
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	singleSecret, err := json.Marshal(secret)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body []byte

		switch path := strings.TrimSuffix(r.URL.Path, "/"); {
		case path == "/version":
			body = []byte(`{"major":"1","minor":"30","gitVersion":"v1.30.0"}`)
		case strings.HasSuffix(path, "/secrets/"+secret.Name):
			body = singleSecret
		case strings.HasSuffix(path, "/secrets"):
			body = secretList
		default:
			ginkgo.GinkgoWriter.Printf("kube stub: unhandled %s %s\n", r.Method, r.URL)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(body); err != nil {
			ginkgo.GinkgoWriter.Printf("kube stub: write response for %s: %s\n", r.URL, err)
		}
	}))
	ginkgo.DeferCleanup(server.Close)

	kubeConfigPath := filepath.Join(ginkgo.GinkgoT().TempDir(), "kubeconfig")
	gomega.Expect(os.WriteFile(kubeConfigPath, fmt.Appendf(nil, `apiVersion: v1
kind: Config
clusters:
- name: stub
  cluster:
    server: %s
contexts:
- name: stub
  context:
    cluster: stub
    user: stub
    namespace: %s
current-context: stub
users:
- name: stub
  user: {}
`, server.URL, testReleaseNamespace), 0o600)).To(gomega.Succeed())

	return kubeConfigPath
}

func releaseSecret(revision int) corev1.Secret {
	ginkgo.GinkgoHelper()

	rel := &releasev1.Release{
		Name:      testReleaseName,
		Namespace: testReleaseNamespace,
		Version:   revision,
		Info: &releasev1.Info{
			FirstDeployed: time.Unix(0, 0).UTC(),
			LastDeployed:  time.Unix(0, 0).UTC(),
			Status:        releasecommon.StatusDeployed,
			Description:   "Install complete",
		},
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{APIVersion: chart.APIVersionV2, Name: testReleaseName, Version: "1.0.0"},
		},
		Manifest: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: mytest\n",
	}

	encoded, err := json.Marshal(rel)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	var gzipped bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipped)
	_, err = gzipWriter.Write(encoded)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(gzipWriter.Close()).To(gomega.Succeed())

	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("sh.helm.release.v1.%s.v%d", testReleaseName, revision),
			Namespace: testReleaseNamespace,
			Labels: map[string]string{
				"name":    testReleaseName,
				"owner":   "helm",
				"status":  string(releasecommon.StatusDeployed),
				"version": strconv.Itoa(revision),
			},
		},
		Type: "helm.sh/release.v1",
		Data: map[string][]byte{"release": []byte(base64.StdEncoding.EncodeToString(gzipped.Bytes()))},
	}
}
