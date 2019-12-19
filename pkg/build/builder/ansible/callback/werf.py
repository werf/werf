# -*- coding: utf-8 -*-
# (c) 2018, Ivan Mikheykin <ivan.mikheykin@flant.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

# Make coding more python3-ish
from __future__ import (absolute_import, division, print_function)
__metaclass__ = type


DOCUMENTATION = '''
    callback: werf
    type: stdout
    short_description: Print related werf config section in case of task failure
    version_added: "2.4"
    description:
        - Solo mode with live stdout for raw and script tasks
        - werf specific error messages
    requirements:
      - set as stdout callback in configuration
'''

from callback.live import CallbackModule as CallbackModule_live
from callback.live import vt100, lColor
from ansible import constants as C
from ansible.utils.color import stringc

import os
import json

class CallbackModule(CallbackModule_live):

    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'stdout'
    CALLBACK_NAME = 'werf'

    def __init__(self):
        self.super_ref = super(CallbackModule, self)
        self.super_ref.__init__()

    def v2_runner_on_failed(self, result, **kwargs):
        self.super_ref.v2_runner_on_failed(result, **kwargs)

        # get config sections from werf
        # task config text is in a last tag
        # doctext is in a file WERF_DUMP_CONFIG_DOC_PATH
        self._display_werf_config(result._task)

    def _read_dump_config_doc(self):
        # read content from file in WERF_DUMP_CONFIG_DOC_PATH env
        if 'WERF_DUMP_CONFIG_DOC_PATH' not in os.environ:
            return {}
        dump_path = os.environ['WERF_DUMP_CONFIG_DOC_PATH']
        res = {}
        try:
            fh = open(dump_path, 'r')
            res = json.load(fh)
            fh.close()
        except:
            pass

        return res

    # werf_stage_name commented for consistency with werffile-yml behaviour
    def _display_werf_config(self, task):
        tags = task.tags
        if not tags or len(tags) == 0:
            return

        # last tag is a key to a section dump in dump_config
        dump_config_section_key = tags[-1]

        dump_config = self._read_dump_config_doc()
        dump_config_doc = dump_config.get('dump_config_doc', '')
        dump_config_sections = dump_config.get('dump_config_sections', {})
        dump_config_section = dump_config_sections.get(dump_config_section_key, '')
        self.LogArgs(
            u"\n",
            lColor.COLOR_DEBUG, u"Failed task configuration:\n\n", vt100.reset,
            stringc(dump_config_section, C.COLOR_DEBUG),
            u"\n",
            stringc(dump_config_doc, C.COLOR_DEBUG),
            u"\n")
