module Dapp
  module Dimg
    class Builder::Ansible::Assets
      class << self
        def ansible_cfg(inventory, callback_plugins, become_exe, local_tmp, remote_tmp)
%{
[defaults]
inventory = #{inventory}
transport = local
; do not generate retry files in ro volumes
retry_files_enabled = False
; more verbose stdout like ad-hoc ansible command from flant/ansible fork
callback_plugins = #{callback_plugins}
stdout_callback = dapp
; force color
force_color = 1
module_compression = 'ZIP_STORED'
local_tmp = #{local_tmp}
remote_tmp = #{remote_tmp}
; keep ansiballz for debug
;keep_remote_files = 1
[privilege_escalation]
become = yes
become_method = sudo
become_exe = #{become_exe}
become_flags = -E
}
        end

        def hosts(python_path)
%{
localhost ansible_raw_live_stdout=yes ansible_script_live_stdout=yes ansible_python_interpreter=#{python_path}
}
        end

        # Python script! Do not enable string interpolation!
        def dapp_py
%q{
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

#from ansible.plugins.callback.live import CallbackModule as CallbackModule_live
# live.py moved to dapp
from callback.live import CallbackModule as CallbackModule_live
from ansible import constants as C

import os
import json

class CallbackModule(CallbackModule_live):

    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'stdout'
    CALLBACK_NAME = 'dapp'

    def __init__(self):
        self.super_ref = super(CallbackModule, self)
        self.super_ref.__init__()

    def v2_runner_on_failed(self, result, ignore_errors=False):
        self.super_ref.v2_runner_on_failed(result, ignore_errors)

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
            res = json.load(fh) #.read()
            fh.close()
        except:
            pass

        return res

    # dapp_stage_name commented for consistency with dappfile-yml behaviour
    def _display_dapp_config(self, task):
        tags = task.tags
        dump_config_section_key = ''
        #dapp_stage_name = ''
        if len(tags) > 0:
            # stage name appended before dump
            #dapp_stage_name = tags[-2]
            # last tag is dump of section
            dump_config_section_key = tags[-1]

        dump_config = self._read_dump_config_doc()
        dump_config_doc = dump_config['dump_config_doc']
        dump_config_section = dump_config['dump_config_sections'][dump_config_section_key]
        self._display.display("\n\n%s\n%s" % (dump_config_section, dump_config_doc), color=C.COLOR_DEBUG)

}
        end

        def crypt_py
%q{
def crypt(word, salt):
  return "FAKE_CRYPT"
}
        end

        def live_py
          %{
# (c) 2012-2014, Michael DeHaan <michael.dehaan@gmail.com>
# (c) 2017 Ansible Project
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

# Make coding more python3-ish
from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

DOCUMENTATION = '''
    callback: live
    type: stdout
    short_description: screen output for solo mode
    version_added: historical
    description:
        - Solo mode with live stdout for raw and script tasks with fallback to minimal
'''

from ansible.plugins.callback import CallbackBase
from ansible import constants as C
from ansible.vars.manager import strip_internal_keys

import json, re


class CallbackModule(CallbackBase):

    '''
    This is the default callback interface, which simply prints messages
    to stdout when new callback events are received.
    '''

    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'stdout'
    CALLBACK_NAME = 'live'

    # name for this tasks can be generated from free_form
    FREE_FORM_MODULES = ('raw', 'script', 'command', 'shell')
    #ERROR! this task 'debug' has extra params, which is only allowed in the following modules: command, win_command, shell, win_shell, script, include, include_vars, include_tasks, include_role, import_tasks, import_role, add_host, group_by, set_fact, raw, meta

    def __init__(self):
        super(CallbackModule, self).__init__()
        self._play = None

    def _task_header(self, task, msg):
        name = task.name
        if not name:
            if task.action in self.FREE_FORM_MODULES:
                name = task.args['_raw_params']
            if task.action == 'getent':
                db = task.args.get('database')
                key = task.args.get('key')
                name = '%s %s' % (db, key)
            if task.action == 'apt':
                name = task.args.get('name')
        name = re.sub(r'\s+', r' ', name)
        if len(name) > 25 :
            name = '%s...' % name[0:22]
        return u'%s [%s] %s' % (task.action, name, msg)

    def _display_command_generic_msg(self, task, result, caption, color):
        ''' output the result of a command run '''

        self._display.display("%s | rc=%s >>" % (self._task_header(task, caption), result.get('rc', -1)), color)
        # prevent dublication in case of live_stdout
        if not result.get('live_stdout', False):
            self._display.display("stdout was:", color=C.COLOR_HIGHLIGHT)
            self._display.display(result.get('stdout', ''))
        stderr = result.get('stderr', '')
        if stderr:
            self._display.display("stderr was:", color=C.COLOR_HIGHLIGHT)
            self._display.display(stderr, color=C.COLOR_ERROR)


    def _display_debug_msg(self, task, result):
        color = C.COLOR_OK
        if task.args.get('msg'):
            self._display.display("debug msg", color=C.COLOR_HIGHLIGHT)
            self._display.display(result.get('msg', ''), color)
        if task.args.get('var'):
            self._display.display("debug var \'%s\'" % task.args.get('var'), color=C.COLOR_HIGHLIGHT)
            var_obj = result.get(task.args.get('var'), '')
            if isinstance(var_obj, str):
                if 'IS NOT DEFINED' in var_obj:
                    color = C.COLOR_ERROR
                    path = task.get_path()
                    if path:
                        self._display.display(u"task path: %s" % path, color=C.COLOR_DEBUG)
                self._display.display(var_obj, color)
            else:
                self._display.display(json.dumps(var_obj, indent=4), color)

    # TODO remove stdout here if live_stdout!
    # TODO handle results for looped tasks
    def _dump_results(self, result, indent=None, sort_keys=True, keep_invocation=False):

        if not indent and (result.get('_ansible_verbose_always') or self._display.verbosity > 2):
            indent = 4

        # All result keys stating with _ansible_ are internal, so remove them from the result before we output anything.
        abridged_result = strip_internal_keys(result)

        # remove invocation unless specifically wanting it
        if not keep_invocation and self._display.verbosity < 3 and 'invocation' in result:
            del abridged_result['invocation']

        # remove diff information from screen output
        if self._display.verbosity < 3 and 'diff' in result:
            del abridged_result['diff']

        # remove exception from screen output
        if 'exception' in abridged_result:
            del abridged_result['exception']

        # remove msg, failed, changed
        if 'msg' in abridged_result:
            del abridged_result['msg']
        if 'failed' in abridged_result:
            del abridged_result['failed']
        if 'changed' in abridged_result:
            del abridged_result['changed']

        if len(abridged_result) > 0:
            return json.dumps(abridged_result, indent=indent, ensure_ascii=False, sort_keys=sort_keys)

        return ''

    def v2_playbook_on_play_start(self, play):
        self._play = play

    # command [copy artifacts] started
    # stdout
    # ...
    # command [copy artifacts] OK/FAILED/CHANGED
    # STDERR:  if failed
    # ...
    #
    def v2_playbook_on_task_start(self, task, is_conditional):
        self._display.v("TASK action=%s args=%s" % (task.action, json.dumps(task.args, indent=4)))

        if task.action == 'debug':
            return

        if self._play.strategy != 'free':
            self._display.display(self._task_header(task, "started"), color=C.COLOR_HIGHLIGHT)

    def v2_runner_on_ok(self, result):
        self._display.v("TASK action=%s OK => %s" % (result._task.action, json.dumps(result._result, indent=4)))

        self._clean_results(result._result, result._task.action)
        self._handle_warnings(result._result)

        task = result._task

        if task.action == 'debug':
            self._display_debug_msg(result._task, result._result)
        elif task.action in self.FREE_FORM_MODULES:
            self._display_command_generic_msg(result._task, result._result, "SUCCESS", C.COLOR_OK)
        else:
            if 'changed' in result._result and result._result['changed']:
                self._display.display("%s => %s" % (self._task_header(result._task, "SUCCESS"), self._dump_results(result._result, indent=4)), color=C.COLOR_CHANGED)
                #self._display.display(self._task_header(task, "OK")"%s | SUCCESS => %s" % (result._host.get_name(), self._dump_results(result._result, indent=4)), color=C.COLOR_CHANGED)
                #self._display.display("%s | SUCCESS => %s" % (result._host.get_name(), ), color=C.COLOR_CHANGED)
            else:
                self._display.display("%s => %s" % (self._task_header(result._task, "SUCCESS"), self._dump_results(result._result, indent=4)), color=C.COLOR_OK)
                #self._display.display("%s | SUCCESS => %s" % (result._host.get_name(), self._dump_results(result._result, indent=4)), color=C.COLOR_OK)

    def v2_runner_on_failed(self, result, ignore_errors=False):
        self._display.v("TASK action=%s FAILED => %s" % (result._task.action, json.dumps(result._result, indent=4)))

        self._handle_exception(result._result)
        self._handle_warnings(result._result)

        task = result._task

        if task.action in self.FREE_FORM_MODULES:
            self._display_command_generic_msg(result._task, result._result, "FAILED", C.COLOR_ERROR)
        #elif result._task.action in C.MODULE_NO_JSON and 'module_stderr' not in result._result:
        #    self._display.display(self._command_generic_msg(result._host.get_name(), result._result, "FAILED"), color=C.COLOR_ERROR)
        else:
            self._display.display(self._task_header(result._task, "FAILED"), color=C.COLOR_ERROR)
            if 'msg' in result._result:
                self._display.display(result._result['msg'], color=C.COLOR_ERROR)
            # clean system values from result and return a json
            dump_result = self._dump_results(result._result, indent=4)
            if dump_result:
                self._display.display("Task result => %s" % (self._dump_results(result._result, indent=4)), color=C.COLOR_ERROR)


    def v2_runner_on_skipped(self, result):
        self._display.display("%s | SKIPPED" % (result._host.get_name()), color=C.COLOR_SKIP)

    def v2_runner_on_unreachable(self, result):
        self._display.display("%s | UNREACHABLE! => %s" % (result._host.get_name(), self._dump_results(result._result, indent=4)), color=C.COLOR_UNREACHABLE)

    def v2_on_file_diff(self, result):
        if 'diff' in result._result and result._result['diff']:
            self._display.display(self._get_diff(result._result['diff']))

          }
        end

      end # << self
    end # Builder::Ansible::Assets
  end # Dimg
end # Dapp
