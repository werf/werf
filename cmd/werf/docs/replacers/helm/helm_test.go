package helm

import (
	"os"
	"testing"

	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/helm"
)

func TestReplaceHelmCreateDocs(t *testing.T) {
	cmd := ReplaceHelmCreateDocs(helm_v3.NewCreateCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmCreateDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmEnvDocs(t *testing.T) {
	cmd := ReplaceHelmEnvDocs(helm_v3.NewEnvCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmEnvDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmHistoryDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmHistoryDocs(helm_v3.NewHistoryCmd(actionConfig, os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmHistoryDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmLintDocs(t *testing.T) {
	cmd := ReplaceHelmLintDocs(helm_v3.NewLintCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmLintDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmListDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmListDocs(helm_v3.NewListCmd(actionConfig, os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmListDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmPullDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmPullDocs(helm_v3.NewPullCmd(actionConfig, os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmPullDocs().LongMD {
			t.Error("The annotation does not match!")
		}
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
	} else {
		if ann != GetHelmRollbackDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmStatusDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmStatusDocs(helm_v3.NewStatusCmd(actionConfig, os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmStatusDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmUninstallDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmUninstallDocs(helm_v3.NewUninstallCmd(actionConfig, os.Stdout, helm_v3.UninstallCmdOptions{
		StagesSplitter: helm.NewStagesSplitter(),
	}))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmUninstallDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmVerifyDocs(t *testing.T) {
	cmd := ReplaceHelmVerifyDocs(helm_v3.NewVerifyCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmVerifyDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmVersionDocs(t *testing.T) {
	cmd := ReplaceHelmVersionDocs(helm_v3.NewVersionCmd(os.Stdout))
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		t.Error("There is no annotation!")
	} else {
		if ann != GetHelmVersionDocs().LongMD {
			t.Error("The annotation does not match!")
		}
	}
}

func TestReplaceHelmDependencyDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmDependencyDocs(helm_v3.NewDependencyCmd(actionConfig, os.Stdout))
	for i, c := range cmd.Commands() {
		if ann, ok := c.Annotations[common.DocsLongMD]; !ok {
			t.Errorf("There is no annotation in `%s -> %s` command!", cmd.Use, cmd.Commands()[i].Use)
		} else {
			if ann != GetHelmDependencyBuildDocs().LongMD &&
				ann != GetHelmDependencyListDocs().LongMD &&
				ann != GetHelmDependencyUpdateDocs().LongMD {
				t.Errorf("The annotation in `%s -> %s` command does not match!", cmd.Use, cmd.Commands()[i].Use)
			}
		}
	}
}

func TestReplaceHelmGetDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmGetDocs(helm_v3.NewGetCmd(actionConfig, os.Stdout))
	for i, c := range cmd.Commands() {
		if ann, ok := c.Annotations[common.DocsLongMD]; !ok {
			t.Errorf("There is no annotation in `%s -> %s` command!", cmd.Use, cmd.Commands()[i].Use)
		} else {
			if ann != GetHelmGetHooksDocs().LongMD &&
				ann != GetHelmGetAllDocs().LongMD &&
				ann != GetHelmGetValuesDocs().LongMD &&
				ann != GetHelmGetManifestDocs().LongMD &&
				ann != GetHelmGetNotesDocs().LongMD &&
				ann != GetHelmGetMetadataDocs().LongMD {
				t.Errorf("The annotation in `%s -> %s` command does not match!", cmd.Use, cmd.Commands()[i].Use)
			}
		}
	}
}

func TestReplaceHelmPluginDocs(t *testing.T) {
	cmd := ReplaceHelmPluginDocs(helm_v3.NewPluginCmd(os.Stdout))
	for i, c := range cmd.Commands() {
		if ann, ok := c.Annotations[common.DocsLongMD]; !ok {
			t.Errorf("There is no annotation in `%s -> %s` command!", cmd.Use, cmd.Commands()[i].Use)
		} else {
			if ann != GetHelmPluginListDocs().LongMD &&
				ann != GetHelmPluginUninstallDocs().LongMD &&
				ann != GetHelmPluginUpdateDocs().LongMD &&
				ann != GetHelmPluginInstallDocs().LongMD {
				t.Errorf("The annotation in `%s -> %s` command does not match!", cmd.Use, cmd.Commands()[i].Use)
			}
		}
	}
}

func TestReplaceHelmRepoDocs(t *testing.T) {
	cmd := ReplaceHelmRepoDocs(helm_v3.NewRepoCmd(os.Stdout))
	for i, c := range cmd.Commands() {
		if ann, ok := c.Annotations[common.DocsLongMD]; !ok {
			t.Errorf("There is no annotation in `%s -> %s` command!", cmd.Use, cmd.Commands()[i].Use)
		} else {
			if ann != GetHelmRepoAddDocs().LongMD &&
				ann != GetHelmRepoIndexDocs().LongMD &&
				ann != GetHelmRepoListDocs().LongMD &&
				ann != GetHelmRepoRemoveDocs().LongMD &&
				ann != GetHelmRepoUpdateDocs().LongMD {
				t.Errorf("The annotation in `%s -> %s` command does not match!", cmd.Use, cmd.Commands()[i].Use)
			}
		}
	}
}

func TestReplaceHelmSearchDocs(t *testing.T) {
	cmd := ReplaceHelmSearchDocs(helm_v3.NewSearchCmd(os.Stdout))
	for i, c := range cmd.Commands() {
		if ann, ok := c.Annotations[common.DocsLongMD]; !ok {
			t.Errorf("There is no annotation in `%s -> %s` command!", cmd.Use, cmd.Commands()[i].Use)
		} else {
			if ann != GetHelmSearchHubDocs().LongMD &&
				ann != GetHelmSearchRepoDocs().LongMD {
				t.Errorf("The annotation in `%s -> %s` command does not match!", cmd.Use, cmd.Commands()[i].Use)
			}
		}
	}
}

func TestReplaceHelmShowDocs(t *testing.T) {
	actionConfig := new(action.Configuration)
	cmd := ReplaceHelmShowDocs(helm_v3.NewShowCmd(actionConfig, os.Stdout))
	for i, c := range cmd.Commands() {
		if ann, ok := c.Annotations[common.DocsLongMD]; !ok {
			t.Errorf("There is no annotation in `%s -> %s` command!", cmd.Use, cmd.Commands()[i].Use)
		} else {
			if ann != GetHelmShowAllDocs().LongMD &&
				ann != GetHelmShowChartDocs().LongMD &&
				ann != GetHelmShowCRDsDocs().LongMD &&
				ann != GetHelmShowReadmeDocs().LongMD &&
				ann != GetHelmShowValuesDocs().LongMD {
				t.Errorf("The annotation in `%s -> %s` command does not match!", cmd.Use, cmd.Commands()[i].Use)
			}
		}
	}
}
