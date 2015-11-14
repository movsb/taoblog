# 安装
### 创建数据库 及 授权用户
进入 mysql 并顺序执行以下命令：
```bash
$ mysql -u <用户名> -p
Enter password: <密码>
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is xxxx to server version: x.xx.xx
 
Type 'help;' or '\h' for help. Type '\c' to clear the buffer.
 
mysql> CREATE DATABASE <数据库名>;
Query OK, 1 row affected (0.00 sec)
 
mysql> GRANT ALL PRIVILEGES ON <数据库名>.* TO "用户名"@"主机名"
    -> IDENTIFIED BY "密码";
Query OK, 0 rows affected (0.00 sec)
  
mysql> FLUSH PRIVILEGES;
Query OK, 0 rows affected (0.01 sec)
mysql> EXIT 
Bye
$ 
```
### 创建配置文件
复制 setup/config.sample.php 为 setup/config.php，根据配置文件说明以及上一步创建的数据库、数据库用户名及数据库密码作相应的修改；

评论时的通知邮箱只支持QQ域名邮箱；

百度推送用于实时推送文章到百度，TOKEN需要在百度站长获取；

### 运行安装文件：index.php
用文本编辑器打开 setup/index.php，注释掉第4行的 ```die('Setup: Silence is golden.')```，然后在浏览器中打开 /setup/index.php 进行安装，稍等片刻即可安装完成。

安装完成后，取消注释掉第4行，让其重新生效，以防止恶意脚本运行。

