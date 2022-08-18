package builder

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/oleiade/reflections.v1"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/container_backend/stage_builder"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/util"
)

const scriptFileName = "script.sh"

type Shell struct {
	config *config.Shell
	extra  *Extra
}

func NewShellBuilder(config *config.Shell, extra *Extra) *Shell {
	return &Shell{config: config, extra: extra}
}

func (b *Shell) IsBeforeInstallEmpty(ctx context.Context) bool {
	return b.isEmptyStage(ctx, "BeforeInstall")
}
func (b *Shell) IsInstallEmpty(ctx context.Context) bool { return b.isEmptyStage(ctx, "Install") }
func (b *Shell) IsBeforeSetupEmpty(ctx context.Context) bool {
	return b.isEmptyStage(ctx, "BeforeSetup")
}
func (b *Shell) IsSetupEmpty(ctx context.Context) bool { return b.isEmptyStage(ctx, "Setup") }

func (b *Shell) BeforeInstall(_ context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool) error {
	return b.stage(cr, stageBuilder, useLegacyStapelBuilder, "BeforeInstall")
}

func (b *Shell) Install(_ context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool) error {
	return b.stage(cr, stageBuilder, useLegacyStapelBuilder, "Install")
}

func (b *Shell) BeforeSetup(_ context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool) error {
	return b.stage(cr, stageBuilder, useLegacyStapelBuilder, "BeforeSetup")
}

func (b *Shell) Setup(_ context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool) error {
	return b.stage(cr, stageBuilder, useLegacyStapelBuilder, "Setup")
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

func (b *Shell) stage(cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool, userStageName string) error {
	if useLegacyStapelBuilder {
		container := stageBuilder.LegacyStapelStageBuilder().BuilderContainer()

		stageHostTmpDir, err := b.stageHostTmpDir(userStageName)
		if err != nil {
			return err
		}

		container.AddVolume(
			fmt.Sprintf("%s:%s:rw", stageHostTmpDir, b.containerTmpDir()),
		)

		stageHostTmpScriptFilePath := filepath.Join(stageHostTmpDir, scriptFileName)
		containerTmpScriptFilePath := path.Join(b.containerTmpDir(), scriptFileName)

		if err := stapel.CreateScript(stageHostTmpScriptFilePath, b.stageCommands(userStageName)); err != nil {
			return err
		}

		container.AddServiceRunCommands(containerTmpScriptFilePath)
	} else {
		stageBuilder.StapelStageBuilder().AddCommands(b.stageCommands(userStageName)...)
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

	if err := mkdirP(p); err != nil {
		return "", err
	}

	return p, nil
}

func (b *Shell) containerTmpDir() string {
	return path.Join(b.extra.ContainerWerfPath, "shell")
}
