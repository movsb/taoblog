#!/usr/bin/env python3

import os
import sys
import requests
import getpass
import json

from subprocess import call

class TaoBlog:
    _host    = 'https://blog.twofei.com'
    _login   = ''
    _verify  = True
    _root    = '.'

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
        text = ('1. 更新文章到远程\n'
                '2. 更新文章到本地\n'
                '3. 获取文章标签列表\n'
                '4. 更新文章标签列表'
                )
        print(text)

    def get_post_content(self, id):
        """
            有以下几种路径搜索方式：
                <id>.md         ->  content.html
                <id>.html       ->  <id>.html
                <id>/index.md   ->  <id>/content.html
                <id>/index.html ->  <id>/index.html
        """

        if os.path.exists("%s/%s.md" % (self._root, id)):
            path = "content.html"
        elif os.path.exists("%s/%s.html" % (self._root, id)):
            path = "%s.html" % id
        elif os.path.exists("%s/%s/index.md" % (self._root, id)):
            path = "%s/content.html" % id
        elif os.path.exists("%s/%s/index.html" % (self._root, id)):
            path = "%s/index.html" % id
        else:
            exit(-1)

        path = self._root + "/" + path

        fp = open(path, 'rb')
        content = fp.read()
        fp.close()

        return content

    def save_post_content(self, id, content):
        path = "%s/%s.html" % (self._root, id)
        if not os.path.exists(path):
            path = "%s/%s/index.html" % (self._root, id)
            if not os.path.exists(path):
                path = "%s/%s.html" % (self._root, id)

        fp = open(path, 'wb')
        fp.write(bytes(content,'UTF-8'))
        fp.close()

    def cmd(self):
        while True:
            n = input("输入选项：")
            if n == "1":
                id = input("文章ID: ")
                content = self.get_post_content(id)
                r = self.post('post.update', data={'id': id, 'content': content})
                print(r.text)

            elif n == "2":
                id = input("文章ID: ")
                r = self.post('post.get', data={'id': id})
                r = json.loads(r.text)
                self.save_post_content(id, r["data"]["content"])

            elif n == "3":
                id = input("文章ID: ")
                r = self.post('tag.get', data={'pid': id})
                print(r.text)

            elif n == "4":
                id = input("文章ID: ")
                tags = input("新的标签列表: ")
                r = self.post('tag.update', data={'pid': id, 'tags': tags})
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


