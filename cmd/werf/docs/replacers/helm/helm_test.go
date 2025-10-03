package helm

import (
	"os"
	"testing"

	helm_v3 "github.com/werf/3p-helm-for-werf-helm/cmd/helm"
	"github.com/werf/3p-helm-for-werf-helm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
	helm "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm"
)

func TestReplaceHelmCreateDocs(t *testing.T) {
	cmd := ReplaceHelmCreateDocs(helm_v3.NewCreateCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmCreateDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmEnvDocs(t *testing.T) {
	cmd := ReplaceHelmEnvDocs(helm_v3.NewEnvCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmEnvDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmHistoryDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmHistoryDocs(helm_v3.NewHistoryCmd(actionConfig, os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmHistoryDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmLintDocs(t *testing.T) {
	cmd := ReplaceHelmLintDocs(helm_v3.NewLintCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmLintDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmListDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmListDocs(helm_v3.NewListCmd(actionConfig, os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmListDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmPullDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmPullDocs(helm_v3.NewPullCmd(actionConfig, os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmPullDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmRollbackDocs(t *testing.T) {
	var namespace string
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmRollbackDocs(helm_v3.NewRollbackCmd(actionConfig, os.Stdout, helm_v3.RollbackCmdOptions{
		StagesSplitter:              helm.NewStagesSplitter(),
		StagesExternalDepsGenerator: helm.NewStagesExternalDepsGenerator(&actionConfig.RESTClientGetter, &namespace),
	}))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmRollbackDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmStatusDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmStatusDocs(helm_v3.NewStatusCmd(actionConfig, os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmStatusDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmUninstallDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmUninstallDocs(helm_v3.NewUninstallCmd(actionConfig, os.Stdout, helm_v3.UninstallCmdOptions{
		StagesSplitter: helm.NewStagesSplitter(),
	}))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmUninstallDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmVerifyDocs(t *testing.T) {
	cmd := ReplaceHelmVerifyDocs(helm_v3.NewVerifyCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmVerifyDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}

func TestReplaceHelmVersionDocs(t *testing.T) {
	cmd := ReplaceHelmVersionDocs(helm_v3.NewVersionCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else if ann != GetHelmVersionDocs().LongMD {
		t.Error("The annotation does not match!")
	}
}
