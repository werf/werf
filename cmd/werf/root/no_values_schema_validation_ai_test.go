//go:build ai_tests

package root

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	bundle_apply "github.com/werf/werf/v2/cmd/werf/bundle/apply"
	bundle_plan "github.com/werf/werf/v2/cmd/werf/bundle/plan"
	bundle_render "github.com/werf/werf/v2/cmd/werf/bundle/render"
	"github.com/werf/werf/v2/cmd/werf/converge"
	"github.com/werf/werf/v2/cmd/werf/lint"
	"github.com/werf/werf/v2/cmd/werf/plan"
	"github.com/werf/werf/v2/cmd/werf/render"
	"github.com/werf/werf/v2/cmd/werf/rollback"
)

func TestAI_NoValuesSchemaValidationFlagOnAllRenderingCommands(t *testing.T) {
	ctx := context.Background()

	commands := map[string]func(context.Context) *cobra.Command{
		"render":        render.NewCmd,
		"bundle render": bundle_render.NewCmd,
		"converge":      converge.NewCmd,
		"plan":          plan.NewCmd,
		"lint":          lint.NewCmd,
		"rollback":      rollback.NewCmd,
		"bundle apply":  bundle_apply.NewCmd,
		"bundle plan":   bundle_plan.NewCmd,
	}

	for name, newCmd := range commands {
		cmd := newCmd(ctx)
		if cmd.Flags().Lookup("no-values-schema-validation") == nil {
			t.Errorf("command %q is missing the --no-values-schema-validation flag", name)
		}
	}
}
