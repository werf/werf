package helm_v3

import (
	"github.com/werf/werf/cmd/werf/helm_v3/create"
	"github.com/werf/werf/cmd/werf/helm_v3/dependency"
	"github.com/werf/werf/cmd/werf/helm_v3/env"
	"github.com/werf/werf/cmd/werf/helm_v3/get"
	"github.com/werf/werf/cmd/werf/helm_v3/history"
	"github.com/werf/werf/cmd/werf/helm_v3/install"
	"github.com/werf/werf/cmd/werf/helm_v3/lint"
	"github.com/werf/werf/cmd/werf/helm_v3/list"
	package_ "github.com/werf/werf/cmd/werf/helm_v3/package"
	"github.com/werf/werf/cmd/werf/helm_v3/plugin"
	"github.com/werf/werf/cmd/werf/helm_v3/pull"
	"github.com/werf/werf/cmd/werf/helm_v3/repo"
	"github.com/werf/werf/cmd/werf/helm_v3/rollback"
	"github.com/werf/werf/cmd/werf/helm_v3/search"
	"github.com/werf/werf/cmd/werf/helm_v3/show"
	"github.com/werf/werf/cmd/werf/helm_v3/status"
	"github.com/werf/werf/cmd/werf/helm_v3/template"
	"github.com/werf/werf/cmd/werf/helm_v3/test"
	"github.com/werf/werf/cmd/werf/helm_v3/uninstall"
	"github.com/werf/werf/cmd/werf/helm_v3/upgrade"
	"github.com/werf/werf/cmd/werf/helm_v3/verify"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm-v3",
		Short: "Manage application deployment with helm",
	}
	cmd.AddCommand(
		uninstall.NewCmd(),
		dependency.NewCmd(),
		get.NewCmd(),
		history.NewCmd(),
		lint.NewCmd(),
		list.NewCmd(),
		template.NewCmd(),
		repo.NewCmd(),
		rollback.NewCmd(),
		install.NewCmd(),
		upgrade.NewCmd(),
		create.NewCmd(),
		env.NewCmd(),
		package_.NewCmd(),
		plugin.NewCmd(),
		pull.NewCmd(),
		search.NewCmd(),
		show.NewCmd(),
		status.NewCmd(),
		test.NewCmd(),
		verify.NewCmd(),
	)

	return cmd
}
