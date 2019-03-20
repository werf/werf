package helm

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/kubedog/pkg/tracker"
	"github.com/flant/kubedog/pkg/trackers/rollout"
	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TrackAnno string

const (
	DefaultHelmTimeout = 24 * time.Hour

	TrackAnnoName          = "werf.io/track"
	HelmHookAnnoName       = "helm.sh/hook"
	HelmHookWeightAnnoName = "helm.sh/hook-weight"

	TrackDisabled  TrackAnno = "false"
	TrackTillDone  TrackAnno = "till_done"
	TrackTillReady TrackAnno = "till_ready"
)

var (
	ErrNoDeployedReleaseRevisionFound = errors.New("no DEPLOYED release revision found")
)

func PurgeHelmRelease(releaseName string) error {
	return withLockedHelmRelease(releaseName, func() error {
		return doPurgeHelmRelease(releaseName)
	})
}

func doPurgeHelmRelease(releaseName string) error {
	logProcessMsg := fmt.Sprintf("Checking release %s status", releaseName)
	if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
		stdout, stderr, err := HelmCmd("status", releaseName)
		if err != nil {
			if strings.HasSuffix(stderr, "not found") {
				return fmt.Errorf("helm release %s doesn't exist", releaseName)
			}

			return fmt.Errorf("failed to check release status: %s", FormatHelmCmdError(stdout, stderr, err))
		}

		return nil
	}); err != nil {
		return err
	}

	return logboek.LogSecondaryProcessInline("Running helm purge command", func() error {
		if err := purgeRelease(releaseName); err != nil {
			return fmt.Errorf("purge helm release %s failed: %s", releaseName, err)
		}

		return nil
	})
}

type HelmChartValuesOptions struct {
	Set       []string
	SetString []string
	Values    []string
}

type HelmChartOptions struct {
	Timeout time.Duration

	DryRun bool
	Debug  bool

	HelmChartValuesOptions
}

func withLockedHelmRelease(releaseName string, f func() error) error {
	lockName := fmt.Sprintf("helm_release.%s", releaseName)
	return lock.WithLock(lockName, lock.LockOptions{}, f)
}

func DeployHelmChart(chartPath, releaseName, namespace string, opts HelmChartOptions) error {
	return withLockedHelmRelease(releaseName, func() error {
		return doDeployHelmChart(chartPath, releaseName, namespace, opts)
	})
}

func doDeployHelmChart(chartPath, releaseName, namespace string, opts HelmChartOptions) (err error) {
	var templates ChartTemplates
	var releaseStatus ReleaseStatus

	preDeployFunc := func() error {
		logProcessMsg := fmt.Sprintf("Checking release %s status", releaseName)
		if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
			releaseStatus, err = getReleaseStatus(releaseName)
			return err
		}); err != nil {
			return fmt.Errorf("getting release status failed: %s", err)
		}

		if releaseStatus.PurgeNeeded {
			logProcessMsg := fmt.Sprintf("Purging failed release %s", releaseName)
			if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
				return purgeRelease(releaseName)
			}); err != nil {
				return fmt.Errorf("purge helm release %s failed: %s", releaseName, err)
			}

			releaseStatus.IsExists = false

			if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
				return err
			}
		}

		if releaseStatus.IsExists && releaseStatus.Status == "FAILED" {
			var revision int
			logProcessMsg := fmt.Sprintf("Getting latest deployed release %s revision", releaseName)
			if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
				revision, err = getLatestDeployedReleaseRevision(releaseName)
				if err != nil && err != ErrNoDeployedReleaseRevisionFound {
					return err
				}

				return nil
			}); err != nil {
				return fmt.Errorf("unable to get latest deployed revision of release %s: %s", releaseName, err)
			}

			if err != ErrNoDeployedReleaseRevisionFound {
				logProcessMsg := fmt.Sprintf("Getting templates from release %s revision %d", releaseName, revision)
				if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
					templates, err = GetTemplatesFromRevision(releaseName, revision)
					return err
				}); err != nil {
					return fmt.Errorf("unable to get helm templates from release %s revision %d: %s", releaseName, revision, err)
				}

				watchHookTypes := []string{"pre-rollback", "post-rollback"}

				deployFunc := func(startJobHooksWatcher chan bool) (string, error) {
					logboek.LogServiceF("Running helm rollback command...\n")
					logboek.OptionalLnModeOn()

					startJobHooksWatcher <- true

					output, err := rollbackRelease(releaseName, revision, namespace, opts)
					if err != nil {
						return "", fmt.Errorf("helm release %s rollback to revision %d failed: %s", releaseName, revision, err)
					}

					return output, nil
				}

				logProcessMsg = fmt.Sprintf("Running rollback release %s to revision %d", releaseName, revision)
				if err := logboek.LogSecondaryProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
					return runDeployProcess(releaseName, namespace, opts, templates, watchHookTypes, deployFunc)
				}); err != nil {
					return err
				}
			}
		}

		if err := logboek.LogSecondaryProcessInline("Getting chart templates", func() error {
			templates, err = GetTemplatesFromChart(chartPath, releaseName, opts.Set, opts.SetString, opts.Values)
			return err
		}); err != nil {
			return fmt.Errorf("unable to get templates of chart %s: %s", chartPath, err)
		}

		return nil
	}

	if err := logboek.LogProcess("Running pre-deploy", logboek.LogProcessOptions{}, func() error {
		return preDeployFunc()
	}); err != nil {
		return err
	}

	var watchHookTypes []string
	if releaseStatus.IsExists {
		watchHookTypes = []string{"pre-upgrade", "post-upgrade"}
	} else {
		watchHookTypes = []string{"pre-install", "post-install"}
	}

	var deployFunc func(chan bool) (string, error)
	if releaseStatus.IsExists {
		deployFunc = func(startJobHooksWatcher chan bool) (string, error) {
			logboek.LogServiceF("Running helm upgrade command...\n")
			logboek.OptionalLnModeOn()

			startJobHooksWatcher <- true

			output, err := upgradeRelease(chartPath, releaseName, namespace, opts)
			if err != nil {
				return "", fmt.Errorf("helm release %s upgrade failed: %s", releaseName, err)
			}

			if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
				return "", err
			}

			return output, nil
		}
	} else {
		deployFunc = func(startJobHooksWatcher chan bool) (string, error) {
			logboek.LogServiceF("Running helm install command...\n")
			logboek.OptionalLnModeOn()

			startJobHooksWatcher <- true

			output, err := installRelease(chartPath, releaseName, namespace, opts)
			if err != nil {
				return "", fmt.Errorf("helm release %s install failed: %s", releaseName, err)
			}

			return output, nil
		}
	}

	logProcessMsg := fmt.Sprintf("Running deploy release %s", releaseName)
	return logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		return runDeployProcess(releaseName, namespace, opts, templates, watchHookTypes, deployFunc)
	})
}

func runDeployProcess(releaseName, namespace string, opts HelmChartOptions, templates ChartTemplates, watchHookTypes []string, deployFunc func(chan bool) (string, error)) error {
	if err := removeHelmHooksByRecreatePolicy(templates, namespace); err != nil {
		return fmt.Errorf("unable to remove helm hooks by werf/recreate policy: %s", err)
	}

	deployStartTime := time.Now()

	jobHooksWatcherDone, startJobHooksWatcher, err := watchJobHooks(templates, watchHookTypes, deployStartTime, namespace, opts)
	if err != nil {
		return fmt.Errorf("watching job hooks failed: %s", err)
	}

	helmOutput, err := deployFunc(startJobHooksWatcher)
	if err != nil {
		return err
	}

	if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
		return err
	}

	<-jobHooksWatcherDone

	logboek.LogInfoLn(logboek.FitText(helmOutput, logboek.FitTextOptions{MaxWidth: 120}))
	logboek.OptionalLnModeOn()

	if err := trackPods(templates, deployStartTime, namespace, opts); err != nil {
		return err
	}
	if err := trackDeployments(templates, deployStartTime, namespace, opts); err != nil {
		return err
	}
	if err := trackStatefulSets(templates, deployStartTime, namespace, opts); err != nil {
		return err
	}
	if err := trackDaemonSets(templates, deployStartTime, namespace, opts); err != nil {
		return err
	}
	if err := trackJobs(templates, deployStartTime, namespace, opts); err != nil {
		return err
	}

	return nil
}

func purgeRelease(releaseName string) error {
	if stdout, stderr, err := HelmCmd("delete", "--purge", releaseName); err != nil {
		return FormatHelmCmdError(stdout, stderr, err)
	}

	return nil
}

func rollbackRelease(releaseName string, revision int, namespace string, chartOpts HelmChartOptions) (string, error) {
	args := commonDeployHelmCommandArgs(chartOpts)
	args = append([]string{"rollback", releaseName, fmt.Sprintf("%d", revision)}, args...)

	stdout, stderr, err := HelmCmd(args...)
	if err != nil {
		return "", FormatHelmCmdError(stdout, stderr, err)
	}

	return FormatHelmCmdOutput(stdout, stderr), nil
}

func installRelease(chartPath, releaseName, namespace string, chartOpts HelmChartOptions) (string, error) {
	args := commonDeployHelmCommandArgs(chartOpts)
	args = append(args, commonInstallOrUpgradeHelmCommandArgs(namespace, chartOpts)...)
	args = append([]string{"install", chartPath, "--name", releaseName}, args...)

	stdout, stderr, err := HelmCmd(args...)
	if err != nil {
		if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
			return "", err
		}

		return "", FormatHelmCmdError(stdout, stderr, err)
	}

	return FormatHelmCmdOutput(stdout, stderr), nil
}

func upgradeRelease(chartPath, releaseName, namespace string, chartOpts HelmChartOptions) (string, error) {
	args := commonDeployHelmCommandArgs(chartOpts)
	args = append(args, commonInstallOrUpgradeHelmCommandArgs(namespace, chartOpts)...)
	args = append([]string{"upgrade", releaseName, chartPath}, args...)

	stdout, stderr, err := HelmCmd(args...)
	if err != nil {
		if strings.HasSuffix(stderr, "has no deployed releases") {
			logboek.LogErrorF("WARNING: Helm release '%s' is in improper state: %s\n", releaseName, stderr)

			if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
				return "", err
			}

			logboek.LogErrorF("WARNING: Helm release %s will be removed with `helm delete --purge` on the next run of `werf deploy`\n", releaseName)
		}

		return "", FormatHelmCmdError(stdout, stderr, err)
	}

	return FormatHelmCmdOutput(stdout, stderr), nil
}

func trackPods(templates ChartTemplates, deployStartTime time.Time, namespace string, opts HelmChartOptions) error {
	for _, template := range templates.Pods() {
		if _, ok := template.Metadata.Annotations[HelmHookAnnoName]; ok {
			continue
		}

		if template.Metadata.Annotations[TrackAnnoName] == string(TrackDisabled) {
			continue
		}

		logProcessMsg := fmt.Sprintf("Tracking po/%s", template.Metadata.Name)
		if err := logboek.LogSecondaryProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
			return rollout.TrackPodTillReady(template.Metadata.Name, template.Namespace(namespace), kube.Kubernetes, tracker.Options{Timeout: time.Second * time.Duration(opts.Timeout), LogsFromTime: deployStartTime})
		}); err != nil {
			return err
		}
	}

	return nil
}

func removeHelmHooksByRecreatePolicy(templates ChartTemplates, namespace string) error {
	for _, jobTemplate := range getHooksJobsToRecreate(templates.Jobs()) {
		logProcessMsg := fmt.Sprintf("Deleting helm hook jobs/%s (werf/recreate)", jobTemplate.Metadata.Name)
		if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
			return removeJob(jobTemplate.Metadata.Name, jobTemplate.Namespace(namespace))
		}); err != nil {
			return fmt.Errorf("unable to remove job '%s': %s", jobTemplate.Metadata.Name, err)
		}
	}

	return nil
}

func trackDeployments(templates ChartTemplates, deployStartTime time.Time, namespace string, opts HelmChartOptions) error {
	for _, template := range templates.Deployments() {
		if _, ok := template.Metadata.Annotations[HelmHookAnnoName]; ok {
			continue
		}

		if template.Metadata.Annotations[TrackAnnoName] == string(TrackDisabled) {
			continue
		}

		logProcessMsg := fmt.Sprintf("Tracking deploy/%s", template.Metadata.Name)
		if err := logboek.LogSecondaryProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
			return rollout.TrackDeploymentTillReady(template.Metadata.Name, template.Namespace(namespace), kube.Kubernetes, tracker.Options{Timeout: opts.Timeout, LogsFromTime: deployStartTime})
		}); err != nil {
			return err
		}
	}

	return nil
}

func trackStatefulSets(templates ChartTemplates, deployStartTime time.Time, namespace string, opts HelmChartOptions) error {
	for _, template := range templates.StatefulSets() {
		if _, ok := template.Metadata.Annotations[HelmHookAnnoName]; ok {
			continue
		}

		if template.Metadata.Annotations[TrackAnnoName] == string(TrackDisabled) {
			continue
		}

		logProcessMsg := fmt.Sprintf("Tracking sts/%s", template.Metadata.Name)
		if err := logboek.LogSecondaryProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
			return rollout.TrackStatefulSetTillReady(template.Metadata.Name, template.Namespace(namespace), kube.Kubernetes, tracker.Options{Timeout: time.Second * time.Duration(opts.Timeout), LogsFromTime: deployStartTime})
		}); err != nil {
			return err
		}
	}

	return nil
}

func trackDaemonSets(templates ChartTemplates, deployStartTime time.Time, namespace string, opts HelmChartOptions) error {
	for _, template := range templates.DaemonSets() {
		if _, ok := template.Metadata.Annotations[HelmHookAnnoName]; ok {
			continue
		}

		if template.Metadata.Annotations[TrackAnnoName] == string(TrackDisabled) {
			continue
		}

		logProcessMsg := fmt.Sprintf("Tracking ds/%s", template.Metadata.Name)
		if err := logboek.LogSecondaryProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
			return rollout.TrackDaemonSetTillReady(template.Metadata.Name, template.Namespace(namespace), kube.Kubernetes, tracker.Options{Timeout: time.Second * time.Duration(opts.Timeout), LogsFromTime: deployStartTime})
		}); err != nil {
			return err
		}
	}

	return nil
}

func trackJobs(templates ChartTemplates, deployStartTime time.Time, namespace string, opts HelmChartOptions) error {
	for _, template := range templates.Jobs() {
		if _, ok := template.Metadata.Annotations[HelmHookAnnoName]; ok {
			continue
		}

		if template.Metadata.Annotations[TrackAnnoName] == string(TrackTillDone) {
			logProcessMsg := fmt.Sprintf("Tracking jobs/%s", template.Metadata.Name)
			if err := logboek.LogSecondaryProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
				return rollout.TrackJobTillDone(template.Metadata.Name, template.Namespace(namespace), kube.Kubernetes, tracker.Options{Timeout: time.Second * time.Duration(opts.Timeout), LogsFromTime: deployStartTime})
			}); err != nil {
				return err
			}
		} else {
			// TODO: https://github.com/flant/werf/issues/1143
			// till_ready by default
			// if werf.io/track=false -- no track at all
		}
	}

	return nil
}

func watchJobHooks(templates ChartTemplates, hookTypes []string, deployStartTime time.Time, namespace string, opts HelmChartOptions) (chan bool, chan bool, error) {
	jobHooksWatcherDone := make(chan bool)
	startJobHooksWatcher := make(chan bool)

	jobHooksToTrack, err := jobHooksToTrack(templates, hookTypes)
	if err != nil {
		return jobHooksWatcherDone, startJobHooksWatcher, err
	}

	go func() {
		<-startJobHooksWatcher

		for _, template := range jobHooksToTrack {
			var jobNamespace string
			if template.Metadata.Namespace != "" {
				jobNamespace = template.Metadata.Namespace
			} else {
				jobNamespace = namespace
			}

			loggerProcessMsg := fmt.Sprintf("Tracking helm hook jobs/%s", template.Metadata.Name)
			if err := logboek.LogSecondaryProcess(loggerProcessMsg, logboek.LogProcessOptions{}, func() error {
				return rollout.TrackJobTillDone(template.Metadata.Name, jobNamespace, kube.Kubernetes, tracker.Options{Timeout: opts.Timeout, LogsFromTime: deployStartTime})
			}); err != nil {
				logboek.LogErrorF("ERROR: %s\n", err)
				break
			}
		}

		jobHooksWatcherDone <- true
	}()

	return jobHooksWatcherDone, startJobHooksWatcher, nil
}

func commonInstallOrUpgradeHelmCommandArgs(namespace string, opts HelmChartOptions) []string {
	var args []string

	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	for _, set := range opts.Set {
		args = append(args, "--set", set)
	}

	for _, setString := range opts.SetString {
		args = append(args, "--set-string", setString)
	}

	for _, values := range opts.Values {
		args = append(args, "--values", values)
	}

	return args
}

func commonDeployHelmCommandArgs(opts HelmChartOptions) []string {
	var args []string

	if opts.DryRun {
		args = append(args, "--dry-run")
		args = append(args, "--debug")
	}

	if opts.Timeout != 0 {
		args = append(args, "--timeout", fmt.Sprintf("%v", opts.Timeout.Seconds()))
	} else {
		args = append(args, "--timeout", fmt.Sprintf("%v", DefaultHelmTimeout.Seconds()))
	}

	return args
}

type ReleaseStatus struct {
	IsExists    bool
	Status      string
	PurgeNeeded bool
}

func getReleaseStatus(releaseName string) (ReleaseStatus, error) {
	var res ReleaseStatus

	helmStatusStdout, _, helmStatusErr := HelmCmd("status", releaseName)
	if helmStatusErr == nil {
		statusLinePrefix := "STATUS: "
		scanner := bufio.NewScanner(strings.NewReader(helmStatusStdout))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, statusLinePrefix) {
				res.Status = line[len(statusLinePrefix):]
				break
			}
		}
	}

	if helmStatusErr != nil {
		res.IsExists = false
		if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
			return ReleaseStatus{}, err
		}
	} else if res.Status != "" && (res.Status == "FAILED" || res.Status == "PENDING_INSTALL") {
		res.IsExists = true

		if exist, err := util.FileExists(autoPurgeTriggerFilePath(releaseName)); err != nil {
			return ReleaseStatus{}, err
		} else if exist {
			res.PurgeNeeded = true
		}
	} else {
		if exist, err := util.FileExists(autoPurgeTriggerFilePath(releaseName)); err != nil {
			return ReleaseStatus{}, err
		} else if exist {
			logboek.LogErrorF("WARNING: Will not purge helm release '%s': expected FAILED or PENDING_INSTALL release status, got %s\n", releaseName, res.Status)
		}

		res.IsExists = true

		if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
			return ReleaseStatus{}, err
		}
	}

	return res, nil
}

func getLatestDeployedReleaseRevision(releaseName string) (int, error) {
	history, err := GetReleaseHistory(releaseName)
	if err != nil {
		return 0, fmt.Errorf("unable to get release history: %s", err)
	}

	var reversedHistory ReleaseHistory
	for _, historyRec := range history {
		reversedHistory = append(ReleaseHistory{historyRec}, reversedHistory...)
	}

	for _, historyRec := range reversedHistory {
		if historyRec.Status == "DEPLOYED" {
			return historyRec.Revision, nil
		}
	}

	return 0, ErrNoDeployedReleaseRevisionFound
}

func getHooksJobsToRecreate(jobsTemplates []Template) []Template {
	var res []Template

	for _, template := range jobsTemplates {
		if _, isHelmHook := template.Metadata.Annotations[HelmHookAnnoName]; !isHelmHook {
			continue
		}

		value, ok := template.Metadata.Annotations["werf/recreate"]
		if ok && (value == "0" || value == "false") {
			continue
		}

		res = append(res, template)
	}

	return res
}

func removeJob(jobName string, namespace string) error {
	isJobExist := func(name string, namespace string) (bool, error) {
		options := v1.GetOptions{}
		_, err := kube.Kubernetes.BatchV1().Jobs(namespace).Get(name, options)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}

			return true, err
		}

		return true, nil
	}

	exist, err := isJobExist(jobName, namespace)
	if err != nil {
		return err
	} else if !exist {
		return nil
	}

	deletePropagation := v1.DeletePropagationForeground
	deleteOptions := &v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	}
	err = kube.Kubernetes.BatchV1().Jobs(namespace).Delete(jobName, deleteOptions)
	if err != nil {
		return err
	}

	for {
		exist, err := isJobExist(jobName, namespace)
		if err != nil {
			return err
		} else if !exist {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func jobHooksToTrack(templates ChartTemplates, hookTypes []string) ([]Template, error) {
	var jobHooksToTrack []Template
	jobHooksByType := make(map[string][]Template)

	for _, template := range templates.Jobs() {
		if anno, ok := template.Metadata.Annotations[HelmHookAnnoName]; ok {
			if template.Metadata.Annotations[TrackAnnoName] == string(TrackDisabled) {
				continue
			}

			for _, hookType := range strings.Split(anno, ",") {
				if _, ok := jobHooksByType[hookType]; !ok {
					jobHooksByType[hookType] = []Template{}
				}
				jobHooksByType[hookType] = append(jobHooksByType[hookType], template)
			}
		}
	}

	for _, templates := range jobHooksByType {
		sort.Slice(templates, func(i, j int) bool {
			toWeight := func(t Template) int {
				val, ok := t.Metadata.Annotations[HelmHookWeightAnnoName]
				if !ok {
					return 0
				}

				i, err := strconv.Atoi(val)
				if err != nil {
					logboek.LogErrorF("WARNING: Incorrect hook-weight anno value '%v'\n", val)
					return 0
				}

				return i
			}

			return toWeight(templates[i]) < toWeight(templates[j])
		})
	}

	for _, hookType := range hookTypes {
		if templates, ok := jobHooksByType[hookType]; ok {
			jobHooksToTrack = append(jobHooksToTrack, templates...)
		}
	}

	return jobHooksToTrack, nil
}

func createAutoPurgeTriggerFilePath(releaseName string) error {
	filePath := autoPurgeTriggerFilePath(releaseName)
	dirPath := path.Dir(filePath)

	if fileExist, err := util.FileExists(filePath); err != nil {
		return err
	} else if !fileExist {
		if dirExist, err := util.FileExists(dirPath); err != nil {
			return err
		} else if !dirExist {
			if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
				return err
			}
		}

		if _, err := os.Create(filePath); err != nil {
			return err
		}
	}

	return nil
}

func deleteAutoPurgeTriggerFilePath(releaseName string) error {
	filePath := autoPurgeTriggerFilePath(releaseName)
	if fileExist, err := util.FileExists(filePath); err != nil {
		return err
	} else if fileExist {
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}

	return nil
}

func autoPurgeTriggerFilePath(releaseName string) string {
	return filepath.Join(werf.GetServiceDir(), "helm", releaseName, "auto_purge_failed_release_on_next_deploy")
}
