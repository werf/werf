package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/oleiade/reflections.v1"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/build/secrets"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/stage_builder"
)

type Extra struct {
	TmpPath string
}

type Shell struct {
	config      *config.Shell
	extra       *Extra
	secrets     []config.Secret
	sshAuthSock string
}

func NewShellBuilder(config *config.Shell, extra *Extra, secrets []config.Secret, sshAuthSock string) *Shell {
	return &Shell{config: config, extra: extra, secrets: secrets, sshAuthSock: sshAuthSock}
}

func (b *Shell) IsBeforeInstallEmpty(ctx context.Context) bool {
	return b.isEmptyStage(ctx, "BeforeInstall")
}
func (b *Shell) IsInstallEmpty(ctx context.Context) bool { return b.isEmptyStage(ctx, "Install") }
func (b *Shell) IsBeforeSetupEmpty(ctx context.Context) bool {
	return b.isEmptyStage(ctx, "BeforeSetup")
}
func (b *Shell) IsSetupEmpty(ctx context.Context) bool { return b.isEmptyStage(ctx, "Setup") }

func (b *Shell) BeforeInstall(_ context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface) error {
	return b.stage(cr, stageBuilder, "BeforeInstall")
}

func (b *Shell) Install(_ context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface) error {
	return b.stage(cr, stageBuilder, "Install")
}

func (b *Shell) BeforeSetup(_ context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface) error {
	return b.stage(cr, stageBuilder, "BeforeSetup")
}

func (b *Shell) Setup(_ context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface) error {
	return b.stage(cr, stageBuilder, "Setup")
}

func (b *Shell) BeforeInstallChecksum(ctx context.Context) string {
	return b.stageChecksum(ctx, "BeforeInstall")
}
func (b *Shell) InstallChecksum(ctx context.Context) string { return b.stageChecksum(ctx, "Install") }
func (b *Shell) BeforeSetupChecksum(ctx context.Context) string {
	return b.stageChecksum(ctx, "BeforeSetup")
}
func (b *Shell) SetupChecksum(ctx context.Context) string { return b.stageChecksum(ctx, "Setup") }

func (b *Shell) isEmptyStage(ctx context.Context, userStageName string) bool {
	return b.stageChecksum(ctx, userStageName) == ""
}

func (b *Shell) stage(cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, userStageName string) error {
	stageHostTmpDir, err := b.stageHostTmpDir(userStageName)
	if err != nil {
		return err
	}

	stageBuilder.StapelStageBuilder().MountSSHAgentSocket(b.sshAuthSock)
	stageBuilder.StapelStageBuilder().AddCommands(b.stageCommands(userStageName)...)

	if err := b.addBuildSecretsVolumes(stageHostTmpDir, func(secretPath string) {
		stageBuilder.StapelStageBuilder().AddBuildVolumes(secretPath)
	}); err != nil {
		return fmt.Errorf("unable to add volumes: %w", err)
	}

	return nil
}

func (b *Shell) stageChecksum(ctx context.Context, userStageName string) string {
	var checksumArgs []string

	checksumArgs = append(checksumArgs, b.stageCommands(userStageName)...)

	if debugUserStageChecksum() {
		logboek.Context(ctx).Debug().LogFHighlight("DEBUG: %s stage tasks checksum dependencies %v\n", userStageName, checksumArgs)
	}

	if stageVersionChecksum := b.stageVersionChecksum(userStageName); stageVersionChecksum != "" {
		if debugUserStageChecksum() {
			logboek.Context(ctx).Debug().LogFHighlight("DEBUG: %s stage version checksum %v\n", userStageName, stageVersionChecksum)
		}
		checksumArgs = append(checksumArgs, stageVersionChecksum)
	}

	if len(checksumArgs) != 0 {
		return util.Sha256Hash(checksumArgs...)
	} else {
		return ""
	}
}

func (b *Shell) stageVersionChecksum(userStageName string) string {
	var stageVersionChecksumArgs []string

	cacheVersionFieldName := "CacheVersion"
	stageCacheVersionFieldName := strings.Join([]string{userStageName, cacheVersionFieldName}, "")

	stageChecksum, ok := b.configFieldValue(stageCacheVersionFieldName).(string)
	if !ok {
		panic(fmt.Sprintf("runtime error: %#v", stageChecksum))
	}

	if stageChecksum != "" {
		stageVersionChecksumArgs = append(stageVersionChecksumArgs, stageChecksum)
	}

	checksum, ok := b.configFieldValue(cacheVersionFieldName).(string)
	if !ok {
		panic(fmt.Sprintf("runtime error: %#v", checksum))
	}

	if checksum != "" {
		stageVersionChecksumArgs = append(stageVersionChecksumArgs, checksum)
	}

	if len(stageVersionChecksumArgs) != 0 {
		return util.Sha256Hash(stageVersionChecksumArgs...)
	} else {
		return ""
	}
}

func (b *Shell) stageCommands(userStageName string) []string {
	commands, err := util.InterfaceToStringArray(b.configFieldValue(userStageName))
	if err != nil {
		panic(fmt.Sprintf("runtime error: %s", err))
	}

	return commands
}

func (b *Shell) configFieldValue(fieldName string) interface{} {
	value, err := reflections.GetField(b.config, fieldName)
	if err != nil {
		panic(fmt.Sprintf("runtime error: %s", err))
	}

	return value
}

func (b *Shell) stageHostTmpDir(userStageName string) (string, error) {
	p := filepath.Join(b.extra.TmpPath, fmt.Sprintf("shell-%s", userStageName))

	if err := os.MkdirAll(p, os.FileMode(0o775)); err != nil {
		return "", err
	}

	return p, nil
}

func (b *Shell) addBuildSecretsVolumes(stageHostTmpDir string, fn func(string)) error {
	for _, s := range b.secrets {
		secretPath, err := secrets.GetMountPath(s, stageHostTmpDir)
		if err != nil {
			return err
		}
		fn(secretPath)
	}
	return nil
}
