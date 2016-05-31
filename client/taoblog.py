#!/usr/bin/env python3

import sys
import requests
import getpass
import json

from subprocess import call

class TaoBlog:
    _host    = 'https://blog.twofei.com'
    _login   = ''
    _verify  = True

    def method(self, name):
        return self._host + '/api/' + name.replace('.', '/')

    def post(self, method, data):
        return requests.post(self.method(method), data=data, verify=self._verify)

    def login(self):
        username = input('Username: ')
        password = getpass.getpass('Password: ')

        print('Logging, please wait ...')
        data = {'user': username, 'passwd': password}
        r = self.post('login.auth', data)
        if r.status_code != 200:
            sys.exit(-1)

        r = json.loads(r.text)
        if r["ret"] != 0:
            print(r["msg"])
            sys.exit(-1)

        self.login = r["data"]["login"]
        print('Login success, cookie: %s' % self.login)

    def main(self):
        self.login()

if __name__ == '__main__':
    blog = TaoBlog()
    blog.main()

