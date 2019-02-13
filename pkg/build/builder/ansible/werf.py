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
        print('qweqweqweqwe')
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
        self._display.display("\n\n%s\n%s\n" % (dump_config_section, dump_config_doc), color=C.COLOR_DEBUG)
