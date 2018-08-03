package builder

import (
	"fmt"
	"strings"

	"gopkg.in/oleiade/reflections.v1"

	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
)

type Shell struct{ config config.Shell }

func NewShellBuilder(config config.Shell) *Shell {
	return &Shell{config}
}

func (b *Shell) IsBeforeInstallEmpty() bool { return b.isEmptyStage("BeforeInstall") }
func (b *Shell) IsInstallEmpty() bool       { return b.isEmptyStage("Install") }
func (b *Shell) IsBeforeSetupEmpty() bool   { return b.isEmptyStage("BeforeSetup") }
func (b *Shell) IsSetupEmpty() bool         { return b.isEmptyStage("Setup") }
func (b *Shell) IsBuildArtifactEmpty() bool { return b.isEmptyStage("BuildArtifact") }

func (b *Shell) BeforeInstall(container Container) error { return b.stage("BeforeInstall", container) }
func (b *Shell) Install(container Container) error       { return b.stage("Install", container) }
func (b *Shell) BeforeSetup(container Container) error   { return b.stage("BeforeSetup", container) }
func (b *Shell) Setup(container Container) error         { return b.stage("Setup", container) }
func (b *Shell) BuildArtifact(container Container) error { return b.stage("BuildArtifact", container) }

func (b *Shell) BeforeInstallChecksum() string { return b.stageChecksum("BeforeInstall") }
func (b *Shell) InstallChecksum() string       { return b.stageChecksum("Install") }
func (b *Shell) BeforeSetupChecksum() string   { return b.stageChecksum("BeforeSetup") }
func (b *Shell) SetupChecksum() string         { return b.stageChecksum("Setup") }
func (b *Shell) BuildArtifactChecksum() string { return b.stageChecksum("BuildArtifact") }

func (b *Shell) isEmptyStage(userStageName string) bool {
	return b.stageChecksum(userStageName) == ""
}

func (b *Shell) stage(userStageName string, container Container) error {
	container.AddRunCommands(b.stageCommands(userStageName))
	return nil
}

func (b *Shell) stageChecksum(userStageName string) string {
	var checksumArgs []string
	if stageVersionChecksum := b.stageVersionChecksum(userStageName); stageVersionChecksum != "" {
		checksumArgs = append(checksumArgs, stageVersionChecksum)
	}
	checksumArgs = append(checksumArgs, b.stageCommands(userStageName)...)

	if len(checksumArgs) != 0 {
		return util.MurmurHash(strings.Join(checksumArgs, ""))
	} else {
		return ""
	}
}

func (b *Shell) stageVersionChecksum(userStageName string) string {
	var stageVersionChecksumArgs []string

	cacheVersionFieldName := "CacheVersion"
	checksum, ok := b.configFieldValue(cacheVersionFieldName).(string)
	if !ok {
		panic(fmt.Sprintf("runtime error: %#v", checksum))
	}

	if checksum != "" {
		stageVersionChecksumArgs = append(stageVersionChecksumArgs, checksum)
	}

	stageCacheVersionFieldName := strings.Join([]string{userStageName, cacheVersionFieldName}, "")
	stageChecksum, ok := b.configFieldValue(stageCacheVersionFieldName).(string)
	if !ok {
		panic(fmt.Sprintf("runtime error: %#v", stageChecksum))
	}

	if stageChecksum != "" {
		stageVersionChecksumArgs = append(stageVersionChecksumArgs, stageChecksum)
	}

	if len(stageVersionChecksumArgs) != 0 {
		return util.MurmurHash(strings.Join(stageVersionChecksumArgs, ""))
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
