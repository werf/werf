package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	helm_kube "k8s.io/helm/pkg/kube"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/kubeutils"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
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

func PurgeHelmRelease(ctx context.Context, releaseName, namespace string, withHooks bool) error {
	if err := logboek.Context(ctx).Info().LogProcess("Checking release existence").DoError(func() error {
		_, err := releaseStatus(ctx, releaseName, releaseStatusOptions{})
		if err != nil {
			if isReleaseNotFoundError(err) {
				return fmt.Errorf("release %s was not found", releaseName)
			}

			return fmt.Errorf("release status failed: %s", err)
		}

		return nil
	}); err != nil {
		if isReleaseNotFoundError(err) {
			logboek.Context(ctx).Default().LogLnDetails(err.Error())
			return nil
		}
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
		if err := logboek.Context(ctx).LogProcess(msg).DoError(func() error {
			for _, rev := range resp.Releases {
				revHooksToDelete := map[string]Template{}
				for _, h := range rev.Hooks {
					t, err := parseTemplate(h.Manifest)
					if err != nil {
						logboek.Context(ctx).Warn().LogF("WARNING: Parsing helm hook %s manifest failed: %s", h.Name, err)
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
					_ = logboek.Context(ctx).Info().LogProcess(msg).DoError(func() error {
						for hookId, hookTemplate := range revHooksToDelete {
							deletedHooks[hookId] = true

							if err := removeReleaseNamespacedResource(ctx, hookTemplate, rev.Namespace); err != nil {
								logboek.Context(ctx).Warn().LogF("WARNING: Failed to delete helm hook %s: %s", hookTemplate.Metadata.Name, err)
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

	if err := logboek.Context(ctx).LogProcess("Deleting release").DoError(func() error {
		return releaseDelete(ctx, releaseName, releaseDeleteOptions{Purge: true})
	}); err != nil {
		return fmt.Errorf("release delete failed: %s", err)
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

func DeployHelmChart(ctx context.Context, chartPath, releaseName, namespace string, opts ChartOptions) (err error) {
	logboek.Context(ctx).Debug().LogBlock("Helm values params").Do(func() {
		logboek.Context(ctx).Debug().LogF("values: %#v\n", opts.ChartValuesOptions.Values)
		logboek.Context(ctx).Debug().LogF("set: %#v\n", opts.ChartValuesOptions.Set)
		logboek.Context(ctx).Debug().LogF("set-string: %#v\n", opts.ChartValuesOptions.SetString)
		logboek.Context(ctx).Debug().LogF("secret-values: %#v\n", opts.ChartValuesOptions.SecretValues)
	})

	var isReleaseExists bool

	preDeployFunc := func() error {
		var latestReleaseRevision int32
		var latestReleaseRevisionStatus string
		var releaseShouldBeDeleted bool
		var releaseShouldBeRolledBack bool
		var latestReleaseThreeWayMergeEnabled bool

		if err := logboek.Context(ctx).Info().LogProcess("Checking release").
			Options(func(options types.LogProcessOptionsInterface) {
				options.SuccessInfoSectionFunc(func() {
					if isReleaseExists {
						logboek.Context(ctx).Info().LogFDetails("revision: %d\n", latestReleaseRevision)
						logboek.Context(ctx).Info().LogFDetails("revision-status: %s\n", latestReleaseRevisionStatus)

						if releaseShouldBeDeleted {
							logboek.Context(ctx).Info().LogLn()

							logboek.Context(ctx).Default().LogLnDetails(
								"Release will be deleted:\n",
								"* the latest release revision might be in an inconsistent state, and\n",
								"* auto purge trigger file is exists.",
							)
						} else if releaseShouldBeRolledBack {
							logboek.Context(ctx).LogLn()
							logboek.Context(ctx).Default().LogLnDetails(
								"Release should be rolled back to the latest successfully deployed revision:\n",
								"* the latest release revision might be in an inconsistent state.",
							)
						}
					} else {
						logboek.Context(ctx).Info().LogLnDetails("Release has not been deployed yet")
					}
				})
			}).
			DoError(func() error {
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
					if err := createAutoPurgeTriggerFilePath(ctx, releaseName); err != nil {
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

						if threeWayMergeMode == ThreeWayMergeDisabled {
							releaseShouldBeRolledBack = true
						} else if threeWayMergeMode == ThreeWayMergeOnlyNewReleases && !latestReleaseThreeWayMergeEnabled {
							releaseShouldBeRolledBack = true
						}
					}
				default:
					if exist, err := util.FileExists(autoPurgeTriggerFilePath(releaseName)); err != nil {
						return err
					} else if exist {
						logboek.Context(ctx).Warn().LogF("WARNING: Improper state:\n")
						logboek.Context(ctx).Warn().LogF("* auto purge trigger file is exists, and\n")
						logboek.Context(ctx).Warn().LogF("* the latest release revision (%s) should not be deleted.\n", latestReleaseRevisionStatus)
						logboek.Context(ctx).LogLn()

						if err := deleteAutoPurgeTriggerFilePath(ctx, releaseName); err != nil {
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
			if err := logboek.Context(ctx).LogProcess("Deleting release").DoError(func() error {
				return releaseDelete(ctx, releaseName, releaseDeleteOptions{Purge: true})
			}); err != nil {
				return fmt.Errorf("release delete failed: %s", err)
			}

			if err := deleteAutoPurgeTriggerFilePath(ctx, releaseName); err != nil {
				return err
			}

			isReleaseExists = false
		} else if releaseShouldBeRolledBack {
			var isRollbackAttempt bool
			var latestSuccessfullyDeployedRevision int32

			logProcessOptionsFunc := func(options types.LogProcessOptionsInterface) {
				options.SuccessInfoSectionFunc(func() {
					if isRollbackAttempt {
						logboek.Context(ctx).Default().LogFDetails("Release was rolled back to revision %d\n", latestSuccessfullyDeployedRevision)
					}
				})
			}
			if err := logboek.Context(ctx).Default().LogProcess("Trying rollback release to the latest successfully deployed revision").
				Options(logProcessOptionsFunc).
				DoError(func() error {
					var latestSuccessfullyDeployedReleaseRevisionErr error

					logProcessOptionsFunc := func(options types.LogProcessOptionsInterface) {
						options.SuccessInfoSectionFunc(func() {
							if latestSuccessfullyDeployedReleaseRevisionErr == nil {
								logboek.Context(ctx).Info().LogFDetails("latest-successfully-deployed-revision: %d\n", latestSuccessfullyDeployedRevision)
							} else {
								logboek.Context(ctx).Info().LogLnDetails("Successfully deployed release revision was not found")
							}
						})
					}
					if err := logboek.Context(ctx).Info().LogProcess("Getting the latest successfully deployed release revision").Options(logProcessOptionsFunc).DoError(func() error {
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
					if err := logboek.Context(ctx).Info().LogProcessInline("Getting templates from release revision %d", latestSuccessfullyDeployedRevision).DoError(func() error {
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
							logboek.Context(ctx).LogF("Running helm rollback (%d try)...\n", i+1)

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

					return runDeployProcess(ctx, releaseName, namespace, opts, templatesFromRevision, rollbackFunc)
				}); err != nil {
				return err
			}
		}

		return nil
	}

	if err := logboek.Context(ctx).Info().LogProcess("Running pre-deploy").DoError(preDeployFunc); err != nil {
		return err
	}

	var deployFunc func() error
	if isReleaseExists {
		deployFunc = func() error {
			logboek.Context(ctx).Info().LogF("Running helm upgrade...\n")

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
				ctx,
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
					logboek.Context(ctx).Warn().LogF("WARNING: Release is in improper state: %s\n", err.Error())

					if err := createAutoPurgeTriggerFilePath(ctx, releaseName); err != nil {
						return err
					}

					logboek.Context(ctx).Warn().LogLn("WARNING: Release will be removed with `helm delete --purge` on the next run of `werf deploy`")
				}

				return fmt.Errorf("release upgrade failed: %s", err)
			}

			if err := deleteAutoPurgeTriggerFilePath(ctx, releaseName); err != nil {
				return err
			}

			return nil
		}
	} else {
		deployFunc = func() error {
			logboek.Context(ctx).Info().LogF("Running helm install...\n")

			releaseInstallOpts := ReleaseInstallOptions{
				releaseInstallOptions: releaseInstallOptions{
					Timeout: int64(opts.Timeout / time.Second),
					Wait:    true,
					DryRun:  opts.DryRun,
				},
				Debug: opts.Debug,
			}

			if err := ReleaseInstall(
				ctx,
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
				if err := createAutoPurgeTriggerFilePath(ctx, releaseName); err != nil {
					return err
				}

				return fmt.Errorf("release install failed: %s", err)
			}

			return nil
		}
	}

	var templatesFromChart ChartTemplates

	if err := logboek.Context(ctx).Info().LogProcessInline("Getting chart templates").DoError(func() error {
		templatesFromChart, err = GetTemplatesFromChart(ctx, chartPath, releaseName, namespace, opts.Values, opts.SecretValues, opts.Set, opts.SetString)
		return err
	}); err != nil {
		return err
	}

	return runDeployProcess(ctx, releaseName, namespace, opts, templatesFromChart, deployFunc)
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

func runDeployProcess(ctx context.Context, releaseName, namespace string, _ ChartOptions, templates ChartTemplates, deployFunc func() error) error {
	oldLogsFromTime := resourcesWaiter.LogsFromTime
	resourcesWaiter.LogsFromTime = time.Now()
	defer func() {
		resourcesWaiter.LogsFromTime = oldLogsFromTime
	}()

	if err := deployFunc(); err != nil {
		return err
	}

	if err := deleteAutoPurgeTriggerFilePath(ctx, releaseName); err != nil {
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

func removeReleaseNamespacedResource(ctx context.Context, template Template, releaseNamespace string) error {
	resourceName := template.Metadata.Name
	resourceKing := template.Kind
	return kubeutils.RemoveResourceAndWaitUntilRemoved(ctx, resourceName, resourceKing, template.Namespace(releaseNamespace))
}

func createAutoPurgeTriggerFilePath(ctx context.Context, releaseName string) error {
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

		logboek.Context(ctx).Info().LogLnDetails("Auto purge trigger file was created")
	}

	return nil
}

func deleteAutoPurgeTriggerFilePath(ctx context.Context, releaseName string) error {
	filePath := autoPurgeTriggerFilePath(releaseName)
	if fileExist, err := util.FileExists(filePath); err != nil {
		return err
	} else if fileExist {
		if err := os.Remove(filePath); err != nil {
			return err
		}

		logboek.Context(ctx).Info().LogLnDetails("Auto purge trigger file was deleted")
	}

	return nil
}

func autoPurgeTriggerFilePath(releaseName string) string {
	return filepath.Join(werf.GetServiceDir(), "helm", releaseName, "auto_purge_failed_release_on_next_deploy")
}
