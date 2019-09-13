package builder

//go:generate esc -no-compress -ignore static.go -o ansible/static.go -pkg ansible ansible

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ghodssYaml "github.com/ghodss/yaml"
	"gopkg.in/oleiade/reflections.v1"
	"gopkg.in/yaml.v2"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/stapel"
	"github.com/flant/werf/pkg/util"
)

type Ansible struct {
	config *config.Ansible
	extra  *Extra
}

type Extra struct {
	ContainerWerfPath string
	TmpPath           string
}

func NewAnsibleBuilder(config *config.Ansible, extra *Extra) *Ansible {
	return &Ansible{config: config, extra: extra}
}

func (b *Ansible) IsBeforeInstallEmpty() bool { return b.isEmptyStage("BeforeInstall") }
func (b *Ansible) IsInstallEmpty() bool       { return b.isEmptyStage("Install") }
func (b *Ansible) IsBeforeSetupEmpty() bool   { return b.isEmptyStage("BeforeSetup") }
func (b *Ansible) IsSetupEmpty() bool         { return b.isEmptyStage("Setup") }

func (b *Ansible) BeforeInstall(container Container) error { return b.stage("BeforeInstall", container) }
func (b *Ansible) Install(container Container) error       { return b.stage("Install", container) }
func (b *Ansible) BeforeSetup(container Container) error   { return b.stage("BeforeSetup", container) }
func (b *Ansible) Setup(container Container) error         { return b.stage("Setup", container) }

func (b *Ansible) BeforeInstallChecksum() string { return b.stageChecksum("BeforeInstall") }
func (b *Ansible) InstallChecksum() string       { return b.stageChecksum("Install") }
func (b *Ansible) BeforeSetupChecksum() string   { return b.stageChecksum("BeforeSetup") }
func (b *Ansible) SetupChecksum() string         { return b.stageChecksum("Setup") }

func (b *Ansible) isEmptyStage(userStageName string) bool {
	return b.stageChecksum(userStageName) == ""
}

func (b *Ansible) stage(userStageName string, container Container) error {
	if len(b.stageTasks(userStageName)) == 0 {
		return nil
	}

	if err := b.createStageWorkDirStructure(userStageName); err != nil {
		return err
	}

	container.AddEnv(
		map[string]string{
			"ANSIBLE_CONFIG":            filepath.Join(b.containerWorkDir(), "ansible.cfg"),
			"WERF_DUMP_CONFIG_DOC_PATH": filepath.Join(b.containerWorkDir(), "dump_config.json"),
			"PYTHONIOENCODING":          "utf-8",
			"PATH":                      fmt.Sprintf("%s:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:%s", stapel.AnsibleToolsOverlayPATH(), stapel.SystemPATH()),
			"LD_LIBRARY_PATH":           stapel.AnsibleLibsOverlayLDPATH(),
			"LANG":                      "C.UTF-8",
			"LC_ALL":                    "C.UTF-8",
		},
	)

	stageHostWorkDir, err := b.stageHostWorkDir(userStageName)
	if err != nil {
		return err
	}

	stageHostTmpDir, err := b.stageHostTmpDir(userStageName)
	if err != nil {
		return err
	}

	container.AddVolume(
		fmt.Sprintf("%s:%s:ro", stageHostWorkDir, b.containerWorkDir()),
		fmt.Sprintf("%s:%s:rw", stageHostTmpDir, b.containerTmpDir()),
	)

	containerName, err := stapel.GetOrCreateContainer()
	if err != nil {
		return err
	}
	container.AddVolumeFrom(fmt.Sprintf("%s:ro", containerName))

	commandParts := []string{
		filepath.Join(b.containerWorkDir(), "ansible-playbook"),
		filepath.Join(b.containerWorkDir(), "playbook.yml"),
	}

	if value, exist := os.LookupEnv("WERF_DEBUG_ANSIBLE_ARGS"); exist {
		commandParts = append(commandParts, value)
	}

	command := strings.Join(commandParts, " ")
	container.AddServiceRunCommands(command)

	return nil
}

func (b *Ansible) stageChecksum(userStageName string) string {
	var checksumArgs []string

	for _, task := range b.stageTasks(userStageName) {
		output, err := yaml.Marshal(task.Config)
		if err != nil {
			panic(fmt.Sprintf("runtime err: %s", err))
		}

		jsonOutput, err := ghodssYaml.YAMLToJSON(output)
		if err != nil {
			panic(fmt.Sprintf("runtime err: %s", err))
		}
		checksumArgs = append(checksumArgs, string(jsonOutput))
	}

	if debugUserStageChecksum() {
		logboek.LogHighlightF("DEBUG: %s stage tasks checksum dependencies %v\n", userStageName, checksumArgs)
	}

	if stageVersionChecksum := b.stageVersionChecksum(userStageName); stageVersionChecksum != "" {
		if debugUserStageChecksum() {
			logboek.LogHighlightF("DEBUG: %s stage version checksum %v\n", userStageName, stageVersionChecksum)
		}

		checksumArgs = append(checksumArgs, stageVersionChecksum)
	}

	if len(checksumArgs) != 0 {
		return util.Sha256Hash(checksumArgs...)
	} else {
		return ""
	}
}

func (b *Ansible) stageVersionChecksum(userStageName string) string {
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

func (b *Ansible) stageTasks(userStageName string) []*config.AnsibleTask {
	value := b.configFieldValue(userStageName)
	ansibleTasks, ok := value.([]*config.AnsibleTask)
	if !ok {
		panic(fmt.Sprintf("runtime error"))
	}

	return ansibleTasks
}

func (b *Ansible) configFieldValue(fieldName string) interface{} {
	value, err := reflections.GetField(b.config, fieldName)
	if err != nil {
		panic(fmt.Sprintf("runtime error: %s", err))
	}

	return value
}

func (b *Ansible) stageHostWorkDir(userStageName string) (string, error) {
	path := filepath.Join(b.extra.TmpPath, fmt.Sprintf("ansible-workdir-%s", userStageName))

	if err := mkdirP(path); err != nil {
		return "", err
	}

	return path, nil
}
