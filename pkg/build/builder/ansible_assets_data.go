package builder

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/build/builder/ansible"
	"github.com/flant/werf/pkg/stapel"
)

func (b *Ansible) assetsAnsibleCfg() string {
	hostsPath := filepath.Join(b.containerWorkDir(), "hosts")
	callbackPluginsPath := filepath.Join(b.containerWorkDir(), "lib", "callback")
	sudoBinPath := stapel.SudoBinPath()
	localTmpDirPath := filepath.Join(b.containerTmpDir(), "local")
	remoteTmpDirPath := filepath.Join(b.containerTmpDir(), "remote")

	format := `[defaults]
inventory = %[1]s
transport = local
; do not generate retry files in ro volumes
retry_files_enabled = False
; more verbose stdout like ad-hoc ansible command from flant/ansible fork
callback_plugins = %[2]s
stdout_callback = werf
; force color
force_color = 1
; use uncompressed modules because of local connection
module_compression = 'ZIP_STORED'
local_tmp = %[3]s
remote_tmp = %[4]s
; keep ansiballz for debug
;keep_remote_files = 1
[privilege_escalation]
become = yes
become_method = sudo
become_exe = %[5]s
become_flags = -E -H`

	return fmt.Sprintf(format, hostsPath, callbackPluginsPath, localTmpDirPath, remoteTmpDirPath, sudoBinPath)
}

func (b *Ansible) assetsHosts() string {
	format := "localhost ansible_raw_live_stdout=yes ansible_script_live_stdout=yes ansible_python_interpreter=%s"
	return fmt.Sprintf(format, stapel.PythonBinPath())
}

func (b *Ansible) assetsCryptPy() string {
	return ansible.FSMustString(false, "/ansible/crypt.py")
}

func (b *Ansible) assetsSiteCustomizePy() string {
	return ansible.FSMustString(false, "/ansible/sitecustomize.py")
}

func (b *Ansible) assetsCallbackInitPy() string {
	return ansible.FSMustString(false, "/ansible/callback/__init__.py")
}

func (b *Ansible) assetsCallbackLivePy() string {
	return ansible.FSMustString(false, "/ansible/callback/live.py")
}

func (b *Ansible) assetsCallbackWerfPy() string {
	return ansible.FSMustString(false, "/ansible/callback/werf.py")
}

func (b *Ansible) assetsWerfInitPy() string {
	return ansible.FSMustString(false, "/ansible/werf/__init__.py")
}

func (b *Ansible) assetsWerfLiveStdoutPy() string {
	return ansible.FSMustString(false, "/ansible/werf/live_stdout.py")
}

func (b *Ansible) assetsWerfTeePopenPy() string {
	return ansible.FSMustString(false, "/ansible/werf/tee_popen.py")
}
