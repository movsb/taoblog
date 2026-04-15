# sync

## 当前的同步方案

原始文件：文件名 + 哈希。

存储文件：

1. 如果不加密：objects/pid/原始文件哈希
2. 如果要加密：objects/pid/加密后的文件哈希

加密判定条件： `!post_public || (path[0] == '.' || path[0] == '_')`
