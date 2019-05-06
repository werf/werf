package helm

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
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
	TrackAnnoName                    = "werf.io/track"
	FailModeAnnoName                 = "werf.io/fail-mode"
	AllowFailuresCountAnnoName       = "werf.io/allow-failures-count"
	LogWatchRegexAnnoName            = "werf.io/log-watch-regex"
	ContainerLogWatchRegexAnnoPrefix = "werf.io/log-watch-regex-for-"
	ShowLogsUntilAnnoName            = "werf.io/show-logs-until"
	SkipLogsForContainersAnnoName    = "werf.io/skip-logs-for-containers"
	ShowLogsOnlyForContainers        = "werf.io/show-logs-only-for-containers"

	TrackAnnoEnabledValue  TrackAnno = "true"
	TrackAnnoDisabledValue TrackAnno = "false"

	HelmHookAnnoName       = "helm.sh/hook"
	HelmHookWeightAnnoName = "helm.sh/hook-weight"
)

func PurgeHelmRelease(releaseName, namespace string, withNamespace, withHooks bool) error {
	return withLockedHelmRelease(releaseName, func() error {
		return doPurgeHelmRelease(releaseName, namespace, withNamespace, withHooks)
	})
}

func doPurgeHelmRelease(releaseName, namespace string, withNamespace, withHooks bool) error {
	logProcessMsg := fmt.Sprintf("Checking release %s status", releaseName)
	if err := logboek.LogSecondaryProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
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

	if withHooks {
		resp, err := releaseHistory(releaseName, releaseHistoryOptions{Max: 1})
		if err != nil {
			return err
		}

		resp, err = releaseHistory(releaseName, releaseHistoryOptions{Max: resp.Releases[0].Version})
		if err != nil {
			return err
		}

		deletedHooks := map[string]bool{}
		msg := fmt.Sprintf("Deleting helm hooks getting from existing release %s revisions (%d)", releaseName, len(resp.Releases))
		if err := logboek.LogSecondaryProcess(msg, logboek.LogProcessOptions{}, func() error {
			for _, rev := range resp.Releases {
				revHooksToDelete := map[string]Template{}
				for _, h := range rev.Hooks {
					t, err := parseTemplate(h.Manifest)
					if err != nil {
						logboek.LogErrorF("WARNING: Parsing helm hook %s manifest failed: %s", h.Name, err)
						continue
					}

					hookName := t.Metadata.Name
					hookNamespace := t.Namespace(rev.Namespace)
					hookId := hookName + hookNamespace
					if _, exist := deletedHooks[hookId]; !exist {
						revHooksToDelete[hookId] = t
					}
				}

				if len(revHooksToDelete) != 0 {
					msg := fmt.Sprintf("Processing helm hooks getting from revision %d", rev.Version)
					_ = logboek.LogSecondaryProcess(msg, logboek.LogProcessOptions{}, func() error {
						for hookId, hookTemplate := range revHooksToDelete {
							deletedHooks[hookId] = true

							if err := removeReleaseNamespacedResource(hookTemplate, rev.Namespace); err != nil {
								logboek.LogErrorF("WARNING: Deleting helm hook %s %s failed: %s", hookTemplate.Metadata.Name, err)
							}
						}

						return nil
					})
				}
			}

			return nil
		}); err != nil {
			return err
		}
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
		if err := logboek.LogSecondaryProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
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
					releaseRollbackOpts := ReleaseRollbackOptions{
						releaseRollbackOptions: releaseRollbackOptions{
							Timeout:       int64(opts.Timeout / time.Second),
							CleanupOnFail: true,
							DryRun:        opts.DryRun,
						},
					}

					var err error

					for i := 0; i < 5; i++ {
						out := &bytes.Buffer{}

						logboek.LogServiceF("Running helm rollback (%d try)...\n", i+1)

						err = ReleaseRollback(
							out,
							releaseName,
							revision,
							releaseRollbackOpts,
						)

						if err == nil {
							return out.String(), nil
						}
					}

					if err != nil {
						return "", fmt.Errorf("helm release %s rollback to revision %d have failed: %s", releaseName, revision, err)
					}

					panic("unexpected")
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

			releaseUpdateOpts := ReleaseUpdateOptions{
				releaseUpdateOptions: releaseUpdateOptions{
					Timeout:       int64(opts.Timeout / time.Second),
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

			releaseInstallOpts := ReleaseInstallOptions{
				releaseInstallOptions: releaseInstallOptions{
					Timeout: int64(opts.Timeout / time.Second),
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

func runDeployProcess(releaseName, namespace string, _ ChartOptions, templates ChartTemplates, deployFunc func() (string, error)) error {
	oldLogsFromTime := resourcesWaiter.LogsFromTime
	resourcesWaiter.LogsFromTime = time.Now()
	defer func() {
		resourcesWaiter.LogsFromTime = oldLogsFromTime
	}()

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
	jobsToDelete := getHooksJobsToRecreate(templates.Jobs())
	if len(jobsToDelete) != 0 {
		if err := logboek.LogSecondaryProcess("Applying helm hooks recreation policy (werf.io/recreate annotation)", logboek.LogProcessOptions{}, func() error {
			for _, jobTemplate := range jobsToDelete {
				if err := removeReleaseNamespacedResource(jobTemplate, namespace); err != nil {
					return fmt.Errorf("unable to remove job '%s': %s", jobTemplate.Metadata.Name, err)
				}
			}

			return nil
		}); err != nil {
			return err
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
	} else if releaseState.StatusCode == "FAILED" || releaseState.StatusCode == "PENDING_INSTALL" || releaseState.StatusCode == "DELETING" {
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
			logboek.LogErrorF("WARNING: Will not purge helm release %s: expected FAILED, DELETING or PENDING_INSTALL release status, got %s\n", releaseName, releaseState.StatusCode)
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

func removeReleaseNamespacedResource(template Template, releaseNamespace string) error {
	resourceName := template.Metadata.Name
	resourceKing := template.Kind
	return removeNamespacedResource(resourceName, resourceKing, template.Namespace(releaseNamespace))
}

func removeNamespacedResource(name, kind, namespace string) error {
	groupVersionResource, err := kube.GroupVersionResourceByKind(kind)
	if err != nil {
		return err
	}

	res := kube.DynamicClient.Resource(groupVersionResource).Namespace(namespace)

	isExist := func() (bool, error) {
		_, err := res.Get(name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}

			return true, err
		}

		return true, nil
	}

	exist, err := isExist()
	if err != nil {
		return err
	} else if !exist {
		return nil
	}

	logProcessMsg := fmt.Sprintf("Deleting %s/%s from namespace %s", groupVersionResource.Resource, name, namespace)
	return logboek.LogSecondaryProcessInline(logProcessMsg, func() error {
		deletePropagation := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &deletePropagation,
		}
		err = res.Delete(name, deleteOptions)
		if err != nil {
			return err
		}

		for {
			exist, err := isExist()
			if err != nil {
				return err
			} else if !exist {
				break
			}

			time.Sleep(500 * time.Millisecond)
		}

		return nil
	})
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
