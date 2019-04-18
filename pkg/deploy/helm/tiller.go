package helm

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/flant/logboek"
	"github.com/gosuri/uitable"
	"github.com/gosuri/uitable/util/strutil"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/storage"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/tiller"
	tiller_env "k8s.io/helm/pkg/tiller/environment"
	"k8s.io/helm/pkg/timeconv"

	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	tillerReleaseServer = &tiller.ReleaseServer{}
	tillerSettings      = tiller_env.New()
	helmSettings        helm_env.EnvSettings
	resourcesWaiter     *ResourcesWaiter
	releaseLogMessages  []string

	WerfTemplateEngine     = NewWerfEngine()
	WerfTemplateEngineName = "werfGoTpl"

	defaultTimeout           = int64((24 * time.Hour).Seconds())
	defaultReleaseHistoryMax = int32(256)

	DefaultReleaseStorageNamespace = "kube-system"

	ConfigMapStorage = "configmap"
	SecretStorage    = "secret"

	ErrNoDeployedReleaseRevisionFound = errors.New("no DEPLOYED release revision found")
)

func Init(kubeConfig, kubeContext, helmReleaseStorageNamespace, helmReleaseStorageType string) error {
	if err := initTiller(kubeConfig, kubeContext, helmReleaseStorageNamespace, helmReleaseStorageType); err != nil {
		return err
	}

	return nil
}

func initTiller(kubeConfig, kubeContext, helmReleaseStorageNamespace, helmReleaseStorageType string) error {
	helmSettings.KubeConfig = kubeConfig
	helmSettings.KubeContext = kubeContext
	helmSettings.TillerNamespace = helmReleaseStorageNamespace

	configFlags := genericclioptions.NewConfigFlags(true)
	configFlags.Context = &helmSettings.KubeContext
	configFlags.KubeConfig = &helmSettings.KubeConfig
	configFlags.Namespace = &helmReleaseStorageNamespace

	kubeClient := kube.New(configFlags)

	resourcesWaiter = &ResourcesWaiter{Client: kubeClient}
	kubeClient.SetResourcesWaiter(resourcesWaiter)

	tillerSettings.KubeClient = kubeClient
	tillerSettings.EngineYard[WerfTemplateEngineName] = WerfTemplateEngine

	clientset, err := kubeClient.KubernetesClientSet()
	if err != nil {
		return err
	}

	if _, err := clientset.CoreV1().Namespaces().Get(helmReleaseStorageNamespace, metav1.GetOptions{}); err != nil {
		if kubeErrors.IsNotFound(err) {
			if _, err := clientset.CoreV1().Namespaces().Create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: helmReleaseStorageNamespace}}); err != nil {
				return fmt.Errorf("unable to create helm release storage namespace '%s': %s", helmReleaseStorageNamespace, err)
			}
			logboek.LogInfoF("Created helm release storage namespace '%s'\n", helmReleaseStorageNamespace)
		} else {
			return fmt.Errorf("unable to initialize helm release storage in namespace '%s': %s", helmReleaseStorageNamespace, err)
		}
	}

	switch helmReleaseStorageType {
	case ConfigMapStorage:
		cfgmaps := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(helmReleaseStorageNamespace))
		cfgmaps.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("ConfigMaps release storage driver: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}
		tillerSettings.Releases = storage.Init(cfgmaps)
		tillerSettings.Releases.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("Release storage: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}
	case SecretStorage:
		secrets := driver.NewSecrets(clientset.CoreV1().Secrets(helmReleaseStorageNamespace))
		secrets.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("Secrets release storage driver: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}
		tillerSettings.Releases = storage.Init(secrets)
		tillerSettings.Releases.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("Release storage: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}
	default:
		return fmt.Errorf("unknown helm release storage type '%s'", helmReleaseStorageType)
	}

	tillerReleaseServer = tiller.NewReleaseServer(tillerSettings, clientset, false)
	tillerReleaseServer.Log = func(f string, args ...interface{}) {
		msg := fmt.Sprintf(fmt.Sprintf("Release server: %s", f), args...)
		releaseLogMessages = append(releaseLogMessages, msg)
	}

	return nil
}

type releaseContentOptions struct {
	Version int32
}

func releaseContent(releaseName string, opts releaseContentOptions) (*services.GetReleaseContentResponse, error) {
	ctx := helm.NewContext()
	req := &services.GetReleaseContentRequest{
		Name:    releaseName,
		Version: opts.Version,
	}

	return tillerReleaseServer.GetReleaseContent(ctx, req)
}

type releaseHistoryOptions struct {
	Max int32
}

func releaseHistory(releaseName string, opts releaseHistoryOptions) (*services.GetHistoryResponse, error) {
	max := opts.Max
	if opts.Max == 0 {
		max = defaultReleaseHistoryMax
	}

	ctx := helm.NewContext()
	req := &services.GetHistoryRequest{
		Name: releaseName,
		Max:  max,
	}

	return tillerReleaseServer.GetHistory(ctx, req)
}

type releaseStatusOptions struct {
	Version int32
}

func releaseStatus(releaseName string, opts releaseStatusOptions) (*services.GetReleaseStatusResponse, error) {
	ctx := helm.NewContext()
	req := &services.GetReleaseStatusRequest{
		Name:    releaseName,
		Version: opts.Version,
	}

	return tillerReleaseServer.GetReleaseStatus(ctx, req)
}

type releaseStatusCodeOptions struct {
	releaseStatusOptions
}

func releaseStatusCode(releaseName string, opts releaseStatusCodeOptions) (string, error) {
	resp, err := releaseStatus(releaseName, opts.releaseStatusOptions)
	if err != nil {
		return "", err
	}

	return resp.Info.Status.Code.String(), nil
}

type releaseDeleteOptions struct {
	Purge   bool
	Timeout int64
}

func releaseDelete(releaseName string, opts releaseDeleteOptions) error {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	timeout := opts.Timeout
	if opts.Timeout == 0 {
		timeout = defaultTimeout
	}

	ctx := helm.NewContext()
	req := &services.UninstallReleaseRequest{
		Name:    releaseName,
		Purge:   opts.Purge,
		Timeout: timeout,
	}

	_, err := tillerReleaseServer.UninstallRelease(ctx, req)
	if err != nil {
		for _, msg := range releaseLogMessages {
			logboek.LogInfoF("%s\n", msg)
		}
		return err
	}

	return nil
}

func isReleaseNotFoundError(err error) bool {
	return strings.HasSuffix(err.Error(), "not found")
}

type ReleaseInstallOptions struct {
	releaseInstallOptions

	Debug bool
}

func ReleaseInstall(out io.Writer, chartPath, releaseName, namespace string, values, set, setString []string, opts ReleaseInstallOptions) error {
	rawVals, err := vals(values, set, setString, []string{}, "", "", "")
	if err != nil {
		return err
	}

	if msgs := validation.IsDNS1123Subdomain(releaseName); releaseName != "" && len(msgs) > 0 {
		return fmt.Errorf("release name %s is invalid: %s", releaseName, strings.Join(msgs, ";"))
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	loadedChart, err := chartutil.Load(chartPath)
	if err != nil {
		return err
	}

	if req, err := chartutil.LoadRequirements(loadedChart); err == nil {
		// If checkDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/kubernetes/helm/issues/2209
		if err := renderutil.CheckDependencies(loadedChart, req); err != nil {
			return err
		}
	} else if err != chartutil.ErrRequirementsNotFound {
		return fmt.Errorf("cannot load requirements: %v", err)
	}

	resp, err := releaseInstall(loadedChart, releaseName, namespace, &chart.Config{Raw: string(rawVals)}, opts.releaseInstallOptions)
	if err != nil {
		return err
	}

	rel := resp.GetRelease()
	if rel == nil {
		return nil
	}

	if err := printRelease(out, rel, opts.DryRun); err != nil {
		return err
	}

	// If this is a dry run, we can't display status.
	if opts.DryRun {
		// This is special casing to avoid breaking backward compatibility:
		if resp.Release.Info.Description != "Dry run complete" {
			fmt.Fprintf(out, "WARNING: %s\n", resp.Release.Info.Description)
		}

		return nil
	}

	// Print the status like status command does
	status, err := releaseStatus(releaseName, releaseStatusOptions{})
	if err != nil {
		return err
	}
	printStatus(out, status)

	return nil
}

type ReleaseUpdateOptions struct {
	releaseUpdateOptions

	Debug bool
}

func ReleaseUpdate(out io.Writer, chartPath, releaseName string, values, set, setString []string, opts ReleaseUpdateOptions) error {
	rawVals, err := vals(values, set, setString, []string{}, "", "", "")
	if err != nil {
		return err
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	loadedChart, err := chartutil.Load(chartPath)
	if err == nil {
		if req, err := chartutil.LoadRequirements(loadedChart); err == nil {
			if err := renderutil.CheckDependencies(loadedChart, req); err != nil {
				return err
			}
		} else if err != chartutil.ErrRequirementsNotFound {
			return fmt.Errorf("cannot load requirements: %v", err)
		}
	} else {
		return err
	}

	resp, err := releaseUpdate(loadedChart, releaseName, &chart.Config{Raw: string(rawVals)}, opts.releaseUpdateOptions)
	if err != nil {
		return fmt.Errorf("UPGRADE FAILED: %v", err)
	}

	if opts.Debug {
		printRelease(out, resp.Release, opts.DryRun)
	}

	fmt.Fprintf(out, "Release %q has been upgraded. Happy Helming!\n", releaseName)

	// Print the status like status command does
	status, err := releaseStatus(releaseName, releaseStatusOptions{})
	if err != nil {
		return err
	}
	printStatus(out, status)

	return nil
}

type ReleaseRollbackOptions struct {
	releaseRollbackOptions
}

func ReleaseRollback(out io.Writer, releaseName string, revision int32, opts ReleaseRollbackOptions) error {
	if _, err := releaseRollback(releaseName, revision, opts.releaseRollbackOptions); err != nil {
		return err
	}

	fmt.Fprintf(out, "Rollback was a success! Happy Helming!\n")

	return nil
}

type releaseInstallOptions struct {
	Timeout int64
	Wait    bool
	DryRun  bool
}

func releaseInstall(chart *chart.Chart, releaseName, namespace string, values *chart.Config, opts releaseInstallOptions) (*services.InstallReleaseResponse, error) {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	timeout := opts.Timeout
	if opts.Timeout == 0 {
		timeout = defaultTimeout
	}

	err := chartutil.ProcessRequirementsEnabled(chart, values)
	if err != nil {
		return nil, err
	}

	err = chartutil.ProcessRequirementsImportValues(chart)
	if err != nil {
		return nil, err
	}

	ctx := helm.NewContext()
	req := &services.InstallReleaseRequest{
		Chart:     chart,
		Name:      releaseName,
		Namespace: namespace,
		Values:    values,
		Wait:      opts.Wait,
		DryRun:    opts.DryRun,
		Timeout:   timeout,
	}

	resp, err := tillerReleaseServer.InstallRelease(ctx, req)
	if err != nil {
		for _, msg := range releaseLogMessages {
			logboek.LogInfoF("%s\n", msg)
		}
		return nil, err
	}

	return resp, nil
}

type releaseUpdateOptions struct {
	Timeout       int64
	CleanupOnFail bool
	Wait          bool
	DryRun        bool
}

func releaseUpdate(chart *chart.Chart, releaseName string, values *chart.Config, opts releaseUpdateOptions) (*services.UpdateReleaseResponse, error) {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	timeout := opts.Timeout
	if opts.Timeout == 0 {
		timeout = defaultTimeout
	}

	err := chartutil.ProcessRequirementsEnabled(chart, values)
	if err != nil {
		return nil, err
	}

	err = chartutil.ProcessRequirementsImportValues(chart)
	if err != nil {
		return nil, err
	}

	ctx := helm.NewContext()
	req := &services.UpdateReleaseRequest{
		Chart:         chart,
		Name:          releaseName,
		Values:        values,
		Timeout:       timeout,
		CleanupOnFail: opts.CleanupOnFail,
		Wait:          opts.Wait,
		DryRun:        opts.DryRun,
	}

	resp, err := tillerReleaseServer.UpdateRelease(ctx, req)
	if err != nil {
		for _, msg := range releaseLogMessages {
			logboek.LogInfoF("%s\n", msg)
		}
		return nil, err
	}

	return resp, nil
}

type releaseRollbackOptions struct {
	Timeout       int64
	CleanupOnFail bool
	Wait          bool
	DryRun        bool
}

func releaseRollback(releaseName string, revision int32, opts releaseRollbackOptions) (*services.RollbackReleaseResponse, error) {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	timeout := opts.Timeout
	if opts.Timeout == 0 {
		timeout = defaultTimeout
	}

	ctx := helm.NewContext()
	req := &services.RollbackReleaseRequest{
		Name:          releaseName,
		Version:       revision,
		Timeout:       timeout,
		CleanupOnFail: opts.CleanupOnFail,
		Wait:          opts.Wait,
		DryRun:        opts.DryRun,
	}

	resp, err := tillerReleaseServer.RollbackRelease(ctx, req)
	if err != nil {
		for _, msg := range releaseLogMessages {
			logboek.LogInfoF("%s\n", msg)
		}
		return nil, err
	}

	return resp, nil
}

var printReleaseTemplate = `REVISION: {{.Release.Version}}
RELEASED: {{.ReleaseDate}}
CHART: {{.Release.Chart.Metadata.Name}}-{{.Release.Chart.Metadata.Version}}
USER-SUPPLIED VALUES:
{{.Release.Config.Raw}}
COMPUTED VALUES:
{{.ComputedValues}}
HOOKS:
{{- range .Release.Hooks }}
---
# {{.Name}}
{{.Manifest}}
{{- end }}
MANIFEST:
{{.Release.Manifest}}
`

func printRelease(out io.Writer, rel *release.Release, debug bool) error {
	if rel == nil {
		return nil
	}

	fmt.Fprintf(out, "NAME:   %s\n", rel.Name)
	if debug {
		cfg, err := chartutil.CoalesceValues(rel.Chart, rel.Config)
		if err != nil {
			return err
		}
		cfgStr, err := cfg.YAML()
		if err != nil {
			return err
		}

		data := map[string]interface{}{
			"Release":        rel,
			"ComputedValues": cfgStr,
			"ReleaseDate":    timeconv.Format(rel.Info.LastDeployed, time.ANSIC),
		}

		return tpl(printReleaseTemplate, data, out)
	}

	return nil
}

func tpl(t string, vals map[string]interface{}, out io.Writer) error {
	tt, err := template.New("_").Parse(t)
	if err != nil {
		return err
	}
	return tt.Execute(out, vals)
}

// PrintStatus prints out the status of a release. Shared because also used by
// install / upgrade
func printStatus(out io.Writer, resp *services.GetReleaseStatusResponse) {
	if resp.Info.LastDeployed != nil {
		fmt.Fprintf(out, "LAST DEPLOYED: %s\n", timeconv.String(resp.Info.LastDeployed))
	}
	fmt.Fprintf(out, "NAMESPACE: %s\n", resp.Namespace)
	fmt.Fprintf(out, "STATUS: %s\n", resp.Info.Status.Code)
	fmt.Fprintf(out, "\n")
	if len(resp.Info.Status.Resources) > 0 {
		re := regexp.MustCompile("  +")

		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "RESOURCES:\n%s\n", re.ReplaceAllString(resp.Info.Status.Resources, "\t"))
		w.Flush()
	}
	if resp.Info.Status.LastTestSuiteRun != nil {
		lastRun := resp.Info.Status.LastTestSuiteRun
		fmt.Fprintf(out, "TEST SUITE:\n%s\n%s\n\n%s\n",
			fmt.Sprintf("Last Started: %s", timeconv.String(lastRun.StartedAt)),
			fmt.Sprintf("Last Completed: %s", timeconv.String(lastRun.CompletedAt)),
			formatTestResults(lastRun.Results))
	}

	if len(resp.Info.Status.Notes) > 0 {
		fmt.Fprintf(out, "NOTES:\n%s\n", resp.Info.Status.Notes)
	}
}

func formatTestResults(results []*release.TestRun) string {
	tbl := uitable.New()
	tbl.MaxColWidth = 50
	tbl.AddRow("TEST", "STATUS", "INFO", "STARTED", "COMPLETED")
	for i := 0; i < len(results); i++ {
		r := results[i]
		n := r.Name
		s := strutil.PadRight(r.Status.String(), 10, ' ')
		i := r.Info
		ts := timeconv.String(r.StartedAt)
		tc := timeconv.String(r.CompletedAt)
		tbl.AddRow(n, s, i, ts, tc)
	}
	return tbl.String()
}
