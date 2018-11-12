package main

import (
	"fmt"
	"time"

	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/kubedog/pkg/tracker"
	"github.com/flant/kubedog/pkg/trackers/rollout"
)

func main() {
	var err error

	err = lock.Init()
	if err != nil {
		panic(err)
	}

	err = kube.Init()
	if err != nil {
		panic(err)
	}

	ruby2go.RunCli("deploy-watcher", func(args map[string]interface{}) (interface{}, error) {
		namespace := args["namespace"].(string)
		if namespace == "" {
			return nil, fmt.Errorf("namespace argument required!")
		}

		resourceName := args["resourceName"].(string)
		if resourceName == "" {
			return nil, fmt.Errorf("resourceName argument required!")
		}

		timeout := args["timeout"].(float64)

		var logsFromTime time.Time
		if logsFromTimeOption, ok := args["logsFromTime"]; ok {
			logsFromTime, err = time.Parse(time.RFC3339, logsFromTimeOption.(string))
			if err != nil {
				return nil, err
			}
		} else {
			logsFromTime = time.Time{}
		}

		switch action := args["action"]; action {
		case "watch job":
			err := rollout.TrackJobTillDone(resourceName, namespace, kube.Kubernetes, tracker.Options{Timeout: time.Second * time.Duration(timeout), LogsFromTime: logsFromTime})
			if err != nil {
				return nil, fmt.Errorf("error tracking job `%s` in namespace `%s`: %s", resourceName, namespace, err)
			}
		case "watch deployment":
			err := rollout.TrackDeploymentTillReady(resourceName, namespace, kube.Kubernetes, tracker.Options{Timeout: time.Second * time.Duration(timeout), LogsFromTime: logsFromTime})
			if err != nil {
				return nil, fmt.Errorf("error tracking deployment `%s` in namespace `%s`: %s", resourceName, namespace, err)
			}
		default:
			return nil, fmt.Errorf("unknown action \"%s\"", action)
		}

		return nil, nil
	})
}
