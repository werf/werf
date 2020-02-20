package builder

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/oleiade/reflections.v1"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/stapel"
	"github.com/flant/werf/pkg/util"
)

const scriptFileName = "script.sh"

type Shell struct {
	config *config.Shell
	extra  *Extra
}

func NewShellBuilder(config *config.Shell, extra *Extra) *Shell {
	return &Shell{config: config, extra: extra}
}

func (b *Shell) IsBeforeInstallEmpty() bool { return b.isEmptyStage("BeforeInstall") }
func (b *Shell) IsInstallEmpty() bool       { return b.isEmptyStage("Install") }
func (b *Shell) IsBeforeSetupEmpty() bool   { return b.isEmptyStage("BeforeSetup") }
func (b *Shell) IsSetupEmpty() bool         { return b.isEmptyStage("Setup") }

func (b *Shell) BeforeInstall(container Container) error { return b.stage("BeforeInstall", container) }
func (b *Shell) Install(container Container) error       { return b.stage("Install", container) }
func (b *Shell) BeforeSetup(container Container) error   { return b.stage("BeforeSetup", container) }
func (b *Shell) Setup(container Container) error         { return b.stage("Setup", container) }

func (b *Shell) BeforeInstallChecksum() string { return b.stageChecksum("BeforeInstall") }
func (b *Shell) InstallChecksum() string       { return b.stageChecksum("Install") }
func (b *Shell) BeforeSetupChecksum() string   { return b.stageChecksum("BeforeSetup") }
func (b *Shell) SetupChecksum() string         { return b.stageChecksum("Setup") }

func (b *Shell) isEmptyStage(userStageName string) bool {
	return b.stageChecksum(userStageName) == ""
}

func (b *Shell) stage(userStageName string, container Container) error {
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
	return nil
}

func (b *Shell) stageChecksum(userStageName string) string {
	var checksumArgs []string

	checksumArgs = append(checksumArgs, b.stageCommands(userStageName)...)

	if debugUserStageChecksum() {
		logboek.Debug.LogFHighlight("DEBUG: %s stage tasks checksum dependencies %v\n", userStageName, checksumArgs)
	}

	if stageVersionChecksum := b.stageVersionChecksum(userStageName); stageVersionChecksum != "" {
		if debugUserStageChecksum() {
			logboek.Debug.LogFHighlight("DEBUG: %s stage version checksum %v\n", userStageName, stageVersionChecksum)
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
