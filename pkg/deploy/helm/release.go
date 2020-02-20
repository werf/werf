package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	helm_kube "k8s.io/helm/pkg/kube"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"
	"github.com/flant/shluz"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

const (
	TrackTerminationModeAnnoName = "werf.io/track-termination-mode"

	FailModeAnnoName                  = "werf.io/fail-mode"
	FailuresAllowedPerReplicaAnnoName = "werf.io/failures-allowed-per-replica"

	LogRegexAnnoName      = "werf.io/log-regex"
	LogRegexForAnnoPrefix = "werf.io/log-regex-for-"

	SkipLogsAnnoName              = "werf.io/skip-logs"
	SkipLogsForContainersAnnoName = "werf.io/skip-logs-for-containers"
	ShowLogsOnlyForContainers     = "werf.io/show-logs-only-for-containers"
	ShowLogsUntilAnnoName         = "werf.io/show-logs-until"

	ShowEventsAnnoName = "werf.io/show-service-messages"

	HelmHookAnnoName = "helm.sh/hook"
)

var (
	werfAnnoList = []string{
		TrackTerminationModeAnnoName,
		FailModeAnnoName,
		FailuresAllowedPerReplicaAnnoName,
		LogRegexAnnoName,
		SkipLogsAnnoName,
		SkipLogsForContainersAnnoName,
		ShowLogsOnlyForContainers,
		ShowLogsUntilAnnoName,
		ShowEventsAnnoName,
		helm_kube.SetReplicasOnlyOnCreationAnnotation,
		helm_kube.SetResourcesOnlyOnCreationAnnotation,
	}

	werfAnnoPrefixList = []string{
		LogRegexForAnnoPrefix,
	}
)

func PurgeHelmRelease(releaseName, namespace string, withNamespace, withHooks bool) error {
	return withLockedHelmRelease(releaseName, func() error {
		return doPurgeHelmRelease(releaseName, namespace, withNamespace, withHooks)
	})
}

func doPurgeHelmRelease(releaseName, namespace string, withNamespace, withHooks bool) error {
	if err := logboek.LogProcess("Checking release existence", logboek.LogProcessOptions{}, func() error {
		_, err := releaseStatus(releaseName, releaseStatusOptions{})
		if err != nil {
			if isReleaseNotFoundError(err) {
				return fmt.Errorf("release %s is not found", releaseName)
			}

			return fmt.Errorf("release status failed: %s", err)
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
		msg := fmt.Sprintf("Deleting helm hooks from all existing release revisions (%d)", len(resp.Releases))
		if err := logboek.LogProcess(msg, logboek.LogProcessOptions{}, func() error {
			for _, rev := range resp.Releases {
				revHooksToDelete := map[string]Template{}
				for _, h := range rev.Hooks {
					t, err := parseTemplate(h.Manifest)
					if err != nil {
						logboek.LogWarnF("WARNING: Parsing helm hook %s manifest failed: %s", h.Name, err)
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
					msg := fmt.Sprintf("Processing release %s revision %d", releaseName, rev.Version)
					_ = logboek.LogProcess(msg, logboek.LogProcessOptions{}, func() error {
						for hookId, hookTemplate := range revHooksToDelete {
							deletedHooks[hookId] = true

							if err := removeReleaseNamespacedResource(hookTemplate, rev.Namespace); err != nil {
								logboek.LogWarnF("WARNING: Failed to delete helm hook %s: %s", hookTemplate.Metadata.Name, err)
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

	if err := logboek.LogProcess("Deleting release", logboek.LogProcessOptions{}, func() error {
		return releaseDelete(releaseName, releaseDeleteOptions{Purge: true})
	}); err != nil {
		return fmt.Errorf("release delete failed: %s", err)
	}

	if withNamespace {
		if err := removeResource(namespace, "Namespace", ""); err != nil {
			return fmt.Errorf("delete namespace %s failed: %s", namespace, err)
		}
		logboek.LogOptionalLn()
	}

	return nil
}

type ChartValuesOptions struct {
	SecretValues []map[string]interface{}
	Set          []string
	SetString    []string
	Values       []string
}

type ChartOptions struct {
	Timeout time.Duration

	DryRun            bool
	Debug             bool
	ThreeWayMergeMode ThreeWayMergeModeType

	ChartValuesOptions
}

func withLockedHelmRelease(releaseName string, f func() error) error {
	lockName := fmt.Sprintf("helm_release.%s-kube_context.%s", releaseName, helmSettings.KubeContext)
	return shluz.WithLock(lockName, shluz.LockOptions{}, f)
}

func DeployHelmChart(chartPath, releaseName, namespace string, opts ChartOptions) error {
	return withLockedHelmRelease(releaseName, func() error {
		return doDeployHelmChart(chartPath, releaseName, namespace, opts)
	})
}

func doDeployHelmChart(chartPath, releaseName, namespace string, opts ChartOptions) (err error) {
	var isReleaseExists bool

	preDeployFunc := func() error {
		var latestReleaseRevision int32
		var latestReleaseRevisionStatus string
		var releaseShouldBeDeleted bool
		var releaseShouldBeRolledBack bool
		var latestReleaseThreeWayMergeEnabled bool

		logProcessOptions := logboek.LevelLogProcessOptions{
			SuccessInfoSectionFunc: func() {
				if isReleaseExists {
					logboek.Default.LogFDetails("revision: %d\n", latestReleaseRevision)
					logboek.Default.LogFDetails("revision-status: %s\n", latestReleaseRevisionStatus)

					if releaseShouldBeDeleted {
						logboek.LogLn()

						logboek.Default.LogLnDetails(
							"Release will be deleted:",
							"* the latest release revision might be in an inconsistent state, and",
							"* auto purge trigger file is exists.",
						)
					} else if releaseShouldBeRolledBack {
						logboek.LogLn()
						logboek.Default.LogLnDetails(
							"Release should be rolled back to the latest successfully deployed revision:",
							"* the latest release revision might be in an inconsistent state.",
						)
					}
				} else {
					logboek.Default.LogLnDetails("Release has not been deployed yet")
				}
			},
		}
		if err := logboek.Default.LogProcess("Checking release", logProcessOptions, func() error {
			resp, err := releaseHistory(releaseName, releaseHistoryOptions{Max: 1})
			if err != nil && !isReleaseNotFoundError(err) {
				return fmt.Errorf("get release history failed: %s", err)
			}

			if resp == nil {
				latestReleaseRevisionStatus = ""
			} else {
				latestReleaseRevisionNamespace := resp.Releases[0].Namespace
				if latestReleaseRevisionNamespace != namespace {
					return fmt.Errorf("existing release has been deployed in namespace %s (not in specified %s): check --namespace option value", latestReleaseRevisionNamespace, namespace)
				}

				latestReleaseRevision = resp.Releases[0].Version
				latestReleaseRevisionStatus = resp.Releases[0].Info.Status.Code.String()
				latestReleaseThreeWayMergeEnabled = resp.Releases[0].ThreeWayMergeEnabled
			}

			switch latestReleaseRevisionStatus {
			case "":
				isReleaseExists = false
				if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
					return fmt.Errorf("create auto purge trigger file failed: %s", err)
				}
			case "FAILED", "PENDING_INSTALL", "PENDING_UPGRADE", "DELETING":
				isReleaseExists = true

				exist, err := util.FileExists(autoPurgeTriggerFilePath(releaseName))
				if err != nil {
					return fmt.Errorf("file exists failed: %s", err)
				}

				if exist {
					releaseShouldBeDeleted = true
				} else if latestReleaseRevisionStatus == "FAILED" {
					threeWayMergeMode := getActualThreeWayMergeMode(opts.ThreeWayMergeMode)

					if threeWayMergeMode == threeWayMergeDisabled {
						releaseShouldBeRolledBack = true
					} else if threeWayMergeMode == threeWayMergeOnlyNewReleases && !latestReleaseThreeWayMergeEnabled {
						releaseShouldBeRolledBack = true
					}
				}
			default:
				if exist, err := util.FileExists(autoPurgeTriggerFilePath(releaseName)); err != nil {
					return err
				} else if exist {
					logboek.LogWarnF("WARNING: Improper state:\n")
					logboek.LogWarnF("* auto purge trigger file is exists, and\n")
					logboek.LogWarnF("* the latest release revision (%s) should not be deleted.\n", latestReleaseRevisionStatus)
					logboek.LogLn()

					if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
						return fmt.Errorf("delete auto purge trigger file failed: %s", err)
					}
				}

				isReleaseExists = true
			}

			return nil
		}); err != nil {
			return fmt.Errorf("get release status failed: %s", err)
		}

		if releaseShouldBeDeleted {
			if err := logboek.LogProcess("Deleting release", logboek.LogProcessOptions{}, func() error {
				return releaseDelete(releaseName, releaseDeleteOptions{Purge: true})
			}); err != nil {
				return fmt.Errorf("release delete failed: %s", err)
			}

			if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
				return err
			}

			isReleaseExists = false
		} else if releaseShouldBeRolledBack {
			var isRollbackAttempt bool
			var latestSuccessfullyDeployedRevision int32

			logProcessMsg := "Trying rollback release to the latest successfully deployed revision"
			logProcessOptions := logboek.LevelLogProcessOptions{
				SuccessInfoSectionFunc: func() {
					if isRollbackAttempt {
						logboek.Default.LogFDetails("Release was rolled back to revision %d\n", latestSuccessfullyDeployedRevision)
					}
				},
			}
			if err := logboek.Default.LogProcess(logProcessMsg, logProcessOptions, func() error {
				var latestSuccessfullyDeployedReleaseRevisionErr error

				logProcessOptions := logboek.LevelLogProcessOptions{
					SuccessInfoSectionFunc: func() {
						if latestSuccessfullyDeployedReleaseRevisionErr == nil {
							logboek.Default.LogFDetails("latest-successfully-deployed-revision: %d\n", latestSuccessfullyDeployedRevision)
						} else {
							logboek.Default.LogLnDetails("Successfully deployed release revision was not found")
						}
					},
				}
				if err := logboek.Default.LogProcess(
					"Getting the latest successfully deployed release revision",
					logProcessOptions,
					func() error {
						latestSuccessfullyDeployedRevision, latestSuccessfullyDeployedReleaseRevisionErr = latestSuccessfullyDeployedReleaseRevision(releaseName)
						if latestSuccessfullyDeployedReleaseRevisionErr != nil && latestSuccessfullyDeployedReleaseRevisionErr != ErrNoSuccessfullyDeployedReleaseRevisionFound {
							return latestSuccessfullyDeployedReleaseRevisionErr
						}

						return nil
					},
				); err != nil {
					return fmt.Errorf("get latest successfully deployed release revision failed: %s", err)
				}

				if latestSuccessfullyDeployedReleaseRevisionErr == ErrNoSuccessfullyDeployedReleaseRevisionFound {
					return nil
				} else {
					isRollbackAttempt = true
				}

				var templatesFromRevision ChartTemplates
				logProcessMsg := fmt.Sprintf("Getting templates from release revision %d", latestSuccessfullyDeployedRevision)
				if err := logboek.LogProcessInline(logProcessMsg, logboek.LogProcessInlineOptions{}, func() error {
					templatesFromRevision, latestSuccessfullyDeployedReleaseRevisionErr = GetTemplatesFromReleaseRevision(releaseName, latestSuccessfullyDeployedRevision)
					return latestSuccessfullyDeployedReleaseRevisionErr
				}); err != nil {
					return fmt.Errorf("get templates from release revision failed: %s", err)
				}

				rollbackFunc := func() error {
					releaseRollbackOpts := ReleaseRollbackOptions{
						releaseRollbackOptions: releaseRollbackOptions{
							Timeout:       int64(opts.Timeout / time.Second),
							CleanupOnFail: true,
							DryRun:        opts.DryRun,
						},
					}

					var err error
					for i := 0; i < 5; i++ {
						logboek.LogF("Running helm rollback (%d try)...\n", i+1)

						err = ReleaseRollback(
							releaseName,
							latestSuccessfullyDeployedRevision,
							opts.ThreeWayMergeMode,
							releaseRollbackOpts,
						)

						if err == nil {
							return nil
						}
					}

					if err != nil {
						return fmt.Errorf("release rollback to revision %d failed: %s", latestSuccessfullyDeployedRevision, err)
					}

					panic("unexpected")
				}

				return runDeployProcess(releaseName, namespace, opts, templatesFromRevision, rollbackFunc)
			}); err != nil {
				return err
			}
		}

		return nil
	}

	if err := logboek.LogProcess("Running pre-deploy", logboek.LogProcessOptions{}, func() error {
		return preDeployFunc()
	}); err != nil {
		return err
	}

	var deployFunc func() error
	if isReleaseExists {
		deployFunc = func() error {
			logboek.LogF("Running helm upgrade...\n")
			logboek.LogOptionalLn()

			releaseUpdateOpts := ReleaseUpdateOptions{
				releaseUpdateOptions: releaseUpdateOptions{
					Timeout:       int64(opts.Timeout / time.Second),
					CleanupOnFail: true,
					Wait:          true,
					DryRun:        opts.DryRun,
				},
				Debug: opts.Debug,
			}

			if err := ReleaseUpdate(
				chartPath,
				releaseName,
				opts.Values,
				opts.SecretValues,
				opts.Set,
				opts.SetString,
				opts.ThreeWayMergeMode,
				releaseUpdateOpts,
			); err != nil {
				if strings.HasSuffix(err.Error(), "has no deployed releases") {
					logboek.LogWarnF("WARNING: Release is in improper state: %s\n", err.Error())

					if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
						return err
					}

					logboek.LogWarnLn("WARNING: Release will be removed with `helm delete --purge` on the next run of `werf deploy`")
				}

				return fmt.Errorf("release upgrade failed: %s", err)
			}

			if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
				return err
			}

			return nil
		}
	} else {
		deployFunc = func() error {
			logboek.LogF("Running helm install...\n")
			logboek.LogOptionalLn()

			releaseInstallOpts := ReleaseInstallOptions{
				releaseInstallOptions: releaseInstallOptions{
					Timeout: int64(opts.Timeout / time.Second),
					Wait:    true,
					DryRun:  opts.DryRun,
				},
				Debug: opts.Debug,
			}

			if err := ReleaseInstall(
				chartPath,
				releaseName,
				namespace,
				opts.Values,
				opts.SecretValues,
				opts.Set,
				opts.SetString,
				opts.ThreeWayMergeMode,
				releaseInstallOpts,
			); err != nil {
				if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
					return err
				}

				return fmt.Errorf("release install failed: %s", err)
			}

			return nil
		}
	}

	logProcessOptions := logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()}
	return logboek.Default.LogProcess("Running deploy", logProcessOptions, func() error {
		var templatesFromChart ChartTemplates

		if err := logboek.LogProcessInline("Getting chart templates", logboek.LogProcessInlineOptions{}, func() error {
			templatesFromChart, err = GetTemplatesFromChart(chartPath, releaseName, namespace, opts.Values, opts.SecretValues, opts.Set, opts.SetString)
			return err
		}); err != nil {
			return err
		}

		return runDeployProcess(releaseName, namespace, opts, templatesFromChart, deployFunc)
	})
}

func latestSuccessfullyDeployedReleaseRevision(releaseName string) (int32, error) {
	resp, err := releaseHistory(releaseName, releaseHistoryOptions{})
	if err != nil {
		return 0, fmt.Errorf("unable to get release history: %s", err)
	}

	for _, r := range resp.Releases {
		if r.Info.Status.Code.String() == "DEPLOYED" {
			return r.Version, nil
		}
	}

	return 0, ErrNoSuccessfullyDeployedReleaseRevisionFound
}

func runDeployProcess(releaseName, namespace string, _ ChartOptions, templates ChartTemplates, deployFunc func() error) error {
	oldLogsFromTime := resourcesWaiter.LogsFromTime
	resourcesWaiter.LogsFromTime = time.Now()
	defer func() {
		resourcesWaiter.LogsFromTime = oldLogsFromTime
	}()

	if err := deployFunc(); err != nil {
		return err
	}

	if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
		return err
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
		return fmt.Errorf("existing release is deployed in namespace %s (not in %s): check --namespace option value", resp.Release.Namespace, namespace)
	}

	return nil
}

func removeReleaseNamespacedResource(template Template, releaseNamespace string) error {
	resourceName := template.Metadata.Name
	resourceKing := template.Kind
	return removeResource(resourceName, resourceKing, template.Namespace(releaseNamespace))
}

// Not namespaced resource specifies without namespace
func removeResource(name, kind, namespace string) error {
	isNamespacedResource := namespace != ""

	groupVersionResource, err := kube.GroupVersionResourceByKind(kind)
	if err != nil {
		return err
	}

	var res dynamic.ResourceInterface
	if isNamespacedResource {
		res = kube.DynamicClient.Resource(groupVersionResource).Namespace(namespace)
	} else {
		res = kube.DynamicClient.Resource(groupVersionResource)
	}

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

	var logProcessMsg string
	if isNamespacedResource {
		logProcessMsg = fmt.Sprintf("Deleting %s/%s from namespace %s", groupVersionResource.Resource, name, namespace)
	} else {
		logProcessMsg = fmt.Sprintf("Deleting %s/%s", groupVersionResource.Resource, name)
	}

	return logboek.LogProcessInline(logProcessMsg, logboek.LogProcessInlineOptions{}, func() error {
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
	dirPath := filepath.Dir(filePath)

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

		logboek.Default.LogLnDetails("Auto purge trigger file was created")
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

		logboek.Default.LogLnDetails("Auto purge trigger file was deleted")
	}

	return nil
}

func autoPurgeTriggerFilePath(releaseName string) string {
	return filepath.Join(werf.GetServiceDir(), "helm", releaseName, "auto_purge_failed_release_on_next_deploy")
}
