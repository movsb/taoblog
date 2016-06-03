#!/usr/bin/env python3

import sys
import requests
import getpass
import json

from subprocess import call

class TaoBlog:
    _host    = 'https://local.twofei.com'
    _login   = ''
    _verify  = False

    def method(self, name):
        return self._host + '/api/' + name.replace('.', '/')

    def post(self, method, data):
        cookies ={'login': self._login}
        return requests.post(self.method(method), cookies=cookies, data=data, verify=self._verify)

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

        self._login = r["data"]["login"]
        print('Login success, cookie: %s' % self._login)

    def help(self):
        text = ('1. 发表说说\n'
                '2. 修改文章'
                )
        print(text)

    def cmd(self):
        while True:
            n = input("输入选项：")
            if n == "1":
                t = input("说说内容：")
                r = self.post('shuoshuo.post', data={'content': t})
                print(r.text)
            pass
        pass

    def main(self):
        self.login()
        self.help()
        self.cmd()

if __name__ == '__main__':
    blog = TaoBlog()
    blog.main()

