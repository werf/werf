package secret

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	secret_common "github.com/flant/werf/cmd/werf/helm/secret/common"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	Values bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "edit FILE_PATH",
		DisableFlagsInUseLine: true,
		Short:                 "Edit or create new secret file",
		Long: common.GetLongCommandDescription(`Edit or create new secret file.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file`),
		Example: `  # Create/edit existing secret file
  $ werf helm secret file edit .helm/secret/privacy`,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
				return err
			}

			return runSecretEdit(args[0])
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	return cmd
}

func runSecretEdit(filepPath string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	m, err := secret.GetManager(projectDir)
	if err != nil {
		return err
	}

	return secret_common.SecretEdit(m, filepPath, false)
}
