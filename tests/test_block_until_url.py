#!/usr/bin/python2.7
# Copyright (c) 2013 The CoreOS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import BaseHTTPServer
import os
import select
import signal
import subprocess
import threading
import time
import unittest

script_path = os.path.abspath('%s/../../bin/block-until-url' % __file__)


class UsageTestCase(unittest.TestCase):

    def test_no_url(self):
        proc = subprocess.Popen([script_path],
                                stdout=subprocess.PIPE,
                                stderr=subprocess.PIPE)
        out, err = proc.communicate()
        self.assertEquals(proc.returncode, 1)
        self.assertEquals(out, '')
        self.assertIn('invalid url', err)

    def test_invalid_url(self):
        proc = subprocess.Popen([script_path, 'fooshizzle'],
                                stdout=subprocess.PIPE,
                                stderr=subprocess.PIPE)
        out, err = proc.communicate()
        self.assertEquals(proc.returncode, 1)
        self.assertEquals(out, '')
        self.assertIn('invalid url', err)


class TestRequestHandler(BaseHTTPServer.BaseHTTPRequestHandler):

    def send_test_data(self):
        if self.path == '/ok':
            ok_data = 'OK!\n'
            self.send_response(200)
            self.send_header('Content-type', 'text/plain')
            self.send_header('Content-Length', str(len(ok_data)))
            if self.command != 'HEAD':
                self.wfile.write(ok_data)
        elif self.path == '/404':
            self.send_error(404)
        else:
            # send nothing so curl fails
            pass

    def do_GET(self):
        self.send_test_data()

    def do_HEAD(self):
        self.send_test_data()

    def log_message(self, format, *args):
        pass


class HttpTestCase(unittest.TestCase):

    def setUp(self):
        self.server = BaseHTTPServer.HTTPServer(
                ('localhost', 0), TestRequestHandler)
        self.server_url = 'http://%s:%s' % self.server.server_address
        server_thread = threading.Thread(target=self.server.serve_forever)
        server_thread.daemon = True
        server_thread.start()

    def tearDown(self):
        self.server.shutdown()

    def test_quick_ok(self):
        proc = subprocess.Popen([script_path, '%s/ok' % self.server_url],
                                stdout=subprocess.PIPE,
                                stderr=subprocess.PIPE)
        out, err = proc.communicate()
        self.assertEquals(proc.returncode, 0)
        self.assertEquals(out, '')
        self.assertEquals(err, '')

    def test_quick_404(self):
        proc = subprocess.Popen([script_path, '%s/404' % self.server_url],
                                stdout=subprocess.PIPE,
                                stderr=subprocess.PIPE)
        out, err = proc.communicate()
        self.assertEquals(proc.returncode, 0)
        self.assertEquals(out, '')
        self.assertEquals(err, '')

    def test_timeout(self):
        proc = subprocess.Popen([script_path, '%s/bogus' % self.server_url],
                                bufsize=4096,
                                stdout=subprocess.PIPE,
                                stderr=subprocess.PIPE)
        timeout = time.time() + 2 # kill after 2 seconds
        while time.time() < timeout:
            time.sleep(0.1)
            self.assertIs(proc.poll(), None, 'script terminated early!')
        proc.terminate()
        out, err = proc.communicate()
        self.assertEquals(proc.returncode, -signal.SIGTERM)
        self.assertEquals(out, '')
        self.assertEquals(err, '')


if __name__ == '__main__':
    unittest.main()
