from werf.tee_popen import TeePopen
import sys, re
import os


# __del__ is called when python assigns real arguments into sys.argv
class IsAnsiballZ(object):
    ansiballz_re = re.compile(r'/AnsiballZ.*\.py$')

    def __del__(self):
        if len(sys.argv) > 0:
            if self.ansiballz_re.search(sys.argv[0]):
                mock_subprocess_Popen_for_AnsiballZ()
            else:
                mock_subprocess_Popen_in_low_level_execute()


sys.argv = IsAnsiballZ()


# Mock subprocess.Popen with werf.TeePopen.
# Use TeePopen only if payload.zip is in sys.path.
def mock_subprocess_Popen_for_AnsiballZ():
    ansiballz_path_re = re.compile(r'/ansible_(command|apt|apk|yum)_payload\.zip$')
    import subprocess
    original_popen = subprocess.Popen
    def new_Popen(args, **kwargs):
        import sys
        if ansiballz_path_re.search(sys.path[0]):
            return TeePopen(args, original_popen=original_popen, **kwargs)
        else:
            return original_popen(args, **kwargs)
    subprocess.Popen = new_Popen


# Mock suprocess.Popen for _low_level_execute_command
# if raw or script action are in use.
def mock_subprocess_Popen_in_low_level_execute():
    from ansible.plugins.action import ActionBase
    import subprocess
    original_low_exec = ActionBase._low_level_execute_command
    def new_low_exec(self, cmd, **kwargs):
        le_original_popen = subprocess.Popen

        self_name = str(type(self))
        use_tee = False
        if 'ansible.plugins.action.raw.ActionModule' in self_name:
            use_tee = True

        if 'ansible.plugins.action.script.ActionModule' in self_name:
            use_tee = True

        if use_tee:
            def new_popen(args, **kwargs):
                return TeePopen(args, original_popen=le_original_popen, **kwargs)
            subprocess.Popen = new_popen

        res = original_low_exec(self, cmd, **kwargs)

        subprocess.Popen = le_original_popen
        return res
    ActionBase._low_level_execute_command = new_low_exec
