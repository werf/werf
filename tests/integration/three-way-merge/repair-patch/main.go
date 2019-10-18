package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flant/kubedog/pkg/kube"
)

var (
	appState string
)

func setAppState(directoryName string) {
	appState = directoryName
}

func werfCommandSimpleOutput(arg ...string) error {
	cmd := exec.Command("werf", arg...)

	if appState == "" {
		panic("app state should be set by setAppState")
	}

	cmd.Dir = appState
	cmd.Env = os.Environ()

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error getting stdout pipe: %s", err)
	}
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error getting stderr pipe: %s", err)
	}
	outReader := io.MultiReader(stdoutReader, stderrReader)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %s", err)
	}

	buf := make([]byte, 4096)
	for {
		n, err := outReader.Read(buf)
		if n > 0 {
			fmt.Printf("%s", buf[:n])
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("error reading command output: %s", err)
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting command: %s", err)
	}

	return nil
}

func werfDeploy() error {
	return werfCommandSimpleOutput("deploy", "--env", "dev")

}

func werfDismiss() error {
	return werfCommandSimpleOutput("dismiss", "--env", "dev", "--with-namespace")
}

func test() error {
	defer func() {
		werfDismiss()
	}()

	if err := kube.Init(kube.InitOptions{}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	setAppState("app1")

	if err := werfDeploy(); err != nil {
		return fmt.Errorf("werf deploy 1 failed: %s", err)
	}

	namespace := "three-way-merge-repair-patch-dev"

	if mycm1, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get("mycm1", metav1.GetOptions{}); err == nil {
		mycm1.Data = make(map[string]string)
		mycm1.Data["newKey"] = "newValue"
		if _, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Update(mycm1); err != nil {
			return fmt.Errorf("update mycm1 error: %s", err)
		}
	} else {
		return fmt.Errorf("get mycm1 error: %s", err)
	}

	if mydeploy1, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{}); err == nil {
		var replicas int32 = 2
		mydeploy1.Spec.Replicas = &replicas
		if _, err := kube.Kubernetes.AppsV1().Deployments(namespace).Update(mydeploy1); err != nil {
			return fmt.Errorf("update mydeploy1 error: %s", err)
		}
	} else {
		return fmt.Errorf("get mydeploy1 error: %s", err)
	}

	if err := werfDeploy(); err != nil {
		return fmt.Errorf("werf deploy 2 failed: %s", err)
	}

	if mycm1, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get("mycm1", metav1.GetOptions{}); err == nil {
		d, err := json.Marshal(mycm1.Data)
		if err != nil {
			return err
		}
		if string(d) != `{"newKey":"newValue"}` {
			return fmt.Errorf("unexpected configmap/mycm1 data after werf deploy 2: %s", d)
		}
		if mycm1.Annotations["debug.werf.io/repair-patch"] != `{"data":{"aloe":"aloha","moloko":"omlet"}}` {
			return fmt.Errorf("unexpected configmap/mycm1 debug.werf.io/repair-patch annotation after werf deploy 2: %s", mycm1.Annotations["debug.werf.io/repair-patch"])
		}

		if _, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Patch("mycm1", types.StrategicMergePatchType, []byte(mycm1.Annotations["debug.werf.io/repair-patch"])); err != nil {
			return fmt.Errorf("patch mycm1 error: %s", err)
		}
	} else {
		return fmt.Errorf("get mycm1 error: %s", err)
	}

	if mydeploy1, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{}); err == nil {
		if *mydeploy1.Spec.Replicas != 2 {
			return fmt.Errorf("unexpected deployment/mydeploy1 replicas after werf deploy 2: %d", *mydeploy1.Spec.Replicas)
		}
		if mydeploy1.Annotations["debug.werf.io/repair-patch"] != `{"spec":{"replicas":1}}` {
			return fmt.Errorf("unexpected deployment/mydeploy1 debug.werf.io/repair-patch annotation after werf deploy 2: %s", mydeploy1.Annotations["debug.werf.io/repair-patch"])
		}
	} else {
		return fmt.Errorf("get mydeploy1 error: %s", err)
	}

	setAppState("app2")

	if err := werfDeploy(); err != nil {
		return fmt.Errorf("werf deploy 3 failed: %s", err)
	}

	if mycm1, err := kube.Kubernetes.CoreV1().ConfigMaps(namespace).Get("mycm1", metav1.GetOptions{}); err == nil {
		d, err := json.Marshal(mycm1.Data)
		if err != nil {
			return err
		}
		if string(d) != `{"aloe":"aloha","moloko":"omlet","newKey":"newValue"}` {
			return fmt.Errorf("unexpected configmap/mycm1 data after werf deploy 3: %s", d)
		}
		if mycm1.Annotations["debug.werf.io/repair-patch"] != `{}` {
			return fmt.Errorf("unexpected configmap/mycm1 debug.werf.io/repair-patch annotation after werf deploy 3: %s", mycm1.Annotations["debug.werf.io/repair-patch"])
		}
	} else {
		return fmt.Errorf("get mycm1 error: %s", err)
	}

	if mydeploy1, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get("mydeploy1", metav1.GetOptions{}); err == nil {
		if *mydeploy1.Spec.Replicas != 2 {
			return fmt.Errorf("unexpected deployment/mydeploy1 replicas after werf deploy 3: %d", *mydeploy1.Spec.Replicas)
		}
		if mydeploy1.Annotations["debug.werf.io/repair-patch"] != `{}` {
			return fmt.Errorf("unexpected deployment/mydeploy1 debug.werf.io/repair-patch annotation after werf deploy 3: %s", mydeploy1.Annotations["debug.werf.io/repair-patch"])
		}
	} else {
		return fmt.Errorf("get mydeploy1 error: %s", err)
	}

	return nil
}

func main() {
	if err := test(); err != nil {
		log.Fatal(err)
	}
}
