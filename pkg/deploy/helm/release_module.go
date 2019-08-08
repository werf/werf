package helm

import (
	"fmt"

	"k8s.io/helm/pkg/hooks"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/tiller"
	"k8s.io/helm/pkg/tiller/environment"
)

type ReleaseModule struct {
	tiller.ReleaseModule
}

func (m *ReleaseModule) Create(target *release.Release, req *services.InstallReleaseRequest, env *environment.Environment) error {
	return WithKubeClientCreatePerformHook(target, req.Timeout, func() error {
		return m.ReleaseModule.Create(target, req, env)
	})
}

func (m *ReleaseModule) Update(current, target *release.Release, req *services.UpdateReleaseRequest, env *environment.Environment) error {
	return withKubeClientUpdatePerformHook(target, req.Timeout, func() error {
		return m.ReleaseModule.Update(current, target, req, env)
	})
}

func WithKubeClientCreatePerformHook(newRelease *release.Release, timeout int64, f func() error) error {
	oldCreatePerformHooks := kubeClient.createPerformHookFunc

	kubeClient.createPerformHookFunc = func() error {
		return withoutKubeClientPerformHooks(func() error {
			return execReleaseHook(newRelease, hooks.PostApplyOnInstall, timeout)
		})
	}

	err := f()

	kubeClient.createPerformHookFunc = oldCreatePerformHooks

	return err
}

func withKubeClientUpdatePerformHook(newRelease *release.Release, timeout int64, f func() error) error {
	oldUpdatePerformHooks := kubeClient.updatePerformHookFunc

	kubeClient.updatePerformHookFunc = func() error {
		return withoutKubeClientPerformHooks(func() error {
			return execReleaseHook(newRelease, hooks.PostApplyOnUpgrade, timeout)
		})
	}

	err := f()

	kubeClient.updatePerformHookFunc = oldUpdatePerformHooks

	return err
}

func withoutKubeClientPerformHooks(f func() error) error {
	oldCreatePerformHooks := kubeClient.createPerformHookFunc
	oldUpdatePerformHooks := kubeClient.updatePerformHookFunc

	kubeClient.createPerformHookFunc = nil
	kubeClient.updatePerformHookFunc = nil

	err := f()

	kubeClient.createPerformHookFunc = oldCreatePerformHooks
	kubeClient.updatePerformHookFunc = oldUpdatePerformHooks

	return err
}

func execReleaseHook(newRelease *release.Release, hook string, timeout int64) error {
	if err := tillerReleaseServer.ExecHook(newRelease.Hooks, newRelease.Name, newRelease.Namespace, hook, timeout); err != nil {
		msg := fmt.Sprintf("Release %q failed %s: %s", newRelease.Name, hook, err)
		tillerReleaseServer.Log("warning: %s", msg)
		newRelease.Info.Status.Code = release.Status_FAILED
		newRelease.Info.Description = msg
		tillerReleaseServer.RecordRelease(newRelease, true)
		return err
	}

	return nil
}
