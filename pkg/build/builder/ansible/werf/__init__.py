# -*- coding: utf-8 -*-

def termPrint(args):
    import sys
    orig_stdout = sys.stdout
    term = open("/dev/pts/0", "wb+", buffering=0)
    sys.stdout = term
    print(args)
    sys.stdout = orig_stdout
    term.close()

STDOUT_UNIX_SOCK_NAME = '/.werf/ansible-tmpdir/local/ansible_live_stdout.sock'
STDERR_UNIX_SOCK_NAME = '/.werf/ansible-tmpdir/local/ansible_live_stderr.sock'
