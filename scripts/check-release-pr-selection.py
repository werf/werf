import os
from pathlib import Path
import subprocess
import tempfile
import textwrap

workflow = Path('.github/workflows/tag_auto-create.yml').read_text()
step = workflow.split('      - name: Relabel merged release PR\n', 1)[1].split('\n      - name:', 1)[0]
assert 'RELEASE_BRANCH: ${{ github.ref_name }}' in step
script = textwrap.dedent(step.split('        run: |\n', 1)[1])

with tempfile.TemporaryDirectory() as directory:
    root = Path(directory)
    gh = root / 'gh'
    gh.write_text('#!/bin/sh\nprintf "%s\\n" "$*" >> "$GH_TEST_LOG"\n'
                  'if [ "$1 $2" = "pr list" ]; then printf "%s" "$GH_TEST_PR"; fi\n')
    gh.chmod(0o755)
    log = root / 'calls'
    for branch in ['main', '2']:
        for number in ['', '123']:
            log.write_text('')
            env = os.environ | {'PATH': f'{root}:{os.environ["PATH"]}',
                                'RELEASE_BRANCH': branch, 'GH_TEST_PR': number,
                                'GH_TEST_LOG': str(log)}
            subprocess.run(['bash', '-e', '-c', script], env=env, check=True)
            calls = log.read_text().splitlines()
            assert calls[0] == (f'pr list --state merged --base {branch} '
                                '--label autorelease: pending --limit 1 --json number '
                                '-q .[0].number // empty'), calls
            expected = ([f'pr edit {number} --remove-label autorelease: pending '
                         '--add-label autorelease: tagged'] if number else [])
            assert calls[1:] == expected, calls
print('Release tagging selects only merged release PRs for its own branch; empty selection is safe')
