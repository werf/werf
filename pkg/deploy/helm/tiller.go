package helm

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/flant/werf/pkg/util/secretvalues"

	"github.com/gosuri/uitable"
	"github.com/gosuri/uitable/util/strutil"
	"k8s.io/helm/pkg/proto/hapi/release"

	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/storage"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/tiller"
	tiller_env "k8s.io/helm/pkg/tiller/environment"
	"k8s.io/helm/pkg/timeconv"

	"github.com/flant/logboek"
)

type ThreeWayMergeModeType string

var (
	threeWayMergeEnabled         ThreeWayMergeModeType = "enabled"
	threeWayMergeOnlyNewReleases ThreeWayMergeModeType = "onlyNewReleases"
	threeWayMergeDisabled        ThreeWayMergeModeType = "disabled"
)

var (
	tillerReleaseServer          = &tiller.ReleaseServer{}
	tillerSettings               = tiller_env.New()
	helmSettings                 helm_env.EnvSettings
	resourcesWaiter              *ResourcesWaiter
	releaseLogMessages           []string
	releaseLogSecretValuesToMask []string

	WerfTemplateEngine     = NewWerfEngine()
	WerfTemplateEngineName = "werfGoTpl"

	defaultTimeout           = int64((24 * time.Hour).Seconds())
	defaultReleaseHistoryMax = int32(256)

	DefaultReleaseStorageNamespace = "kube-system"

	ConfigMapStorage = "configmap"
	SecretStorage    = "secret"

	LoadChartfileFunc = func(chartPath string) (*chart.Chart, error) {
		return chartutil.Load(chartPath)
	}

	ErrNoSuccessfullyDeployedReleaseRevisionFound = errors.New("no DEPLOYED release revision found")

	currentDate                                 = time.Now()
	threeWayMergeOnlyNewReleasesEnabledDeadline = time.Date(2019, 12, 1, 0, 0, 0, 0, time.UTC)
	threeWayMergeEnabledDeadline                = time.Date(2019, 12, 15, 0, 0, 0, 0, time.UTC)
	noDisableThreeWayMergeDeadline              = time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)
)

func SetReleaseLogSecretValuesToMask(secretValuesToMask []string) {
	releaseLogSecretValuesToMask = secretValuesToMask
}

func loadChartfile(chartPath string) (*chart.Chart, error) {
	return LoadChartfileFunc(chartPath)
}

type InitOptions struct {
	KubeConfig                  string
	KubeContext                 string
	HelmReleaseStorageNamespace string
	HelmReleaseStorageType      string

	InitNamespace bool

	StatusProgressPeriod      time.Duration
	HooksStatusProgressPeriod time.Duration

	ReleasesMaxHistory int

	WithoutKube bool
}

func Init(options InitOptions) error {
	if options.WithoutKube {
		tillerSettings.EngineYard[WerfTemplateEngineName] = WerfTemplateEngine

		tillerReleaseServer = tiller.NewReleaseServer(tillerSettings, nil, false)
		tillerReleaseServer.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("Release server: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}

		return nil
	}

	helmSettings.KubeConfig = options.KubeConfig
	helmSettings.KubeContext = options.KubeContext
	helmSettings.TillerNamespace = options.HelmReleaseStorageNamespace

	configFlags := genericclioptions.NewConfigFlags(true)
	configFlags.Context = &helmSettings.KubeContext
	configFlags.KubeConfig = &helmSettings.KubeConfig
	configFlags.Namespace = &options.HelmReleaseStorageNamespace

	kubeClient := kube.New(configFlags)
	kubeClient.Log = func(f string, args ...interface{}) {
		msg := fmt.Sprintf(fmt.Sprintf("Kube client: %s", f), args...)
		releaseLogMessages = append(releaseLogMessages, msg)
	}

	resourcesWaiter = &ResourcesWaiter{
		Client:                    kubeClient,
		StatusProgressPeriod:      options.StatusProgressPeriod,
		HooksStatusProgressPeriod: options.HooksStatusProgressPeriod,
	}
	kubeClient.SetResourcesWaiter(resourcesWaiter)

	tillerSettings.KubeClient = kubeClient
	tillerSettings.EngineYard[WerfTemplateEngineName] = WerfTemplateEngine

	clientset, err := kubeClient.KubernetesClientSet()
	if err != nil {
		return err
	}

	if options.InitNamespace {
		if _, err := clientset.CoreV1().Namespaces().Get(options.HelmReleaseStorageNamespace, metav1.GetOptions{}); err != nil {
			if kubeErrors.IsNotFound(err) {
				if _, err := clientset.CoreV1().Namespaces().Create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: options.HelmReleaseStorageNamespace}}); err != nil {
					return fmt.Errorf("unable to create helm release storage namespace '%s': %s", options.HelmReleaseStorageNamespace, err)
				}

				logboek.Default.LogFDetails("Created helm release storage namespace '%s'\n", options.HelmReleaseStorageNamespace)
			} else {
				return fmt.Errorf("unable to initialize helm release storage in namespace '%s': %s", options.HelmReleaseStorageNamespace, err)
			}
		}
	}

	switch options.HelmReleaseStorageType {
	case ConfigMapStorage:
		cfgmaps := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(options.HelmReleaseStorageNamespace))
		cfgmaps.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("ConfigMaps release storage driver: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}
		tillerSettings.Releases = storage.Init(cfgmaps)
		tillerSettings.Releases.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("Release storage: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}

		if options.ReleasesMaxHistory > 0 {
			tillerSettings.Releases.MaxHistory = options.ReleasesMaxHistory
		}

	case SecretStorage:
		secrets := driver.NewSecrets(clientset.CoreV1().Secrets(options.HelmReleaseStorageNamespace))
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
		return fmt.Errorf("unknown helm release storage type '%s'", options.HelmReleaseStorageType)
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
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	ctx := helm.NewContext()
	req := &services.GetReleaseStatusRequest{
		Name:    releaseName,
		Version: opts.Version,
	}

	res, err := tillerReleaseServer.GetReleaseStatus(ctx, req)
	if err != nil {
		for _, msg := range releaseLogMessages {
			logboek.Default.LogFDetails("%s\n", secretvalues.MaskSecretValuesInString(releaseLogSecretValuesToMask, msg))
		}
	}
	return res, err
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
			logboek.Default.LogFDetails("%s\n", secretvalues.MaskSecretValuesInString(releaseLogSecretValuesToMask, msg))
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

func ReleaseInstall(chartPath, releaseName, namespace string, values []string, secretValues []map[string]interface{}, set, setString []string, threeWayMergeMode ThreeWayMergeModeType, opts ReleaseInstallOptions) error {
	rawVals, err := vals(values, secretValues, set, setString, []string{}, "", "", "")
	if err != nil {
		return err
	}

	if msgs := validation.IsDNS1123Subdomain(releaseName); releaseName != "" && len(msgs) > 0 {
		return fmt.Errorf("release name %s is invalid: %s", releaseName, strings.Join(msgs, ";"))
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	loadedChart, err := loadChartfile(chartPath)
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

	_, err = releaseInstall(loadedChart, releaseName, namespace, &chart.Config{Raw: string(rawVals)}, threeWayMergeMode, opts.releaseInstallOptions)
	if err != nil {
		return err
	}

	return fprintReleaseStatus(logboek.GetOutStream(), releaseName)
}

type ReleaseUpdateOptions struct {
	releaseUpdateOptions

	Debug bool
}

func ReleaseUpdate(chartPath, releaseName string, values []string, secretValues []map[string]interface{}, set, setString []string, threeWayMergeMode ThreeWayMergeModeType, opts ReleaseUpdateOptions) error {
	rawVals, err := vals(values, secretValues, set, setString, []string{}, "", "", "")
	if err != nil {
		return err
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	loadedChart, err := loadChartfile(chartPath)
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

	_, err = releaseUpdate(loadedChart, releaseName, &chart.Config{Raw: string(rawVals)}, threeWayMergeMode, opts.releaseUpdateOptions)
	if err != nil {
		return err
	}

	return fprintReleaseStatus(logboek.GetOutStream(), releaseName)
}

type ReleaseRollbackOptions struct {
	releaseRollbackOptions
}

func ReleaseRollback(releaseName string, revision int32, threeWayMergeMode ThreeWayMergeModeType, opts ReleaseRollbackOptions) error {
	if _, err := releaseRollback(releaseName, revision, threeWayMergeMode, opts.releaseRollbackOptions); err != nil {
		return err
	}

	return nil
}

type releaseInstallOptions struct {
	Timeout int64
	Wait    bool
	DryRun  bool
}

func releaseInstall(chart *chart.Chart, releaseName, namespace string, values *chart.Config, userSpecifiedThreeWayMergeMode ThreeWayMergeModeType, opts releaseInstallOptions) (*services.InstallReleaseResponse, error) {
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
		Chart:             chart,
		Name:              releaseName,
		Namespace:         namespace,
		Values:            values,
		Wait:              opts.Wait,
		DryRun:            opts.DryRun,
		Timeout:           timeout,
		ThreeWayMergeMode: convertThreeWayMergeModeForHelm(getActualThreeWayMergeMode(userSpecifiedThreeWayMergeMode)),
	}

	logboek.Default.LogFHighlight("Using three-way-merge mode \"%s\"\n", getActualThreeWayMergeMode(userSpecifiedThreeWayMergeMode))

	resp, err := tillerReleaseServer.InstallRelease(ctx, req)
	if err != nil {
		displayReleaseLogMessages()
		if resp != nil {
			displayWarnings(userSpecifiedThreeWayMergeMode, resp.Release)
		}
		return nil, err
	}
	displayWarnings(userSpecifiedThreeWayMergeMode, resp.Release)

	return resp, nil
}

type releaseUpdateOptions struct {
	Timeout       int64
	CleanupOnFail bool
	Wait          bool
	DryRun        bool
}

func releaseUpdate(chart *chart.Chart, releaseName string, values *chart.Config, userSpecifiedThreeWayMergeMode ThreeWayMergeModeType, opts releaseUpdateOptions) (*services.UpdateReleaseResponse, error) {
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
		Chart:             chart,
		Name:              releaseName,
		Values:            values,
		Timeout:           timeout,
		CleanupOnFail:     opts.CleanupOnFail,
		Wait:              opts.Wait,
		DryRun:            opts.DryRun,
		ThreeWayMergeMode: convertThreeWayMergeModeForHelm(getActualThreeWayMergeMode(userSpecifiedThreeWayMergeMode)),
	}

	logboek.Default.LogFHighlight("Using three-way-merge mode \"%s\"\n", getActualThreeWayMergeMode(userSpecifiedThreeWayMergeMode))

	resp, err := tillerReleaseServer.UpdateRelease(ctx, req)
	if err != nil {
		displayReleaseLogMessages()
		if resp != nil {
			displayWarnings(userSpecifiedThreeWayMergeMode, resp.Release)
		}
		return nil, err
	}
	displayWarnings(userSpecifiedThreeWayMergeMode, resp.Release)

	return resp, nil
}

type releaseRollbackOptions struct {
	Timeout       int64
	CleanupOnFail bool
	Wait          bool
	DryRun        bool
}

func releaseRollback(releaseName string, revision int32, userSpecifiedThreeWayMergeMode ThreeWayMergeModeType, opts releaseRollbackOptions) (*services.RollbackReleaseResponse, error) {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	timeout := opts.Timeout
	if opts.Timeout == 0 {
		timeout = defaultTimeout
	}

	ctx := helm.NewContext()
	req := &services.RollbackReleaseRequest{
		Name:              releaseName,
		Version:           revision,
		Timeout:           timeout,
		CleanupOnFail:     opts.CleanupOnFail,
		Wait:              opts.Wait,
		DryRun:            opts.DryRun,
		ThreeWayMergeMode: convertThreeWayMergeModeForHelm(getActualThreeWayMergeMode(userSpecifiedThreeWayMergeMode)),
	}

	logboek.Default.LogFHighlight("Using three-way-merge mode \"%s\"\n", getActualThreeWayMergeMode(userSpecifiedThreeWayMergeMode))

	resp, err := tillerReleaseServer.RollbackRelease(ctx, req)
	if err != nil {
		displayReleaseLogMessages()
		if resp != nil {
			displayWarnings(userSpecifiedThreeWayMergeMode, resp.Release)
		}
		return nil, err
	}
	displayWarnings(userSpecifiedThreeWayMergeMode, resp.Release)

	return resp, nil
}

func displayReleaseLogMessages() {
	logboek.LogOptionalLn()
	_ = logboek.Default.LogBlock("Debug info", logboek.LevelLogBlockOptions{}, func() error {
		for _, msg := range releaseLogMessages {
			_, _ = logboek.OutF("%s\n", logboek.DetailsStyle().Colorize(secretvalues.MaskSecretValuesInString(releaseLogSecretValuesToMask, msg)))
		}

		return nil
	})
}

func fprintReleaseStatus(out io.Writer, releaseName string) error {
	ctx := helm.NewContext()
	req := &services.GetReleaseStatusRequest{Name: releaseName}

	status, err := tillerReleaseServer.GetReleaseStatus(ctx, req)
	if err != nil {
		return fmt.Errorf("error getting release %v status: %s", releaseName, err)
	}

	fmt.Fprintf(out, "NAME: %s\n", releaseName)
	printStatus(out, status)

	return nil
}

func printStatus(out io.Writer, res *services.GetReleaseStatusResponse) {
	if res.Info.LastDeployed != nil {
		fmt.Fprintf(out, "LAST DEPLOYED: %s\n", timeconv.String(res.Info.LastDeployed))
	}
	fmt.Fprintf(out, "NAMESPACE: %s\n", res.Namespace)
	fmt.Fprintf(out, "STATUS: %s\n", res.Info.Status.Code)
	fmt.Fprintf(out, "\n")
	if len(res.Info.Status.Resources) > 0 {
		re := regexp.MustCompile("  +")

		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "RESOURCES:\n%s\n", re.ReplaceAllString(res.Info.Status.Resources, "\t"))
		w.Flush()
	}
	if res.Info.Status.LastTestSuiteRun != nil {
		lastRun := res.Info.Status.LastTestSuiteRun
		fmt.Fprintf(out, "TEST SUITE:\n%s\n%s\n\n%s\n",
			fmt.Sprintf("Last Started: %s", timeconv.String(lastRun.StartedAt)),
			fmt.Sprintf("Last Completed: %s", timeconv.String(lastRun.CompletedAt)),
			formatTestResults(lastRun.Results))
	}

	if len(res.Info.Status.Notes) > 0 {
		fmt.Fprintf(out, "NOTES:\n%s\n", res.Info.Status.Notes)
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

func convertThreeWayMergeModeForHelm(threeWayMergeMode ThreeWayMergeModeType) services.ThreeWayMergeMode {
	switch threeWayMergeMode {
	case threeWayMergeEnabled:
		return services.ThreeWayMergeMode_enabled
	case threeWayMergeDisabled:
		return services.ThreeWayMergeMode_disabled
	case threeWayMergeOnlyNewReleases:
		return services.ThreeWayMergeMode_onlyNewReleases
	default:
		panic("non empty threeWayMergeMode required!")
	}
}

func getActualThreeWayMergeMode(userSpecifiedThreeWayMergeMode ThreeWayMergeModeType) ThreeWayMergeModeType {
	if currentDate.After(noDisableThreeWayMergeDeadline) {
		return threeWayMergeEnabled
	}

	if userSpecifiedThreeWayMergeMode != "" {
		return userSpecifiedThreeWayMergeMode
	} else if currentDate.Before(threeWayMergeOnlyNewReleasesEnabledDeadline) {
		return threeWayMergeDisabled
	} else if currentDate.Before(threeWayMergeEnabledDeadline) {
		return threeWayMergeOnlyNewReleases
	} else {
		return threeWayMergeEnabled
	}
}

func displayWarnings(userSpecifiedThreeWayMergeMode ThreeWayMergeModeType, newRelease *release.Release) {
	threeWayMergeMode := getActualThreeWayMergeMode(userSpecifiedThreeWayMergeMode)
	enabledDescExtra := ""
	disabledDescExtra := ""
	onlyNewReleasesDescExtra := ""
	switch threeWayMergeMode {
	case threeWayMergeEnabled:
		enabledDescExtra = "(CURRENT)"
	case threeWayMergeDisabled:
		disabledDescExtra = "(CURRENT)"
	case threeWayMergeOnlyNewReleases:
		onlyNewReleasesDescExtra = "(CURRENT)"
	}

	logboek.LogOptionalLn()

	releaseResourceName := fmt.Sprintf("%s/%s.v%d", strings.ToLower(tillerSettings.Releases.Name()), newRelease.GetName(), newRelease.GetVersion())

	if currentDate.After(noDisableThreeWayMergeDeadline) {
		if userSpecifiedThreeWayMergeMode != "" && userSpecifiedThreeWayMergeMode != threeWayMergeEnabled {
			logboek.LogWarnF("  WARNING Specified three-way-merge-mode \"%s\" cannot be activated anymore!", userSpecifiedThreeWayMergeMode)
			logboek.LogWarnF("  WARNING werf will always use \"enabled\" three-way-merge-mode.")
		}
	} else if userSpecifiedThreeWayMergeMode != threeWayMergeEnabled {
		logboek.Default.LogFHighlight("ATTENTION Current three-way-merge-mode for updates is \"%s\".\n", threeWayMergeMode)

		if newRelease.ThreeWayMergeEnabled {
			logboek.Default.LogFHighlight("ATTENTION Three way merge is ENABLED for the release %q.\n", newRelease.Name)
		} else {
			logboek.Default.LogFHighlight("ATTENTION Three way merge is DISABLED for the release %q.\n", newRelease.Name)
		}

		logboek.Default.LogLnHighlight(
			"ATTENTION",
			"ATTENTION Note that three-way-merge-mode does not affect resources adoption,",
			"ATTENTION resources adoption will always use three-way-merge patches.",
		)

		logboek.Default.LogLnHighlight(
			"ATTENTION",
			"ATTENTION To force werf to use specific three-way-merge mode",
			"ATTENTION and prevent auto selecting of three-way-merge-mode",
			"ATTENTION please set WERF_THREE_WAY_MERGE_MODE env var",
			"ATTENTION (or cli option --three-way-merge-mode) to one of the following values:",
			"ATTENTION  - \"enabled\"         — always use three-way-merge patches during updates",
			fmt.Sprintf("ATTENTION                        for already existing and new releases; %s", enabledDescExtra),
			"ATTENTION  - \"disabled\"        — do not use three-way-merge patches during updates",
			fmt.Sprintf("ATTENTION                        neither for already existing nor new releases; %s", disabledDescExtra),
			"ATTENTION  - \"onlyNewReleases\" — new releases created since that mode is active",
			"ATTENTION                        will use three-way-merge patches during updates,",
			"ATTENTION                        while already existing releases continue to use old",
			fmt.Sprintf("ATTENTION                        helm two-way-merge patches and repair patches approach; %s", onlyNewReleasesDescExtra),
			"ATTENTION",
		)

		if currentDate.Before(threeWayMergeOnlyNewReleasesEnabledDeadline) {
			logboek.Default.LogLnHighlight(
				fmt.Sprintf("ATTENTION Starting with %s", threeWayMergeOnlyNewReleasesEnabledDeadline),
				"ATTENTION werf will select \"onlyNewReleases\" three-way-merge-mode by default!",
				"ATTENTION",
				"ATTENTION It is strongly recommended to set three-way-merge-mode",
				"ATTENTION to \"enabled\" manually already now, unless there is a reason not to do that,",
				fmt.Sprintf("ATTENTION because starting with %s", threeWayMergeEnabledDeadline),
				"ATTENTION werf will select \"enabled\" three-way-merge-mode by default!",
			)
		} else if currentDate.Before(threeWayMergeEnabledDeadline) {
			logboek.Default.LogLnHighlight(
				fmt.Sprintf("ATTENTION Starting with %s", threeWayMergeEnabledDeadline),
				"ATTENTION werf will select \"enabled\" three-way-merge-mode by default!",
			)

			if userSpecifiedThreeWayMergeMode == threeWayMergeDisabled {
				logboek.LogWarnF("  WARNING Three-way-merge-mode is set to \"disabled\" and\n")
				logboek.LogWarnF("  WARNING should be switched to \"onlyNewReleases\" or \"enabled\"\n")
				logboek.LogWarnF("  WARNING as soon as possible, because two-way-merge will be DEPRECATED soon!\n")
			} else if userSpecifiedThreeWayMergeMode == threeWayMergeOnlyNewReleases {
				if !newRelease.ThreeWayMergeEnabled {
					logboek.LogWarnF("  WARNING Three-way-merge-mode is set to \"onlyNewReleases\" and current release\n")
					logboek.LogWarnF("  WARNING %q is old and does not use three way merge.\n", newRelease.GetName())
					logboek.LogWarnF("  WARNING Manually add annotation \"%s\": \"true\" to the %s\n", driver.ThreeWayMergeEnabledAnnotation, releaseResourceName)
					logboek.LogWarnF("  WARNING to enable three-way-merge for this release as soon as possible,\n")
					logboek.LogWarnF("  WARNING because two-way-merge will be DEPRECATED soon!\n")
				}
			}
		} else {
			if userSpecifiedThreeWayMergeMode != "" {
				logboek.LogWarnF("  WARNING Three-way-merge-mode is set to \"%s\" and\n", userSpecifiedThreeWayMergeMode)
				logboek.LogWarnF("  WARNING should be switched to \"enabled\" as soon as possible,\n")
				logboek.LogWarnF("  WARNING because two-way-merge will be DEPRECATED soon!\n")
			}
		}
	}

	if len(kube.LastClientWarnings) > 0 {
		logboek.LogWarnF("  WARNING ### Following problems detected during deploy process ###\n")

		for _, msg := range kube.LastClientWarnings {
			logboek.LogWarnF("  WARNING %s\n", msg)
		}
	}

	logboek.LogOptionalLn()
}
