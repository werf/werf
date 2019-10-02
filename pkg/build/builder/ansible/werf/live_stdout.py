# -*- coding: utf-8 -*-

import threading
import os
import socket
import select

import logboek
from werf import STDOUT_UNIX_SOCK_NAME

'''

'''
class LiveStdoutListener(object):
    def __init__(self):
        self._stdout_sock = None
        self._stderr_sock = None
        self._stop = False
        self._enable_live_stdout = True
        self._live_stdout = False

    def start(self):
        self._stdout_sock = self._open_socket(STDOUT_UNIX_SOCK_NAME)
        #self._stderr_sock = self._open_socket(STDERR_UNIX_SOCK_NAME)
        self._reader = LiveStdoutReader(stdout_sock=self._stdout_sock, stderr_sock=self._stderr_sock, listener=self)
        self._reader.start()

    def stop(self):
        self._stop = True
        self._reader.stop()
        self._reader.join(10)
        self._stdout_sock.close()
        os.unlink(STDOUT_UNIX_SOCK_NAME)

    def _open_socket(self, filename):
        try:
            os.unlink(filename)
        except OSError:
            if os.path.exists(filename):
                raise

        # Create a UDS socket
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.bind(filename)

        return sock

    # is there was live output?
    def is_live_stdout(self):
        return self._enable_live_stdout and self._live_stdout

    # set live_stdout flag
    def set_live_stdout(self, live_stdout):
        self._live_stdout = live_stdout

    # is live output enabled?
    def is_enabled_live_stdout(self):
        return self._enable_live_stdout

    # enabled or disable live_output
    def set_enable_live_stdout(self, enable_live_stdout):
        self._enable_live_stdout = enable_live_stdout



class LiveStdoutReader(threading.Thread):
    def __init__(self, stdout_sock=None, stderr_sock=None, listener=None):
        self.stdout_sock=stdout_sock
        self.stderr_sock=stderr_sock
        self.listener=listener
        self._stop = False
        threading.Thread.__init__(self)
        self.setDaemon(True)

    def stop(self):
        self._stop = True

    def run(self):
        self.stdout_sock.listen(1)

        rsockets = [self.stdout_sock]
        while True:
            # timeout 100ms to stop quicker
            rfds, wfds, efds = select.select(rsockets, [], [], 0.1)

            for s in rfds:
                if s is self.stdout_sock:
                    connection, client_address = self.stdout_sock.accept()
                    rsockets.append(connection)
                else:
                    part = s.recv(1024)
                    if part:
                        self.listener.set_live_stdout(True)
                        if self.listener.is_enabled_live_stdout():
                            logboek.Out(part)
                    else:
                        s.close()
                        rsockets.remove(s)

            if self._stop:
                break

