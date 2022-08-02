package builder

//go:generate esc -no-compress -ignore static.go -o ansible/static.go -pkg ansible ansible

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	ghodssYaml "github.com/ghodss/yaml"
	"gopkg.in/oleiade/reflections.v1"
	"gopkg.in/yaml.v2"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/container_backend/stage_builder"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/util"
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

func (b *Ansible) IsBeforeInstallEmpty(ctx context.Context) bool {
	return b.isEmptyStage(ctx, "BeforeInstall")
}
func (b *Ansible) IsInstallEmpty(ctx context.Context) bool { return b.isEmptyStage(ctx, "Install") }
func (b *Ansible) IsBeforeSetupEmpty(ctx context.Context) bool {
	return b.isEmptyStage(ctx, "BeforeSetup")
}
func (b *Ansible) IsSetupEmpty(ctx context.Context) bool { return b.isEmptyStage(ctx, "Setup") }

func (b *Ansible) BeforeInstall(ctx context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool) error {
	return b.stage(ctx, cr, stageBuilder, useLegacyStapelBuilder, "BeforeInstall")
}

func (b *Ansible) Install(ctx context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool) error {
	return b.stage(ctx, cr, stageBuilder, useLegacyStapelBuilder, "Install")
}

func (b *Ansible) BeforeSetup(ctx context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool) error {
	return b.stage(ctx, cr, stageBuilder, useLegacyStapelBuilder, "BeforeSetup")
}

func (b *Ansible) Setup(ctx context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool) error {
	return b.stage(ctx, cr, stageBuilder, useLegacyStapelBuilder, "Setup")
}

func (b *Ansible) BeforeInstallChecksum(ctx context.Context) string {
	return b.stageChecksum(ctx, "BeforeInstall")
}
func (b *Ansible) InstallChecksum(ctx context.Context) string { return b.stageChecksum(ctx, "Install") }
func (b *Ansible) BeforeSetupChecksum(ctx context.Context) string {
	return b.stageChecksum(ctx, "BeforeSetup")
}
func (b *Ansible) SetupChecksum(ctx context.Context) string { return b.stageChecksum(ctx, "Setup") }

func (b *Ansible) isEmptyStage(ctx context.Context, userStageName string) bool {
	return b.stageChecksum(ctx, userStageName) == ""
}

func (b *Ansible) stage(ctx context.Context, cr container_backend.ContainerBackend, stageBuilder stage_builder.StageBuilderInterface, useLegacyStapelBuilder bool, userStageName string) error {
	if useLegacyStapelBuilder {
		container := stageBuilder.LegacyStapelStageBuilder().BuilderContainer()

		if len(b.stageTasks(userStageName)) == 0 {
			return nil
		}

		if err := b.createStageWorkDirStructure(userStageName); err != nil {
			return err
		}

		container.AddEnv(
			map[string]string{
				"ANSIBLE_CONFIG":              path.Join(b.containerWorkDir(), "ansible.cfg"),
				"WERF_DUMP_CONFIG_DOC_PATH":   path.Join(b.containerWorkDir(), "dump_config.json"),
				"PYTHONIOENCODING":            "utf-8",
				"ANSIBLE_PREPEND_SYSTEM_PATH": stapel.AnsibleToolsOverlayPATH(),
				"ANSIBLE_APPEND_SYSTEM_PATH":  stapel.SystemPATH(),
				"LD_LIBRARY_PATH":             stapel.AnsibleLibsOverlayLDPATH(),
				"LANG":                        "C.UTF-8",
				"LC_ALL":                      "C.UTF-8",
				"LOGBOEK_SO_PATH":             "/.werf/stapel/embedded/lib/python2.7/_logboek.so",
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

		containerName, err := stapel.GetOrCreateContainer(ctx)
		if err != nil {
			return err
		}
		container.AddVolumeFrom(fmt.Sprintf("%s:ro", containerName))

		commandParts := []string{
			path.Join(b.containerWorkDir(), "ansible-playbook"),
			path.Join(b.containerWorkDir(), "playbook.yml"),
		}

		if value, exist := os.LookupEnv("WERF_DEBUG_ANSIBLE_ARGS"); exist {
			commandParts = append(commandParts, value)
		}

		command := strings.Join(commandParts, " ")
		container.AddServiceRunCommands(command)

		return nil
	} else {
		return fmt.Errorf("ansible builder is not supported when using buildah backend, please use shell builder instead")
	}
}

func (b *Ansible) stageChecksum(ctx context.Context, userStageName string) string {
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
		panic("runtime error")
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
	p := filepath.Join(b.extra.TmpPath, fmt.Sprintf("ansible-workdir-%s", userStageName))

	if err := mkdirP(p); err != nil {
		return "", err
	}

	return p, nil
}
