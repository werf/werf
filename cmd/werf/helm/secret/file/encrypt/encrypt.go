package secret

import (
	"cmp"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/nelm/pkg/action"
	secret_common "github.com/werf/nelm/pkg/legacy/secret"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/cmd/werf/docs/replacers/helm"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/werf"
)

var cmdData struct {
	OutputFilePath string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "encrypt [FILE_PATH]",
		DisableFlagsInUseLine: true,
		Short:                 "Encrypt file data",
		Long:                  common.GetLongCommandDescription(helm.GetHelmSecretFileEncryptDocs().Long),
		Example: `  # Encrypt and save result in file
  $ werf helm secret file encrypt tls.crt -o .helm/secret/tls.crt`,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
			common.DocsLongMD: helm.GetHelmSecretFileEncryptDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			var filePath string

			if len(args) > 0 {
				filePath = args[0]
			}

			if err := runSecretEncrypt(ctx, filePath); err != nil {
				if strings.HasSuffix(err.Error(), secret_common.ExpectedFilePathOrPipeError().Error()) {
					common.PrintHelp(cmd)
				}

				return err
			}

			return nil
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.OutputFilePath, "output-file-path", "o", "", "Write to file instead of stdout")

	return cmd
}

func runSecretEncrypt(ctx context.Context, filePath string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	workingDir := common.GetWorkingDir(&commonCmdData)

	ctx = log.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultSecretFileEncryptLogLevel), log.SetupLoggingOptions{
		ColorMode:      *commonCmdData.LogColorMode,
		LogIsParseable: true,
	})

	if err := action.SecretFileEncrypt(ctx, filePath, action.SecretFileEncryptOptions{
		OutputFilePath: cmdData.OutputFilePath,
		SecretWorkDir:  workingDir,
		TempDirPath:    *commonCmdData.TmpDir,
	}); err != nil {
		return fmt.Errorf("secret file encrypt: %w", err)
	}

	return nil
}
