package deploy

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/kubedog/pkg/tracker"
	"github.com/flant/kubedog/pkg/trackers/rollout"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/releaseutil"
	"k8s.io/kubernetes/pkg/util/file"
)

type CommonHelmOptions struct {
	KubeContext string
}

func PurgeHelmRelease(releaseName string, opts CommonHelmOptions) error {
	return withLockedHelmRelease(releaseName, func() error {
		return doPurgeHelmRelease(releaseName, opts)
	})
}

func doPurgeHelmRelease(releaseName string, opts CommonHelmOptions) error {
	args := []string{}
	if opts.KubeContext != "" {
		args = append(args, "--kube-context")
		args = append(args, opts.KubeContext)
	}

	helmStatusStdout, helmStatusStderr, helmStatusErr := HelmCmd(append([]string{"status", releaseName}, args...)...)
	if helmStatusErr != nil {
		if strings.HasSuffix(helmStatusStderr, "not found") {
			return fmt.Errorf("Helm release '%s' doesn't exist", releaseName)
		}
		return fmt.Errorf("failed to check release status: %s\n%s\n%s", helmStatusErr, helmStatusStdout, helmStatusStderr)
	}

	fmt.Printf("# Purging helm release '%s'...\n", releaseName)
	helmPurgeStdout, helmPurgeStderr, helmPurgeErr := HelmCmd(append([]string{"delete", "--purge", releaseName}, args...)...)
	if helmPurgeErr != nil {
		return fmt.Errorf("failed to purge release: %s\n%s\n%s", helmPurgeErr, helmPurgeStdout, helmPurgeStderr)
	}

	return nil
}

type HelmChartOptions struct {
	Set     []string
	Values  []string
	DryRun  bool
	Debug   bool
	Timeout time.Duration
	CommonHelmOptions
}

func withLockedHelmRelease(releaseName string, f func() error) error {
	lockName := fmt.Sprintf("helm_release.%s", releaseName)
	return lock.WithLock(lockName, lock.LockOptions{}, f)
}

func DeployHelmChart(chartPath string, releaseName string, namespace string, opts HelmChartOptions) error {
	return withLockedHelmRelease(releaseName, func() error {
		return doDeployHelmChart(chartPath, releaseName, namespace, opts)
	})
}

func doDeployHelmChart(chartPath string, releaseName string, namespace string, opts HelmChartOptions) error {
	releaseExist, err := isReleaseExist(releaseName)
	if err != nil {
		return fmt.Errorf("checking release failed: %s", err)
	}

	templates, err := parseTemplates(chartPath, releaseName)
	if err != nil {
		return fmt.Errorf("parsing templates failed: %s", err)
	}

	if err := removeOldJobs(templates, namespace); err != nil {
		return fmt.Errorf("removing old jobs failed: %s", err)
	}

	deployStartTime := time.Now()

	jobHooksWatcherDone, err := watchJobHooks(templates, releaseExist, deployStartTime, namespace, opts)
	if err != nil {
		return fmt.Errorf("watching job hooks failed: %s", err)
	}

	args := commonDeployCommandArgs(namespace, opts)
	if releaseExist {
		args = append([]string{"upgrade", releaseName, chartPath}, args...)
		fmt.Printf("# Upgrading helm release '%s'...\n", releaseName)
	} else {
		args = append([]string{"install", chartPath, "--name", releaseName}, args...)
		fmt.Printf("# Installing helm release '%s'...\n", releaseName)
	}

	stdout, stderr, err := HelmCmd(args...)
	if err != nil {
		if strings.HasSuffix(stderr, "has no deployed releases\n") {
			fmt.Printf("WARN: Helm release '%s' is in improper state: %s", releaseName, stderr)
			fmt.Printf("WARN: Helm release %s will be removed with `helm delete --purge` on the next run of `dapp kube deploy`", releaseName)
		}

		if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
			return err
		}

		return fmt.Errorf("%s\n%s", stdout, stderr)
	}

	if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
		return err
	}

	<-jobHooksWatcherDone

	fmt.Printf("%s\n%s\n", stdout, stderr)

	if err := watchDeployments(templates, deployStartTime, namespace, opts); err != nil {
		return fmt.Errorf("watching deployments failed: %s", err)
	}

	return nil
}

func watchDeployments(templates *ChartTemplates, deployStartTime time.Time, namespace string, opts HelmChartOptions) error {
	for _, template := range templates.Deployments() {
		fmt.Printf("# Run watch for deployment '%s'\n", template.Metadata.Name)

		err := rollout.TrackDeploymentTillReady(template.Metadata.Name, template.Namespace(namespace), kube.Kubernetes, tracker.Options{Timeout: time.Second * time.Duration(opts.Timeout), LogsFromTime: deployStartTime})
		if err != nil {
			return err
		}
	}

	return nil
}

func watchJobHooks(templates *ChartTemplates, releaseExist bool, deployStartTime time.Time, namespace string, opts HelmChartOptions) (chan bool, error) {
	jobHooksWatcherDone := make(chan bool)

	jobHooksToWatch, err := jobHooksToWatch(templates, releaseExist)
	if err != nil {
		return jobHooksWatcherDone, err
	}

	go func() {
		for _, template := range jobHooksToWatch {
			fmt.Printf("# Run watch for job '%s'\n", template.Metadata.Name)

			var jobNamespace string
			if template.Metadata.Namespace != "" {
				jobNamespace = template.Metadata.Namespace
			} else {
				jobNamespace = namespace
			}

			err := rollout.TrackJobTillDone(template.Metadata.Name, jobNamespace, kube.Kubernetes, tracker.Options{Timeout: time.Second * time.Duration(opts.Timeout), LogsFromTime: deployStartTime})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR %s\n", err)
				break
			}
		}

		jobHooksWatcherDone <- true
	}()

	return jobHooksWatcherDone, nil
}

func commonDeployCommandArgs(namespace string, opts HelmChartOptions) []string {
	var args []string

	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	for _, set := range opts.Set {
		args = append(args, "--set", set)
	}

	for _, values := range opts.Values {
		args = append(args, "--values", values)
	}

	if opts.KubeContext != "" {
		args = append(args, "--kube-context", opts.KubeContext)
	}

	if opts.DryRun {
		args = append(args, "--dry-run")
		args = append(args, "--debug")
	}

	if opts.Timeout != 0 {
		args = append(args, "--timeout", fmt.Sprintf("%v", opts.Timeout.Seconds()))
	}

	return args
}

func isReleaseExist(releaseName string) (bool, error) {
	var releaseStatus string
	var releaseExist bool

	helmStatusStdout, _, helmStatusErr := HelmCmd("status", releaseName)
	if helmStatusErr == nil {
		statusLinePrefix := "STATUS: "
		scanner := bufio.NewScanner(strings.NewReader(helmStatusStdout))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, statusLinePrefix) {
				releaseStatus = line[len(statusLinePrefix):]
				break
			}
		}
	}

	if helmStatusErr != nil {
		releaseExist = false
		if err := createAutoPurgeTriggerFilePath(releaseName); err != nil {
			return false, err
		}
	} else if releaseStatus != "" && (releaseStatus == "FAILED" || releaseStatus == "PENDING_INSTALL") {
		releaseExist = true

		if exist, err := file.FileExists(autoPurgeTriggerFilePath(releaseName)); err != nil {
			return false, err
		} else if exist {
			fmt.Printf("# Delete release '%s'\n", releaseName)
			if _, _, err := HelmCmd("delete", "--purge", releaseName); err != nil {
				return false, err
			}

			releaseExist = false
		}
	} else {
		if exist, err := file.FileExists(autoPurgeTriggerFilePath(releaseName)); err != nil {
			return false, err
		} else if exist {
			fmt.Printf("WARN: Will not purge helm release '%s': expected FAILED or PENDING_INSTALL release status, got %s\n", releaseName, releaseStatus)
		}

		releaseExist = true

		if err := deleteAutoPurgeTriggerFilePath(releaseName); err != nil {
			return false, err
		}
	}

	return releaseExist, nil
}

type ChartTemplates []*Template

func (templates *ChartTemplates) Jobs() []*Template {
	return templates.ByKind("Job")
}

func (templates *ChartTemplates) Deployments() []*Template {
	return templates.ByKind("Deployment")
}

func (templates *ChartTemplates) ByKind(kind string) []*Template {
	var resultTemplates []*Template

	for _, template := range []*Template(*templates) {
		if template.Kind == kind {
			resultTemplates = append(resultTemplates, template)
		}
	}

	return resultTemplates
}

type Template struct {
	Version  string `yaml:"apiVersion"`
	Kind     string `yaml:"kind,omitempty"`
	Metadata *struct {
		Name        string            `yaml:"name"`
		Namespace   string            `yaml:"namespace"`
		Annotations map[string]string `yaml:"annotations"`
		Uid         string            `yaml:"uid"`
	} `yaml:"metadata,omitempty"`
	Status string `yaml:"status,omitempty"`
}

func (t *Template) Namespace(namespace string) string {
	if t.Metadata.Namespace != "" {
		return t.Metadata.Namespace
	}

	return namespace
}

func parseTemplates(chartPath, releaseName string) (*ChartTemplates, error) {
	var templates []*Template

	stdout, stderr, err := HelmCmd("template", chartPath, "--name", releaseName)
	if err != nil {
		return nil, fmt.Errorf(stderr)
	}

	for _, doc := range releaseutil.SplitManifests(stdout) {
		var t Template
		err := yaml.Unmarshal([]byte(doc), &t)
		if err != nil {
			return nil, err
		}

		if t.Metadata != nil && t.Metadata.Name != "" {
			templates = append(templates, &t)
		}
	}

	chartTemplates := ChartTemplates(templates)
	return &chartTemplates, nil
}

func removeOldJobs(templates *ChartTemplates, namespace string) error {
	isJobExist := func(name string, namespace string) (bool, error) {
		options := v1.GetOptions{IncludeUninitialized: true}
		_, err := kube.Kubernetes.BatchV1().Jobs(namespace).Get(name, options)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}

			return true, err
		}

		return true, nil
	}

	for _, template := range templates.Jobs() {
		value, ok := template.Metadata.Annotations["dapp/recreate"]
		if ok && (value == "0" || value == "false") {
			continue
		}

		exist, err := isJobExist(template.Metadata.Name, template.Namespace(namespace))
		if err != nil {
			return err
		} else if !exist {
			continue
		}

		fmt.Printf("# Deleting hook job '%s' (dapp/recreate)...\n", template.Metadata.Name)

		deletePropagation := v1.DeletePropagationForeground
		deleteOptions := &v1.DeleteOptions{
			PropagationPolicy: &deletePropagation,
		}
		err = kube.Kubernetes.BatchV1().Jobs(template.Namespace(namespace)).Delete(template.Metadata.Name, deleteOptions)
		if err != nil {
			return err
		}

		for {
			exist, err := isJobExist(template.Metadata.Name, template.Namespace(namespace))
			if err != nil {
				return err
			} else if !exist {
				break
			}

			time.Sleep(500 * time.Millisecond)
		}
	}

	return nil
}

func jobHooksToWatch(templates *ChartTemplates, releaseExist bool) ([]*Template, error) {
	var jobHooksToWatch []*Template
	jobHooksByType := make(map[string][]*Template)

	for _, template := range templates.Jobs() {
		if anno, ok := template.Metadata.Annotations["helm.sh/hook"]; ok {
			for _, hookType := range strings.Split(anno, ",") {
				if _, ok := jobHooksByType[hookType]; !ok {
					jobHooksByType[hookType] = []*Template{}
				}
				jobHooksByType[hookType] = append(jobHooksByType[hookType], template)
			}
		}
	}

	for _, templates := range jobHooksByType {
		sort.Slice(templates, func(i, j int) bool {
			toWeight := func(t *Template) int {
				val, ok := t.Metadata.Annotations["helm.sh/hook-weight"]
				if !ok {
					return 0
				}

				i, err := strconv.Atoi(val)
				if err != nil {
					fmt.Printf("WARN: Incorrect hook-weight anno value '%v'\n", val)
					return 0
				}

				return i
			}

			return toWeight(templates[i]) < toWeight(templates[j])
		})
	}

	if releaseExist {
		for _, hookType := range []string{"pre-upgrade", "post-upgrade"} {
			if templates, ok := jobHooksByType[hookType]; ok {
				jobHooksToWatch = append(jobHooksToWatch, templates...)
			}
		}
	} else {
		for _, hookType := range []string{"pre-install", "post-install"} {
			if templates, ok := jobHooksByType[hookType]; ok {
				jobHooksToWatch = append(jobHooksToWatch, templates...)
			}
		}
	}

	return jobHooksToWatch, nil
}

func createAutoPurgeTriggerFilePath(releaseName string) error {
	filePath := autoPurgeTriggerFilePath(releaseName)
	dirPath := path.Dir(filePath)

	if fileExist, err := file.FileExists(filePath); err != nil {
		return err
	} else if !fileExist {
		if dirExist, err := file.FileExists(dirPath); err != nil {
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
	if fileExist, err := file.FileExists(filePath); err != nil {
		return err
	} else if fileExist {
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}

	return nil
}

func autoPurgeTriggerFilePath(releaseName string) string {
	return filepath.Join(dapp.GetHomeDir(), "helm", releaseName, "auto_purge_failed_release_on_next_deploy")
}
