package lock_manager

import (
	"testing"

	"github.com/werf/werf/v2/pkg/storage"
)

func TestParseKubernetesParams(t *testing.T) {
	// Check bad scheme.
	{
		params, err := ParseKubernetesParams("kubertenes://allo")
		if err == nil {
			t.Errorf("expected error, got %#v", params)
		} else if err.Error() != `bad address "kubertenes://allo": expected kubernetes:// scheme` {
			t.Errorf("unexpected error: %v", err)
		}
	}

	checkParseKubernetesParams(t, storage.DefaultKubernetesStorageAddress, &KubernetesParams{
		ConfigContext: "",
		ConfigPath:    "",
		Namespace:     "werf-synchronization",
	})

	checkParseKubernetesParams(t, "kubernetes://mynamespace:mycontext@/tmp/kubeconfig", &KubernetesParams{
		Namespace:     "mynamespace",
		ConfigContext: "mycontext",
		ConfigPath:    "/tmp/kubeconfig",
	})

	checkParseKubernetesParams(t, "kubernetes://werf-synchronization-2@base64:YXBpVmVyc2lvbjogdjEKY2x1c3RlcnM6Ci0gY2x1c3RlcjoKICAgIGNlcnRpZmljYXRlLWF1dGhvcml0eTogL2hvbWUvbXlob21lLy5taW5pa3ViZS9jYS5jcnQKICAgIHNlcnZlcjogaHR0cHM6Ly8xNzIuMTcuMC40Ojg0NDMKICBuYW1lOiBtaW5pa3ViZQpjb250ZXh0czoKLSBjb250ZXh0OgogICAgY2x1c3RlcjogbWluaWt1YmUKICAgIHVzZXI6IG1pbmlrdWJlCiAgbmFtZTogbWluaWt1YmUKY3VycmVudC1jb250ZXh0OiAiIgpraW5kOiBDb25maWcKcHJlZmVyZW5jZXM6IHt9CnVzZXJzOgotIG5hbWU6IG1pbmlrdWJlCiAgdXNlcjoKICAgIGNsaWVudC1jZXJ0aWZpY2F0ZTogL2hvbWUvbXlob21lLy5taW5pa3ViZS9wcm9maWxlcy9taW5pa3ViZS9jbGllbnQuY3J0CiAgICBjbGllbnQta2V5OiAvaG9tZS9teWhvbWUvLm1pbmlrdWJlL3Byb2ZpbGVzL21pbmlrdWJlL2NsaWVudC5rZXkK", &KubernetesParams{
		Namespace:        "werf-synchronization-2",
		ConfigDataBase64: "YXBpVmVyc2lvbjogdjEKY2x1c3RlcnM6Ci0gY2x1c3RlcjoKICAgIGNlcnRpZmljYXRlLWF1dGhvcml0eTogL2hvbWUvbXlob21lLy5taW5pa3ViZS9jYS5jcnQKICAgIHNlcnZlcjogaHR0cHM6Ly8xNzIuMTcuMC40Ojg0NDMKICBuYW1lOiBtaW5pa3ViZQpjb250ZXh0czoKLSBjb250ZXh0OgogICAgY2x1c3RlcjogbWluaWt1YmUKICAgIHVzZXI6IG1pbmlrdWJlCiAgbmFtZTogbWluaWt1YmUKY3VycmVudC1jb250ZXh0OiAiIgpraW5kOiBDb25maWcKcHJlZmVyZW5jZXM6IHt9CnVzZXJzOgotIG5hbWU6IG1pbmlrdWJlCiAgdXNlcjoKICAgIGNsaWVudC1jZXJ0aWZpY2F0ZTogL2hvbWUvbXlob21lLy5taW5pa3ViZS9wcm9maWxlcy9taW5pa3ViZS9jbGllbnQuY3J0CiAgICBjbGllbnQta2V5OiAvaG9tZS9teWhvbWUvLm1pbmlrdWJlL3Byb2ZpbGVzL21pbmlrdWJlL2NsaWVudC5rZXkK",
	})
}

func checkParseKubernetesParams(t *testing.T, address string, expected *KubernetesParams) {
	params, err := ParseKubernetesParams(address)
	if err != nil {
		t.Error(err)
	}

	if expected == nil && params != nil {
		t.Errorf("expected nil kubernetes params, got %#v", params)
	}

	if params.ConfigContext != expected.ConfigContext {
		t.Errorf("expected kube context %#v, got %#v", expected.ConfigContext, params.ConfigContext)
	}
	if params.ConfigDataBase64 != expected.ConfigDataBase64 {
		t.Errorf("expected config data base64:\n%s\n\ngot:\n%s", expected.ConfigDataBase64, params.ConfigDataBase64)
	}
	if params.ConfigPath != expected.ConfigPath {
		t.Errorf("expected config path %#v, got %#v", expected.ConfigPath, params.ConfigPath)
	}
	if params.Namespace != expected.Namespace {
		t.Errorf("expected namespace %#v, got %#v", expected.Namespace, params.Namespace)
	}
}
