package builder

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/dappdeps"
)

func (b *Ansible) assetsAnsibleCfg() string {
	hostsPath := filepath.Join(b.containerWorkDir(), "hosts")
	callbackPluginsPath := filepath.Join(b.containerWorkDir(), "lib", "callback")
	sudoBinPath := dappdeps.BaseBinPath("sudo")
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
	return fmt.Sprintf(format, dappdeps.AnsibleBinPath("python"))
}

func (b *Ansible) assetsCryptPy() string {
	return `def crypt(word, salt):
  return "FAKE_CRYPT"`
}

func (b *Ansible) assetsLivePy() string {
	return `
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
from collections import Iterable
from datetime import datetime


class CallbackModule(CallbackBase):

    '''
    This is the default callback interface, which simply prints messages
    to stdout when new callback events are received.
    '''

    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'stdout'
    CALLBACK_NAME = 'live'

    HEADER_PLACEHOLDER = '...'
    HEADER_NAME_INFO_LEN = 55
    HEADER_INFO_MIN_LEN = 5 + 3 + 5 # 5 letters from the edges and a placeholder length


    # name for this tasks can be generated from free_form
    FREE_FORM_MODULES = ('raw', 'script', 'command', 'shell')
    # this modules are tested and we can get additional info from them
    KNOWN_RESULT_MODULES = ('user', 'group', 'apt', 'copy', 'get_url', 'file')

    #ERROR! this task 'debug' has extra params, which is only allowed in the following modules: command, win_command, shell, win_shell, script, include, include_vars, include_tasks, include_role, import_tasks, import_role, add_host, group_by, set_fact, raw, meta

    def __init__(self):
        super(CallbackModule, self).__init__()
        self._play = None
        # time
        self._task_started=None
        self._play_started=None
        self._item_done=None

    # action(module name)
    # action(module name) 'task name'
    # action(module name) 'task name' [significant arg]

    #    if HEADER_NAME_INFO_LEN - len(taskName) >= space + bracket + 5 left letters + ... + 5 right letters + bracket
    # or if HEADER_NAME_INFO_LEN - len(taskName) >= HEADER_INFO_MIN_LEN + 3 - len(info)
    # then add info or squashed from center info
    def _task_header(self, task, msg, start=False):
        taskName = re.sub(r'\s+', r' ', task.name)
        info = self._get_task_info_from_args(task, start) or ''

        if info != '':
            infoSpace = self.HEADER_NAME_INFO_LEN - len(taskName)
            self._display.v("infoSpace=%d" % infoSpace)
            if infoSpace >= self.HEADER_INFO_MIN_LEN or infoSpace >= len(info):
                info = ' [%s]' % self._squash_center(info, infoSpace-3)
            else:
                info = ''


        if taskName != '':
            if len(taskName)+len(info) > self.HEADER_NAME_INFO_LEN:
                taskName = self._squash_right(taskName, self.HEADER_NAME_INFO_LEN-len(info))
            taskName = " '%s'" % taskName

        return u'%s%s%s %s' % (task.action, taskName, info, msg)

    def _item_header(self, task, item_name, msg):
        msg = u'| Item %s | %s' % (item_name, msg)
        return self._task_header(task, msg)


    def _get_task_info_from_args(self, task, start=False):
        info = ''
        if task.action in self.FREE_FORM_MODULES:
            info = task.args.get('_raw_params', '')
        if task.action == 'file':
            info = task.args.get('path','')
        if task.action == 'copy':
            info = task.args.get('dest','')
        if task.action == 'group':
            info = task.args.get('name','')
        if task.action == 'user':
            info = task.args.get('name','')
        if task.action == 'get_url':
            info = task.args.get('url','')
        if task.action == 'getent':
            db = task.args.get('database','')
            key = task.args.get('key','')
            info = '%s %s' % (db, key)
        if task.action == 'apk':
            info = task.args.get('name', '')
        if task.action == 'apt':
            info1 = task.args.get('name', None)
            info2 = task.args.get('package', None)
            info3 = task.args.get('pkg', None)
            info = ', '.join(list(self._flatten([info1, info2, info3])))
        if task.action == 'apt_repository':
            info = task.args.get('repo', '')
        if task.action == 'apt_key':
            info = task.args.get('id', '')
        if task.action == 'unarchive':
            info = task.args.get('src', '')
        if task.action == 'locale_gen':
            info = task.args.get('name', '')
        if task.action == 'lineinfile':
            info = task.args.get('path', '')
        if task.action == 'blockinfile':
            info = task.args.get('path', '')
        if task.action == 'composer':
            info = task.args.get('command', 'install')

	if task.loop and start:
            loop_args = task.loop_args
            if len(loop_args) > 0:
                 info = "'%s' over %s" % (info, ', '.join(loop_args))
        return info


    def _squash_center(self, s, l, placeholder='...'):
        pl = len(placeholder)
        self._display.v("  len(s)=%d l=%d pl=%d" % (len(s), l, pl))

        if len(s) > l:
            # edge length of s to display
            sp = int((l - pl)/2)
            self._display.v("  sp=%d" % sp)
            return u'%s%s%s' % (s[0:sp], placeholder, s[len(s)-sp-1+(l%2):])
        else:
            return s

    def _squash_right(self, s, l, placeholder='...'):
        pl = len(placeholder)
        if len(s) > l:
            return u'%s%s' % (s[0:l-pl], placeholder)
        else:
            return s

    def _flatten(self, l):
	"""Yield items from any nested iterable; see Reference."""
        if isinstance(l, (unicode, str, bytes)):
            yield l
            return
	for x in l:
            self._display.v('flatten %s' % x)
            if not x:
                continue
            if isinstance(x, Iterable) and not isinstance(x, (unicode, str, bytes)):
                for sub_x in self._flatten(x):
                    yield sub_x
            else:
                yield x

    def _display_msg(self, task, result, caption, color):
        # prevent dublication in case of live_stdout
        if not result.get('live_stdout', False):
            stdout = result.get('stdout', None)
            if stdout:
              self._display.display("stdout:", color=C.COLOR_HIGHLIGHT)
              self._display.display(stdout)
            stderr = result.get('stderr', '')
            if stderr:
                self._display.display("stderr:", color=C.COLOR_HIGHLIGHT)
                self._display.display(stderr, color=C.COLOR_ERROR)

#        if task.loop and 'results' in result:
#            results = result['results']
#            if len(results) > 0:
#                for res in results:
#                    self._display.display("Item: %s" % res.get('item',''), color=C.COLOR_HIGHLIGHT)
#		    if not res.get('live_stdout', False):
#			stdout = res.get('stdout', None)
#			if stdout:
#			  self._display.display("stdout:", color=C.COLOR_HIGHLIGHT)
#			  self._display.display(stdout)
#			stderr = res.get('stderr', '')
#			if stderr:
#			    self._display.display("stderr:", color=C.COLOR_HIGHLIGHT)
#			    self._display.display(stderr, color=C.COLOR_ERROR)
#		    if 'msg' in res:
#			self._display.display(res['msg'], color)
#
#		    if 'rc' in res:
#			exitCode = res['rc']
#			exitColor = C.COLOR_OK
#			if exitCode != '0':
#			    exitColor = C.COLOR_ERROR
#
#			self._display.display('Exit code: %s' % exitCode, exitColor)

        if 'msg' in result:
            self._display.display(result['msg'], color)

        if 'rc' in result:
            exitCode = result['rc']
            exitColor = C.COLOR_OK
            if exitCode != '0':
                exitColor = C.COLOR_ERROR

            self._display.display('Exit code: %s' % exitCode, exitColor)

        if 'item' in result:
            self._display.display(self._item_header(task, result['item'], caption), color)
        else:
            self._display.display(self._task_header(task, caption), color)

    def _display_command_generic_msg(self, task, result, caption, color):
        ''' output the result of a command run '''

        self._display.display("%s | exit code %s >>" % (self._task_header(task, caption), result.get('rc', -1)), color)
        msg = result.get('msg')
        if msg:
            self._display.display(msg, color)
        # prevent dublication in case of live_stdout
        if not result.get('live_stdout', False):
            stdout = result.get('stdout', None)
            if stdout:
              self._display.display("stdout was:", color=C.COLOR_HIGHLIGHT)
              self._display.display(stdout)
        stderr = result.get('stderr', '')
        if stderr:
            self._display.display("stderr was:", color=C.COLOR_HIGHLIGHT)
            self._display.display(stderr, color=C.COLOR_ERROR)


    def _display_debug_msg(self, task, result):
        #if (self._display.verbosity > 0 or '_ansible_verbose_always' in result) and '_ansible_verbose_override' not in result:
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
 #       if 'msg' in abridged_result:
 #           del abridged_result['msg']
        if 'failed' in abridged_result:
            del abridged_result['failed']
        if 'changed' in abridged_result:
            del abridged_result['changed']

        if len(abridged_result) > 0:
            return json.dumps(abridged_result, indent=indent, ensure_ascii=False, sort_keys=sort_keys)

        return ''

    def _duration(self):
        end = datetime.now()
        total_duration = (end - self._task_started)
        duration = total_duration.total_seconds() * 1000
        return duration

    def v2_playbook_on_play_start(self, play):
        self._play = play
        self._play_started = datetime.now()

    # command [copy artifacts] started
    # stdout
    # ...
    # command [copy artifacts] OK/FAILED/CHANGED
    # STDERR:  if failed
    # ...
    #
    def v2_playbook_on_task_start(self, task, is_conditional):
        self._display.v("TASK action=%s args=%s" % (task.action, json.dumps(task.args, indent=4)))
        self._task_started = datetime.now()

        if task.action == 'debug':
            return

        if self._play.strategy != 'free':
            self._display.display(self._task_header(task, "START", start=True), color=C.COLOR_HIGHLIGHT)

        #self._display.v(json.dumps(task.args, indent=4), C.COLOR_HIGHLIGHT)

    def v2_runner_on_ok(self, result):
        self._display.v("TASK action=%s OK => %s" % (result._task.action, json.dumps(result._result, indent=4)))

        self._clean_results(result._result, result._task.action)
        self._handle_warnings(result._result)

        task = result._task

        #self._display.v(json.dumps(task.args, indent=4), C.COLOR_HIGHLIGHT)

        # special display for debug action
        if task.action == 'debug':
            self._display_debug_msg(result._task, result._result)
        # display stdout and stderr for command modules
#        elif task.action in self.FREE_FORM_MODULES:
#            self._display_command_generic_msg(result._task, result._result, "SUCCESS", C.COLOR_OK)
        # display other modules
        else:
            status = u'OK %sms' % self._duration()
            if 'changed' in result._result and result._result['changed']:
                self._display_msg(result._task, result._result, status, C.COLOR_CHANGED)
            else:
                self._display_msg(result._task, result._result, status, C.COLOR_OK)

    def v2_runner_item_on_ok(self, result):
        self._display.v("TASK action=%s item OK => %s" % (result._task.action, json.dumps(result._result, indent=4)))

        self._clean_results(result._result, result._task.action)
        self._handle_warnings(result._result)

        task = result._task

        #self._display.v(json.dumps(task.args, indent=4), C.COLOR_HIGHLIGHT)

        # special display for debug action
        if task.action == 'debug':
            self._display_debug_msg(result._task, result._result)
        # display stdout and stderr for command modules
#        elif task.action in self.FREE_FORM_MODULES:
#            self._display_command_generic_msg(result._task, result._result, "SUCCESS", C.COLOR_OK)
        # display other modules
        else:
            if 'changed' in result._result and result._result['changed']:
                self._display_msg(result._task, result._result, "OK", C.COLOR_CHANGED)
            else:
                self._display_msg(result._task, result._result, "OK", C.COLOR_OK)

    def v2_runner_on_failed(self, result, ignore_errors=False):
        self._display.v("TASK action=%s FAILED => %s" % (result._task.action, json.dumps(result._result, indent=4)))

        self._handle_exception(result._result)
        self._handle_warnings(result._result)

        #task = result._task
        #self._display.v(json.dumps(task.args, indent=4), C.COLOR_HIGHLIGHT)

        status = u'FAIL %sms' % self._duration()

        self._display_msg(result._task, result._result, status, C.COLOR_ERROR)

#        if task.action in self.FREE_FORM_MODULES:
#            self._display_command_generic_msg(result._task, result._result, "FAILED", C.COLOR_ERROR)
#        #elif result._task.action in C.MODULE_NO_JSON and 'module_stderr' not in result._result:
#        #    self._display.display(self._command_generic_msg(result._host.get_name(), result._result, "FAILED"), color=C.COLOR_ERROR)
#        else:
#            if 'msg' in result._result:
#                self._display.display(result._result['msg'], color=C.COLOR_ERROR)
#
#            self._display.display(self._task_header(result._task, "FAILED"), color=C.COLOR_ERROR)
            # clean system values from result and return a json
            #dump_result = self._dump_results(result._result, indent=4)
            #if dump_result:
            #    self._display.display("Task result => %s" % (self._dump_results(result._result, indent=4)), color=C.COLOR_ERROR)

    def v2_runner_item_on_failed(self, result, ignore_errors=False):
        self._display.v("TASK action=%s ITEM FAILED => %s" % (result._task.action, json.dumps(result._result, indent=4)))

        self._handle_exception(result._result)
        self._handle_warnings(result._result)

        #task = result._task
        #self._display.v(json.dumps(task.args, indent=4), C.COLOR_HIGHLIGHT)

        self._display_msg(result._task, result._result, "FAIL", C.COLOR_ERROR)


    def v2_runner_on_skipped(self, result):
        self._display.display("%s | SKIPPED" % (result._host.get_name()), color=C.COLOR_SKIP)

    def v2_runner_on_unreachable(self, result):
        self._display.display("%s | UNREACHABLE! => %s" % (result._host.get_name(), self._dump_results(result._result, indent=4)), color=C.COLOR_UNREACHABLE)

    def v2_on_file_diff(self, result):
        if 'diff' in result._result and result._result['diff']:
            self._display.display(self._get_diff(result._result['diff']))
`
}

func (b *Ansible) assetsWerfPy() string {
	return `
# (c) 2018, Ivan Mikheykin <ivan.mikheykin@flant.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

# Make coding more python3-ish
from __future__ import (absolute_import, division, print_function)
__metaclass__ = type


DOCUMENTATION = '''
    callback: werf
    type: stdout
    short_description: live output for raw and script with werf specific additions
    version_added: "2.4"
    description:
        - Solo mode with live stdout for raw and script tasks
        - Werf specific error messages
    requirements:
      - set as stdout callback in configuration
'''

#from ansible.plugins.callback.live import CallbackModule as CallbackModule_live
# live.py moved to werf
from callback.live import CallbackModule as CallbackModule_live
from ansible import constants as C

import os
import json

class CallbackModule(CallbackModule_live):

    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'stdout'
    CALLBACK_NAME = 'werf'

    def __init__(self):
        self.super_ref = super(CallbackModule, self)
        self.super_ref.__init__()

    def v2_runner_on_failed(self, result, ignore_errors=False):
        self.super_ref.v2_runner_on_failed(result, ignore_errors)

        # get config sections from werf
        # task config text is in a last tag
        # doctext is in a file WERF_DUMP_CONFIG_DOC_PATH
        self._display_werf_config(result._task)

    def _read_dump_config_doc(self):
        # read content from file in WERF_DUMP_CONFIG_DOC_PATH env
        if 'WERF_DUMP_CONFIG_DOC_PATH' not in os.environ:
            return ''
        dump_path = os.environ['WERF_DUMP_CONFIG_DOC_PATH']
        res = ''
        try:
            fh = open(dump_path, 'r')
            res = json.load(fh) #.read()
            fh.close()
        except:
            pass

        return res

    # werf_stage_name commented for consistency with werffile-yml behaviour
    def _display_werf_config(self, task):
        tags = task.tags
        dump_config_section_key = ''
        #werf_stage_name = ''
        if len(tags) > 0:
            # stage name appended before dump
            #werf_stage_name = tags[-2]
            # last tag is dump of section
            dump_config_section_key = tags[-1]

        dump_config = self._read_dump_config_doc()
        dump_config_doc = dump_config['dump_config_doc']
        dump_config_section = dump_config['dump_config_sections'][dump_config_section_key]
        self._display.display("\n\n%s\n%s" % (dump_config_section, dump_config_doc), color=C.COLOR_DEBUG)
`
}
