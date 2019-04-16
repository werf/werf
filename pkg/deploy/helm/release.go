package helm

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

type TrackAnno string

const (
	TrackAnnoName          = "werf.io/track"
	HelmHookAnnoName       = "helm.sh/hook"
	HelmHookWeightAnnoName = "helm.sh/hook-weight"

	TrackDisabled  TrackAnno = "false"
	TrackTillDone  TrackAnno = "till_done"
	TrackTillReady TrackAnno = "till_ready"
)

func PurgeHelmRelease(releaseName, namespace string, withNamespace bool) error {
	return withLockedHelmRelease(releaseName, func() error {
		return doPurgeHelmRelease(releaseName, namespace, withNamespace)
	})
}

func doPurgeHelmRelease(releaseName, namespace string, withNamespace bool) error {
	logProcessMsg := fmt.Sprintf("Checking release %s status", releaseName)
	if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
		_, err := releaseStatus(releaseName, releaseStatusOptions{})
		if err != nil {
			if isReleaseNotFoundError(err) {
				return fmt.Errorf("helm release %s is not found", releaseName)
			}

			return fmt.Errorf("failed to check release status: %s", err)
		}

		return nil
	}); err != nil {
		return err
	}

	if err := validateHelmReleaseNamespace(releaseName, namespace); err != nil {
		return err
	}

	msg := fmt.Sprintf("Deleting helm release %s", releaseName)
	if err := logboek.LogSecondaryProcessInline(msg, func() error {
		return releaseDelete(releaseName, releaseDeleteOptions{Purge: true})
	}); err != nil {
		return fmt.Errorf("purge helm release %s failed: %s", releaseName, err)
	}

	if withNamespace {
		logProcessMsg := fmt.Sprintf("Deleting kubernetes namespace %s", namespace)
		if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
			return kube.Kubernetes.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
		}); err != nil {
			return fmt.Errorf("failed to delete namespace %s: %s", namespace, err)
		}

		logboek.OptionalLnModeOn()
	}

	return nil
}

type ChartValuesOptions struct {
	Set       []string
	SetString []string
	Values    []string
}

type ChartOptions struct {
	Timeout time.Duration

	DryRun bool
	Debug  bool

	ChartValuesOptions
}

func withLockedHelmRelease(releaseName string, f func() error) error {
	lockName := fmt.Sprintf("helm_release.%s-kube_context.%s", releaseName, helmSettings.KubeContext)
	return lock.WithLock(lockName, lock.LockOptions{}, f)
}

func DeployHelmChart(chartPath, releaseName, namespace string, opts ChartOptions) error {
	return withLockedHelmRelease(releaseName, func() error {
		return doDeployHelmChart(chartPath, releaseName, namespace, opts)
	})
}

func doDeployHelmChart(chartPath, releaseName, namespace string, opts ChartOptions) (err error) {
	var templates ChartTemplates
	var releaseState ReleaseState

	preDeployFunc := func() error {
		logProcessMsg := fmt.Sprintf("Checking release %s status", releaseName)
		if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
			releaseState, err = getReleaseState(releaseName)
			return err
		}); err != nil {
			return fmt.Errorf("getting release status failed: %s", err)
		}

		if releaseState.PurgeNeeded {
			logProcessMsg := fmt.Sprintf("Purging failed release %s", releaseName)
			if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
				return releaseDelete(releaseName, releaseDeleteOptions{Purge: true})
			}); err != nil {
				return fmt.Errorf("purge helm release %s failed: %s", releaseName, err)
			}

			releaseState.IsExists = false

			if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
				return err
			}
		}

		if releaseState.IsExists {
			if err := validateHelmReleaseNamespace(releaseName, namespace); err != nil {
				return err
			}
		}

		if releaseState.IsExists && releaseState.StatusCode == "FAILED" {
			var revision int32
			logProcessMsg := fmt.Sprintf("Getting latest deployed release %s revision", releaseName)
			if err := logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
				revision, err = latestDeployedReleaseRevision(releaseName)
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

				deployFunc := func() (string, error) {
					logboek.LogServiceF("Running helm rollback...\n")
					logboek.OptionalLnModeOn()

					releaseRollbackOpts := ReleaseRollbackOptions{
						releaseRollbackOptions: releaseRollbackOptions{
							Timeout:       int64(opts.Timeout),
							CleanupOnFail: true,
							DryRun:        opts.DryRun,
						},
					}

					out := &bytes.Buffer{}
					if err := ReleaseRollback(
						out,
						releaseName,
						revision,
						releaseRollbackOpts,
					); err != nil {
						return "", fmt.Errorf("helm release %s rollback to revision %d failed: %s", releaseName, revision, err)
					}

					return out.String(), nil
				}

				logProcessMsg = fmt.Sprintf("Running rollback release %s to revision %d", releaseName, revision)
				if err := logboek.LogSecondaryProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
					return runDeployProcess(releaseName, namespace, opts, templates, deployFunc)
				}); err != nil {
					return err
				}
			}
		}

		if err := logboek.LogSecondaryProcessInline("Getting chart templates", func() error {
			templates, err = GetTemplatesFromChart(chartPath, releaseName, namespace, opts.Values, opts.Set, opts.SetString)
			return err
		}); err != nil {
			return err
		}

		return nil
	}

	if err := logboek.LogProcess("Running pre-deploy", logboek.LogProcessOptions{}, func() error {
		return preDeployFunc()
	}); err != nil {
		return err
	}

	var deployFunc func() (string, error)
	if releaseState.IsExists {
		deployFunc = func() (string, error) {
			logboek.LogServiceF("Running helm upgrade...\n")
			logboek.OptionalLnModeOn()

			releaseUpdateOpts := ReleaseUpdateOptions{
				releaseUpdateOptions: releaseUpdateOptions{
					Timeout:       int64(opts.Timeout),
					CleanupOnFail: true,
					Wait:          true,
					DryRun:        opts.DryRun,
				},
				Debug: opts.Debug,
			}

			out := &bytes.Buffer{}
			if err := ReleaseUpdate(
				out,
				chartPath,
				releaseName,
				opts.Values,
				opts.Set,
				opts.SetString,
				releaseUpdateOpts,
			); err != nil {
				if strings.HasSuffix(err.Error(), "has no deployed releases") {
					logboek.LogErrorF("WARNING: Helm release %s is in improper state: %s\n", releaseName, err.Error())

					if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
						return "", err
					}

					logboek.LogErrorF("WARNING: Helm release %s will be removed with `helm delete --purge` on the next run of `werf deploy`\n", releaseName)
				}

				return "", fmt.Errorf("helm release %s upgrade failed: %s", releaseName, err)
			}

			if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
				return "", err
			}

			return out.String(), nil
		}
	} else {
		deployFunc = func() (string, error) {
			logboek.LogServiceF("Running helm install...\n")
			logboek.OptionalLnModeOn()

			releaseInstallOpts := ReleaseInstallOptions{
				releaseInstallOptions: releaseInstallOptions{
					Timeout: int64(opts.Timeout),
					Wait:    true,
					DryRun:  opts.DryRun,
				},
				Debug: opts.Debug,
			}

			out := &bytes.Buffer{}
			if err := ReleaseInstall(
				out,
				chartPath,
				releaseName,
				namespace,
				opts.Values,
				opts.Set,
				opts.SetString,
				releaseInstallOpts,
			); err != nil {
				if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
					return "", err
				}

				return "", fmt.Errorf("helm release %s install failed: %s", releaseName, err)
			}

			return out.String(), nil
		}
	}

	logProcessMsg := fmt.Sprintf("Running deploy release %s", releaseName)
	return logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		return runDeployProcess(releaseName, namespace, opts, templates, deployFunc)
	})
}

func latestDeployedReleaseRevision(releaseName string) (int32, error) {
	resp, err := releaseHistory(releaseName, releaseHistoryOptions{})
	if err != nil {
		return 0, fmt.Errorf("unable to get release history: %s", err)
	}

	for _, r := range resp.Releases {
		if r.Info.Status.Code.String() == "DEPLOYED" {
			return r.Version, nil
		}
	}

	return 0, ErrNoDeployedReleaseRevisionFound
}

func runDeployProcess(releaseName, namespace string, opts ChartOptions, templates ChartTemplates, deployFunc func() (string, error)) error {
	if err := removeHelmHooksByRecreatePolicy(templates, namespace); err != nil {
		return fmt.Errorf("unable to remove helm hooks by werf/recreate policy: %s", err)
	}

	helmOutput, err := deployFunc()
	if err != nil {
		return err
	}

	if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
		return err
	}

	logboek.LogInfoF(logboek.FitText(helmOutput, logboek.FitTextOptions{MaxWidth: 120}))

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

type ReleaseState struct {
	IsExists    bool
	StatusCode  string
	PurgeNeeded bool
}

func validateHelmReleaseNamespace(releaseName, namespace string) error {
	resp, err := releaseContent(releaseName, releaseContentOptions{})
	if err != nil {
		return fmt.Errorf("failed to check release namespace: %s", err)
	}

	if resp.Release.Namespace != namespace {
		return fmt.Errorf("existing helm release %s is deployed in namespace %s (not in %s): check --namespace option value", releaseName, resp.Release.Namespace, namespace)
	}

	return nil
}

func getReleaseState(releaseName string) (ReleaseState, error) {
	var releaseState ReleaseState

	code, err := releaseStatusCode(releaseName, releaseStatusCodeOptions{})
	if err != nil && !isReleaseNotFoundError(err) {
		return ReleaseState{}, err
	}

	releaseState.StatusCode = code

	if releaseState.StatusCode == "" {
		releaseState.IsExists = false
		if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
			return ReleaseState{}, err
		}
	} else if releaseState.StatusCode == "FAILED" || releaseState.StatusCode == "PENDING_INSTALL" {
		releaseState.IsExists = true

		if exist, err := util.FileExists(autoPurgeTriggerFilePath(releaseName)); err != nil {
			return ReleaseState{}, err
		} else if exist {
			releaseState.PurgeNeeded = true
		}
	} else {
		if exist, err := util.FileExists(autoPurgeTriggerFilePath(releaseName)); err != nil {
			return ReleaseState{}, err
		} else if exist {
			logboek.LogErrorF("WARNING: Will not purge helm release '%s': expected FAILED or PENDING_INSTALL release status, got %s\n", releaseName, releaseState.StatusCode)
		}

		releaseState.IsExists = true

		if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
			return ReleaseState{}, err
		}
	}

	return releaseState, nil
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
		options := metav1.GetOptions{}
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

	deletePropagation := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
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
