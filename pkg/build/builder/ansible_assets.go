package builder

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flant/dapp/pkg/util"
)

func (b *Ansible) createStageWorkDirStructure(userStageName string) error {
	stageWorkDir, err := b.stageHostWorkDir(userStageName)
	if err != nil {
		return err
	}

	// playbook with tasks for a stage
	stagePlaybook, err := b.stagePlaybook(userStageName)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(stagePlaybook)
	if err != nil {
		return err
	}
	writeFile(filepath.Join(stageWorkDir, "playbook.yml"), string(data))

	// generate inventory with localhost and python in dappdeps-ansible
	writeFile(filepath.Join(stageWorkDir, "hosts"), b.assetsHosts())

	// generate ansible config for solo mode
	writeFile(filepath.Join(stageWorkDir, "ansible.cfg"), b.assetsAnsibleCfg())

	// save config dump for pretty errors
	stageConfig, err := b.stageConfig(userStageName)
	if err != nil {
		return err
	}
	data, err = json.Marshal(stageConfig["dump_config"])
	if err != nil {
		return err
	}
	writeFile(filepath.Join(stageWorkDir, "dump_config.json"), string(data))

	// python modules
	stageWorkDirLib := filepath.Join(stageWorkDir, "lib")
	if err := mkdirP(stageWorkDirLib); err != nil {
		return err
	}

	// crypt.py hack
	// TODO must be in dappdeps-ansible
	writeFile(filepath.Join(stageWorkDirLib, "crypt.py"), b.assetsCryptPy())

	stageCallbackDir := filepath.Join(stageWorkDirLib, "callback")
	if err := mkdirP(stageCallbackDir); err != nil {
		return err
	}

	writeFile(filepath.Join(stageCallbackDir, "__init__.py"), "# module callback")

	writeFile(filepath.Join(stageCallbackDir, "live.py"), b.assetsLivePy())

	// add dapp specific stdout callback for ansible
	writeFile(filepath.Join(stageCallbackDir, "dapp.py"), b.assetsDappPy())

	return nil
}

func (b *Ansible) stagePlaybook(userStageName string) ([]map[string]interface{}, error) {
	playbook := map[string]interface{}{
		"hosts":        "all",
		"gather_facts": "no",
	}
	stageConfig, err := b.stageConfig(userStageName)
	if err != nil {
		return nil, err
	}
	playbook["tasks"] = stageConfig["tasks"]
	playbooks := []map[string]interface{}{playbook}
	return playbooks, nil
}

//query tasks from ansible config
//create dump_config structure
//returns structure:
//{ 'tasks' => [array of tasks for stage],
//'dump_config' => {
//'dump_config_doc' => 'dump of doc',
//'dump_config_sections' => {'task_0'=>'dump for task 0', 'task_1'=>'dump for task 1', ... }}
func (b *Ansible) stageConfig(userStageName string) (map[string]interface{}, error) {
	dumpConfigSections := map[string]interface{}{}
	var tasks []interface{}
	for ind, ansibleTask := range b.stageTasks(userStageName) {
		task, err := util.InterfaceToMapStringInterface(ansibleTask.Config)
		if err != nil {
			return nil, err
		}

		var tags []string
		if _, ok := task["tags"]; ok {
			tags, err = util.InterfaceToStringArray(task["tags"])
			if err != nil {
				return nil, err
			}
		}
		dumpTag := fmt.Sprintf("task_%d", ind)
		tags = append(tags, dumpTag)
		task["tags"] = tags

		dumpConfigSections[dumpTag] = ansibleTask.DumpConfigSection

		tasks = append(tasks, task)
	}

	dumpConfig := map[string]interface{}{
		"dump_config_doc":      b.config.DumpConfigSection,
		"dump_config_sections": dumpConfigSections,
	}

	result := map[string]interface{}{
		"dump_config": dumpConfig,
		"tasks":       tasks,
	}

	return result, nil
}

func (b *Ansible) stageHostTmpDir(userStageName string) (string, error) {
	path := filepath.Join(b.extra.TmpPath, fmt.Sprintf("ansible-tmpdir-%s", userStageName))

	if err := mkdirP(filepath.Join(path, "local")); err != nil {
		return "", err
	}

	if err := mkdirP(filepath.Join(path, "remote")); err != nil {
		return "", err
	}

	return path, nil
}

func (b *Ansible) containerWorkDir() string {
	return filepath.Join(b.extra.ContainerDappPath, "ansible-workdir")
}

func (b *Ansible) containerTmpDir() string {
	return filepath.Join(b.extra.ContainerDappPath, "ansible-tmpdir")
}

func mkdirP(path string) error {
	return os.MkdirAll(path, os.ModeDir)
}

func writeFile(path string, content string) error {
	return ioutil.WriteFile(path, []byte(content), os.ModePerm)
}
