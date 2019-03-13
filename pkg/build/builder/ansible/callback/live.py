# -*- coding: utf-8 -*-
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

# Task output format:
#
# ┌ action 'name' [additional info] [START]
#   stdout:
#     Start installing...
#     operation 1 is done
#     operation 2 complete
#     all done.
#   stderr:
#     ...
#   msg: — if not failed or if failed without output
#   Exit code: 0/1/2..
# └ action 'name' [additional info] [OK]/[FAIL]
#
# Task with loop:
# ┌ action 'name' [additional info] [START]
#   stdout:
#     Start installing...
#     operation 1 is done
#     operation 2 complete
#     all done.
#   stderr:
#     ...
#   Message (if not failed or if failed without output)
#   Exit code: 0/1/2..
# ◆ action 'name' item 'item_var_content' [OK]
# ...
# ...
# └ action 'name' [additional info] [OK]/[FAIL]



from ansible.plugins.callback import CallbackBase
from ansible import constants as C
from ansible.vars.manager import strip_internal_keys
from ansible.utils.color import stringc

import io, sys
import json, re
from collections import Iterable
from datetime import datetime


# Taken from Dstat
class vt100:
    black = '\033[0;30m'
    darkred = '\033[0;31m'
    darkgreen = '\033[0;32m'
    darkyellow = '\033[0;33m'
    darkblue = '\033[0;34m'
    darkmagenta = '\033[0;35m'
    darkcyan = '\033[0;36m'
    gray = '\033[0;37m'

    darkgray = '\033[1;30m'
    red = '\033[1;31m'
    green = '\033[1;32m'
    yellow = '\033[1;33m'
    blue = '\033[1;34m'
    magenta = '\033[1;35m'
    cyan = '\033[1;36m'
    white = '\033[1;37m'

    blackbg = '\033[40m'
    redbg = '\033[41m'
    greenbg = '\033[42m'
    yellowbg = '\033[43m'
    bluebg = '\033[44m'
    magentabg = '\033[45m'
    cyanbg = '\033[46m'
    whitebg = '\033[47m'

    reset = '\033[0;0m'
    bold = '\033[1m'
    reverse = '\033[2m'
    underline = '\033[4m'

    clear = '\033[2J'
    #    clearline = '\033[K'
    clearline = '\033[2K'
    save = '\033[s'
    restore = '\033[u'
    save_all = '\0337'
    restore_all = '\0338'
    linewrap = '\033[7h'
    nolinewrap = '\033[7l'

    up = '\033[1A'
    down = '\033[1B'
    right = '\033[1C'
    left = '\033[1D'


class uSym:
    light_down_right = u'┌'
    light_up_right = u'└'
    light_vertical = u'│'
    light_vertical_and_right = u'├'
    black_diamond = u'◆'

class lColor:
    COLOR_OK = vt100.darkgreen
    COLOR_CHANGED = vt100.darkyellow
    COLOR_ERROR = vt100.darkred
    COLOR_DEBUG = vt100.darkgray


INDENT = vt100.reset+vt100.bold+uSym.light_vertical+vt100.reset+' '

class LiveCallbackHelpers(CallbackBase):
    def __init__(self):
        super(LiveCallbackHelpers, self).__init__()
        self._stdout = io.open(sys.stdout.fileno(), mode='w', encoding='utf8')

    def _print(self, *args):
        self._stdout.write(u''.join(self._flatten(args)))
        self._stdout.flush()

    # nested arrays into flat array    # action(module name)
    # action(module name) 'task name'

    def _flatten(self, l):
        """Yield items from any nested iterable"""
        if isinstance(l, (unicode, str, bytes)):
            yield l
            return
        for x in l:
            if not x:
                continue
            if isinstance(x, Iterable) and not isinstance(x, (unicode, str, bytes)):
                for sub_x in self._flatten(x):
                    yield sub_x
            else:
                yield x

    # string methods
    def _squash_center(self, s, l, placeholder='...'):
        pl = len(placeholder)
        if len(s) > l:
            # edge length of s to display
            sp = int((l - pl)/2)
            return u'%s%s%s' % (s[0:sp], placeholder, s[len(s)-sp-1+(l%2):])
        else:
            return s

    def _squash_right(self, s, l, placeholder='...'):
        pl = len(placeholder)
        if len(s) > l:
            return u'%s%s' % (s[0:l-pl], placeholder)
        else:
            return s

    def _clean_str(self, s):
        s = re.sub(r'\s+', r' ', s, flags=re.UNICODE)
        return s.strip()

    def _indent(self, indent, s):
        parts = re.split(r'(\n)', s)
        return ''.join(p if p == "\n" else '%s%s' % (indent, p) for p in parts)


class CallbackModule(LiveCallbackHelpers):
    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'stdout'
    CALLBACK_NAME = 'live'

    HEADER_PLACEHOLDER = '...'
    HEADER_NAME_INFO_LEN = 55
    HEADER_INFO_MIN_LEN = 5 + 3 + 5  # 5 letters from the edges and a placeholder length

    # name for this tasks can be generated from free_form (_raw_params argument)
    FREE_FORM_MODULES = ('raw', 'script', 'command', 'shell', 'meta')

    # Modules that are optimized by squashing loop items into a single call to
    # the module, mostly packaging modules with name argument
    # (apt, apk, dnf, homebrew, openbsd_pkg, pacman, pkgng, yum, zypper)
    SQUASH_LOOP_MODULES = frozenset(C.DEFAULT_SQUASH_ACTIONS)

    def __init__(self):
        super(CallbackModule, self).__init__()
        self._play = None
        # timestamps for duration
        self._task_started=None
        self._play_started=None
        self._item_done=None

    # header format is:
    # action 'task name' [significant args info]
    # if task name length exceed its maximum then format is:
    # action 'task name'
    # if no task name:
    # action [significant args info]
    # task name and significant args info are squashed to fit into available space
    def _task_details(self, task, start=False):
        task_name = self._clean_str(task.name)
        info = self._get_task_info_from_args(task, start) or ''

        if info != '':
            info_space = self.HEADER_NAME_INFO_LEN - len(task_name)
            if info_space >= self.HEADER_INFO_MIN_LEN or info_space >= len(info):
                info = ' [%s]' % self._squash_center(info, info_space-3)
            else:
                info = ''

        if task_name != '':
            if len(task_name)+len(info) > self.HEADER_NAME_INFO_LEN:
                task_name = self._squash_right(task_name, self.HEADER_NAME_INFO_LEN-len(info))
            task_name = " '%s'" % task_name

        return u'%s%s%s' % (task.action, task_name, info)

    # item details format is:
    # action 'task name' item 'item_name'
    # if no task_name:
    # action item 'item_name'
    # task_name and item_name are squashed if cannot fit into available space
    def _item_details(self, task, item_name):
        task_name = self._clean_str(task.name)
        item_name = self._clean_str(item_name)

        if task_name != '':
            task_space = self.HEADER_NAME_INFO_LEN - len(item_name)
            if task_space >= self.HEADER_INFO_MIN_LEN or task_space >= len(task_name):
                task_name = self._squash_right(task_name, task_space - 3)
                task_name = " '%s'" % task_name
            else:
                task_name = ''

        if item_name != '':
            if len(task_name)+len(item_name) > self.HEADER_NAME_INFO_LEN:
                item_name = self._squash_right(item_name, self.HEADER_NAME_INFO_LEN-len(task_name))
            item_name = " item '%s'" % (item_name)

        return u'%s%s%s' % (task.action, task_name, item_name)

    # Return content from significant arguments for well known modules
    # Also support items for the loops.
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

        return self._clean_str(info)

    # display task result content with indentation
    # Normally each item is displayed separately. But there are squashed
    # modules, where items are squashed into list and the result is in the first 'results' item.
    def _display_msg(self, task, result, color):
        if task.action in self.SQUASH_LOOP_MODULES and 'results' in result:
            if len(result['results']) > 0:
                return self._display_msg(task, result['results'][0], color)

        # prevent dublication of stdout in case of live_stdout
        if not result.get('live_stdout', False):
            stdout = result.get('stdout', None)
            if stdout:
                self._print(INDENT, vt100.bold, "stdout:", vt100.reset, "\n")
                self._print(self._indent(INDENT+'  ', stdout), "\n")
            stderr = result.get('stderr', '')
            if stderr:
                self._print(INDENT, vt100.bold, "stderr:", vt100.reset, "\n")
                self._print(self._indent(INDENT+'  ', stringc(stderr, C.COLOR_ERROR)), "\n")

        if self._msg_is_needed(task, result):
            self._print(INDENT, stringc(result['msg'], color), "\n")

        if 'rc' in result:
            exitCode = result['rc']
            exitColor = C.COLOR_OK
            if exitCode != '0' and exitCode != 0:
                exitColor = C.COLOR_ERROR
            self._print(INDENT, stringc('exit code: %s' % exitCode, exitColor), "\n")

    def _msg_is_needed(self, task, result):
        if 'msg' not in result:
            return False
        # No need to display msg for loop task, because each item is displayed separately.
        # Msg is needed if there are no items.
        if 'results' in result:
            if len(result['results']) > 0:
                return False
        # TODO more variants...
        return True

    def _display_debug_msg(self, task, result):
        #if (self._display.verbosity > 0 or '_ansible_verbose_always' in result) and '_ansible_verbose_override' not in result:
        color = C.COLOR_OK
        if task.args.get('msg'):
            self._print(INDENT, vt100.bold, "debug msg", vt100.reset, "\n")
            self._print(self._indent(INDENT+'  ', stringc(result.get('msg', ''), color)), "\n")
        if task.args.get('var'):
            self._print(INDENT, vt100.bold,
                        "debug var \'%s\'" % task.args.get('var'),
                        vt100.reset, "\n")
            var_obj = result.get(task.args.get('var'), '')
            if isinstance(var_obj, str):
                if 'IS NOT DEFINED' in var_obj:
                    color = C.COLOR_ERROR
                    path = task.get_path()
                    if path:
                        self._print(INDENT, stringc(u"task path: %s" % path, color=C.COLOR_DEBUG), "\n")
                self._print(self._indent(INDENT+'  ', stringc(var_obj, color)), "\n")
            else:
                self._print(self._indent(INDENT+'  ', stringc(json.dumps(var_obj, indent=4), color)), "\n")

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
        #if 'msg' in abridged_result:
        #    del abridged_result['msg']
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

    def v2_playbook_on_task_start(self, task, is_conditional):
        self._display.v("TASK action=%s args=%s" % (task.action, json.dumps(task.args, indent=4)))
        self._task_started = datetime.now()

        #if task.action == 'debug' or self._play.strategy == 'free':
        #    return
        if self._play.strategy == 'free':
            return

        # task header line
        self._print(
            vt100.reset, vt100.bold,
            uSym.light_down_right,
            ' ',
            self._task_details(task, start=True),
            " ",
            "[START]",
            vt100.reset,
            "\n"
        )


    def v2_runner_on_ok(self, result):
        self._display.v("TASK action=%s OK => %s" % (result._task.action, json.dumps(result._result, indent=4)))
        self._clean_results(result._result, result._task.action)
        self._handle_warnings(result._result)

        task = result._task
        color = lColor.COLOR_OK
        ccolor = C.COLOR_OK
        if 'changed' in result._result and result._result['changed']:
            color = lColor.COLOR_CHANGED
            ccolor = C.COLOR_CHANGED

        # task result info if any
        if task.action == 'debug':
            self._display_debug_msg(result._task, result._result)
        else:
            self._display_msg(result._task, result._result, ccolor)

        # task footer line
        status = u'[OK] %sms' % self._duration()
        self._print(
            vt100.reset, vt100.bold,
            uSym.light_up_right,
            ' ',
            self._task_details(task, start=True), vt100.reset,
            ' ',
            color,
            status,
            vt100.reset,
            "\n"
        )

    def v2_runner_item_on_ok(self, result):
        self._display.v("TASK action=%s item OK => %s" % (result._task.action, json.dumps(result._result, indent=4)))
        self._clean_results(result._result, result._task.action)
        self._handle_warnings(result._result)

        task = result._task
        if task.action in self.SQUASH_LOOP_MODULES:
            return
        color = lColor.COLOR_OK
        ccolor = C.COLOR_OK
        if 'changed' in result._result and result._result['changed']:
            color = lColor.COLOR_CHANGED
            ccolor = C.COLOR_CHANGED

        # item result info if any
        if task.action == 'debug':
            self._display_debug_msg(result._task, result._result)
        else:
            self._display_msg(result._task, result._result, ccolor)

        # task item footer line
        # TODO item duration
        status = u'[OK]'
        self._print(
            vt100.reset, vt100.bold,
            uSym.light_vertical_and_right,
            ' ',
            self._item_details(task, result._result.get('item', '')), vt100.reset,
            ' ',
            color,
            status,
            vt100.reset,
            "\n"
        )


    def v2_runner_on_failed(self, result, ignore_errors=False):
        self._display.v("TASK action=%s FAILED => %s" % (result._task.action, json.dumps(result._result, indent=4)))
        self._handle_exception(result._result)
        self._handle_warnings(result._result)

        task = result._task

        # task result info if any
        self._display_msg(task, result._result, C.COLOR_ERROR)

        # task footer line
        status = u'[FAIL] %sms' % self._duration()
        self._print(
            vt100.reset, vt100.bold,
            uSym.light_up_right,
            ' ',
            self._task_details(task, start=True),
            vt100.reset,
            ' ',
            lColor.COLOR_ERROR,
            status,
            vt100.reset,
            "\n"
        )

    def v2_runner_item_on_failed(self, result, ignore_errors=False):
        self._display.v("TASK action=%s ITEM FAILED => %s" % (result._task.action, json.dumps(result._result, indent=4)))
        self._handle_exception(result._result)
        self._handle_warnings(result._result)
        task = result._task
        if task.action in self.SQUASH_LOOP_MODULES:
            return
        # task item result info if any
        self._display_msg(task, result._result, C.COLOR_ERROR)
        # task item footer line
        # TODO item duration
        status = u'[FAIL]'
        self._print(
            vt100.reset, vt100.bold,
            uSym.light_vertical_and_right,
            ' ',
            self._item_details(task, result._result.get('item', '')), vt100.reset,
            ' ',
            C.COLOR_ERROR,
            status,
            vt100.reset,
            "\n"
        )

    def v2_runner_on_skipped(self, result):
        self._print(stringc("%s | SKIPPED" % (result._host.get_name()), C.COLOR_SKIP), "\n")

    def v2_runner_on_unreachable(self, result):
        self._print(stringc("%s | UNREACHABLE! => %s" % (result._host.get_name(), self._dump_results(result._result, indent=4)), color=C.COLOR_UNREACHABLE), "\n")

    def v2_on_file_diff(self, result):
        if 'diff' in result._result and result._result['diff']:
            self._print(self._get_diff(result._result['diff']), "\n")

    def _handle_exception(self, result):
        if 'exception' in result:
            msg = "An exception occurred during task execution. The full traceback is:\n" + result['exception']
            del result['exception']
            self._print(stringc(msg, C.COLOR_ERROR))
