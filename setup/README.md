# 安装手册

## 创建数据库

```sql
CREATE DATABASE taoblog;
```

## 创建用户及授权

```sql
CREATE USER "用户名"@"%" IDENTIFIED BY "密码";
GRANT ALL ON taoblog.* TO "用户名"@"%";
```

## 初始化

```bash
$ mysql -u taoblog -p --database=taoblog < data/schemas.sql
```

```sql
INSERT INTO options (name,value) VALUES ('db_ver', 12);
```
