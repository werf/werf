# -*- coding: utf-8 -*-

# TeePopen is a reader from stdout and stderr of subprocess.Popen into
# pipes to return to ansible and into unix socket to send to werf logger (logboek).
#
# stdout from subprocess.Popen (original_popen)
# |
# | looped read by TeeSplitter and write into 2 streams:
# ↓--------------------------↓
# stdout_live_sock           fd_out_write
# |                          ↓
# |                          fd_out_read
# ↓                          ↓
# LiveStdoutListener         exec_command in AnsiballZ's basic.py
# in live.py
#
# The same is happen with stderr.

import os
import select
import socket
import time
import threading

from werf import STDOUT_UNIX_SOCK_NAME


class TeePopen(object):
    def __init__(self, args, original_popen=None, bufsize=0, **kwargs):
        self.returncode = None
        if original_popen is not None:
            # pipe for stdout back to ansible
            self.fd_out_read, self.fd_out_write = self.pipe_cloexec()
            self.stdout = os.fdopen(self.fd_out_read, 'rb', bufsize)
            self.stdout_back_to_ansible = os.fdopen(self.fd_out_write, 'wb', 0)
            # pipe for stderr back to ansible
            self.fd_err_read, self.fd_err_write = self.pipe_cloexec()
            self.stderr = os.fdopen(self.fd_err_read, 'rb', bufsize)
            self.stderr_back_to_ansible = os.fdopen(self.fd_err_write, 'wb', 0)

            # unix socket for stdout redirect
            self.stdout_live_sock=None
            self.stdout_live_sock = self._open_live_sock(STDOUT_UNIX_SOCK_NAME)
            # unix socket for stderr redirect
            #self.stderr_live_sock = self._open_live_sock(STDERR_UNIX_SOCK_NAME)

            self.cmd = original_popen(args, bufsize=bufsize, **kwargs)
            self.stdin = self.cmd.stdin

            self._splitter = TeeSplitter(
                in_a=self.cmd.stdout,
                in_b=self.cmd.stderr,
                out_a=self.stdout_back_to_ansible,
                out_b=self.stderr_back_to_ansible,
                out_ab=self.stdout_live_sock,
            )
            self._splitter.start()

            # Periodically call poll to check if cmd is done and close fd_out_write
            # and fd_err_write for nonblocking communicate method.
            self.poll_thread = threading.Thread(target=self._poll_checker_thread)
            self.poll_thread.setDaemon(True)
            self.poll_thread.start()

    # from suprocess.Popen
    def pipe_cloexec(self):
        """Create a pipe with FDs set CLOEXEC."""
        # Pipes' FDs are set CLOEXEC by default because we don't want them
        # to be inherited by other subprocesses: the CLOEXEC flag is removed
        # from the child's FDs by _dup2(), between fork() and exec().
        # This is not atomic: we would need the pipe2() syscall for that.
        r, w = os.pipe()
        self._set_cloexec_flag(r)
        self._set_cloexec_flag(w)
        return r, w

    # from suprocess.Popen
    def _set_cloexec_flag(self, fd, cloexec=True):
        import fcntl
        try:
            cloexec_flag = fcntl.FD_CLOEXEC
        except AttributeError:
            cloexec_flag = 1

        old = fcntl.fcntl(fd, fcntl.F_GETFD)
        if cloexec:
            fcntl.fcntl(fd, fcntl.F_SETFD, old | cloexec_flag)
        else:
            fcntl.fcntl(fd, fcntl.F_SETFD, old & ~cloexec_flag)

    def _open_live_sock(self, filename):
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        try:
            sock.connect(filename)
        except socket.error:
            raise
        return sock

    # poll cmd while not None. close stdout and stderr to unblock communicate method
    def _poll_checker_thread(self):
        while True:
            res = self.poll()
            time.sleep(0.1)
            if res is not None:
                break

    def poll(self):
        self.returncode = self.cmd.poll()
        if self.returncode is not None:
            # cmd is finished, so stop splitter thread
            self._splitter.stop()
            self._splitter.join(1)
            # close sockets
            self.stdout_live_sock.close()
            self.stdout_back_to_ansible.close()
            self.stderr_back_to_ansible.close()
        return self.returncode

    def wait(self):
        self.returncode = self.cmd.wait()
        return self.returncode

    def communicate(self, args):
        self.cmd.stdout = self.stdout
        self.cmd.stderr = self.stderr
        stdout, stderr = self.cmd.communicate(args)
        return stdout, stderr


class TeeSplitter(threading.Thread):
    def __init__(self, in_a=None, in_b=None, out_a=None, out_b=None, out_ab=None):
        self.in_a = in_a
        self.in_b = in_b
        self.out_a = out_a
        self.out_b = out_b
        self.out_ab = out_ab
        self._stop = False
        threading.Thread.__init__(self)
        self.setDaemon(True)

    def stop(self):
        self._stop = True

    def run(self):
        rpipes = [self.in_a, self.in_b]
        while True:
            rfds, wfds, efds = select.select(rpipes, [], [], 0.1)

            for s in rfds:
                data = self._read_from_pipes(rpipes, rfds, s)
                if s is self.in_a:
                    self.write(self.out_a, data)
                if s is self.in_b:
                    self.write(self.out_b, data)
                self.write(self.out_ab, data)

            if self._stop:
                break

            if not rpipes:
                break

    def _read_from_pipes(self, rpipes, rfds, file_descriptor):
        data = ''
        if file_descriptor in rfds:
            data = os.read(file_descriptor.fileno(), 9000)
            if data == '':
                rpipes.remove(file_descriptor)

        return data

    def write(self, s, data):
        if isinstance(s, socket.socket):
            s.sendall(data)
        elif isinstance(s, int):
            os.write(s, data)
        elif isinstance(s, file):
            s.write(data)
        else:
            raise TypeError(type(s))
