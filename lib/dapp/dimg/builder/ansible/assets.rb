module Dapp
  module Dimg
    class Builder::Ansible::Assets
      class << self
        def ansible_cfg(builder)
%{
[defaults]
inventory = #{builder.container_playbook_path}/hosts
transport = local
; do not generate retry files in ro volumes
retry_files_enabled = False
; more verbose stdout like ad-hoc ansible command from flant/ansible fork
callback_plugins = #{builder.container_playbook_path}
stdout_callback = dapp
; force color
force_color = 1
module_compression = 'ZIP_STORED'

[privilege_escalation]
become = yes
become_method = sudo
become_exe = #{builder.dimg.dapp.sudo_bin}
become_flags = -E
}
        end

        def hosts(builder)
%{
localhost ansible_raw_live_stdout=yes ansible_script_live_stdout=yes ansible_python_interpreter=#{builder.python_path}
}
        end

        def dapp_py(builder)
%{
# (c) 2018, Ivan Mikheykin <ivan.mikheykin@flant.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

# Make coding more python3-ish
from __future__ import (absolute_import, division, print_function)
__metaclass__ = type


DOCUMENTATION = '''
    callback: dapp
    type: stdout
    short_description: live output for raw and script with dapp specific additions
    version_added: "2.4"
    description:
        - Solo mode with live stdout for raw and script tasks
        - Dapp specific error messages
    requirements:
      - set as stdout callback in configuration
'''

from ansible.plugins.callback.live import CallbackModule as CallbackModule_live


class CallbackModule(CallbackModule_live):

    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'stdout'
    CALLBACK_NAME = 'dapp'

    def __init__(self):
        self.super_ref = super(CallbackModule, self)
        self.super_ref.__init__()

    def v2_runner_on_failed(self, result, ignore_errors=False):
        super().v2_runner_on_failed(self, result, ignore_errors)

        # get config sections from dapp
        # task config text is in a last tag
        # doctext is in a file DAPP_DUMP_CONFIG_DOC_PATH
        self._display_dapp_config(result._task)

    def _read_dump_config_doc(self):
        # read content from file in DAPP_DUMP_CONFIG_DOC_PATH env
        if 'DAPP_DUMP_CONFIG_DOC_PATH' not in os.environ:
            return ''
        dump_path = os.environ['DAPP_DUMP_CONFIG_DOC_PATH']
        res = ''
        try:
            fh = open(dump_path, 'r')
            res = fh.read()
            fh.close()
        except:
            pass

        return res

    # dapp_stage_name commented for consistency with dappfile-yml behaviour
    def _display_dapp_config(self, task):
        tags = task.tags
        dump_config_section = ''
        #dapp_stage_name = ''
        if len(tags) > 0:
            # stage name appended before dump
            #dapp_stage_name = tags[-2]
            # last tag is dump of section
            dump_config_section = tags[-1]
        dump_config_doc = self._read_dump_config_doc()
        self._display.display("\n\n%s\n%s" % (dump_config_section, dump_config_doc), color=C.COLOR_DEBUG)

}
        end
      end # << self
    end # Builder::Ansible::Assets
  end # Dimg
end # Dapp
