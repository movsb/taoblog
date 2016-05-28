#!/usr/bin/env python3

import sys
import requests
import getpass
import json

from subprocess import call

class TaoBlog:
    host    = 'https://blog.twofei.com'
    login   = '' 
    verify  = True

    def main(self):
        username = input('Username: ')
        password = getpass.getpass('Password: ')

        data = {'user': username, 'passwd': password}
        r = requests.post(self.host + '/api/login/auth', data=data, verify=self.verify)
        if r.status_code != 200:
            sys.exit(-1)

        r = json.loads(r.text)
        if r["ret"] != 0:
            print(r["msg"])
            sys.exit(-1)

        self.login = r["data"]["login"]

        print('cookie: ', self.login)

        while True:
            id = input("Post id to udpate: ")
            f = open('p'+id+'/index.html', 'rb')
            content = f.read()

            r = requests.post(self.host + '/api/post/update', {'id': id, 'content': content}, cookies={'login': self.login}, verify=self.verify)
            r = json.loads(r.text)

            if r["ret"] != 0:
                print("update error.");
            else:
                print("update succeeded.");

if __name__ == '__main__':
    blog = TaoBlog()
    blog.main()

