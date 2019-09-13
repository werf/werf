package builder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/flant/werf/pkg/stapel"
	"github.com/flant/werf/pkg/util"
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

	// generate inventory with localhost and python in stapel
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

	// Ansible-playbook starter: setup python path without PYTHONPATH environment var
	ioutil.WriteFile(
		filepath.Join(stageWorkDir, "ansible-playbook"),
		[]byte(fmt.Sprintf(
			`#!%s

import sys
sys.path.append("%s")

execfile("%s")
`, stapel.PythonBinPath(), filepath.Join(b.containerWorkDir(), "lib"), stapel.AnsiblePlaybookBinPath())),
		os.FileMode(0777),
	)

	stageWorkDirLib := filepath.Join(stageWorkDir, "lib")
	if err := mkdirP(stageWorkDirLib); err != nil {
		return err
	}

	// crypt.py hack
	// TODO should be in stapel image
	writeFile(filepath.Join(stageWorkDirLib, "crypt.py"), b.assetsCryptPy())

	// sitecustomize with mocks
	writeFile(filepath.Join(stageWorkDirLib, "sitecustomize.py"), b.assetsSiteCustomizePy())

	callbackPackageDir := filepath.Join(stageWorkDirLib, "callback")
	if err := mkdirP(callbackPackageDir); err != nil {
		return err
	}

	writeFile(filepath.Join(callbackPackageDir, "__init__.py"), b.assetsCallbackInitPy())

	if livePyPath, exist := os.LookupEnv("WERF_DEBUG_ANSIBLE_LIVE_PY_PATH"); exist {
		// hardlink a local live.py into workdir to ease ansible callback development
		er := os.Link(livePyPath, filepath.Join(callbackPackageDir, "live.py"))
		if er != nil {
			return er
		}
	} else {
		writeFile(filepath.Join(callbackPackageDir, "live.py"), b.assetsCallbackLivePy())
	}

	// add werf specific stdout callback for ansible
	if werfPyPath, exist := os.LookupEnv("WERF_DEBUG_ANSIBLE_WERF_PY_PATH"); exist {
		// hardlink to a local live.py into workdir to ease ansible callback development
		er := os.Link(werfPyPath, filepath.Join(callbackPackageDir, "werf.py"))
		if er != nil {
			return er
		}
	} else {
		writeFile(filepath.Join(callbackPackageDir, "werf.py"), b.assetsCallbackWerfPy())
	}

	werfPackageDir := filepath.Join(stageWorkDirLib, "werf")
	if err := mkdirP(werfPackageDir); err != nil {
		return err
	}
	writeFile(filepath.Join(werfPackageDir, "__init__.py"), b.assetsWerfInitPy())
	writeFile(filepath.Join(werfPackageDir, "live_stdout.py"), b.assetsWerfLiveStdoutPy())
	writeFile(filepath.Join(werfPackageDir, "tee_popen.py"), b.assetsWerfTeePopenPy())

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
			if val, ok := task["tags"].(string); ok {
				tags = append(tags, val)
			} else if tags, err = util.InterfaceToStringArray(task["tags"]); err != nil {
				return nil, err
			}
		}
		dumpTag := fmt.Sprintf("task_%d", ind)
		tags = append(tags, dumpTag)
		task["tags"] = tags

		dumpConfigSections[dumpTag] = ansibleTask.GetDumpConfigSection()

		tasks = append(tasks, task)
	}

	dumpConfig := map[string]interface{}{
		"dump_config_doc":      b.config.GetDumpConfigSection(),
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
	return filepath.Join(b.extra.ContainerWerfPath, "ansible-workdir")
}

func (b *Ansible) containerTmpDir() string {
	return filepath.Join(b.extra.ContainerWerfPath, "ansible-tmpdir")
}

func mkdirP(path string) error {
	return os.MkdirAll(path, os.FileMode(0775))
}

func writeFile(path string, content string) error {
	return ioutil.WriteFile(path, []byte(content), os.FileMode(0664))
}

func writeFileBytes(path string, content []byte) error {
	return ioutil.WriteFile(path, content, os.FileMode(0664))
}
