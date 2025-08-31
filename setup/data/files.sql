BEGIN;

-- 设计上按文件 hash 唯一存储。
-- 所以：不能有相同内容而不同文件名的文件（更不能有多个空文件）。

-- 等等，为什么不直接存文件元信息表的编号呢？存 digest 干什么？
-- 编号不会冲突。

-- 暂时对于 0 大小文件不存数据。

CREATE TABLE IF NOT EXISTS `files` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `post_id` INTEGER NOT NULL,
    `digest` TEXT NOT NULL,
    `data` BLOB NOT NULL
);

CREATE UNIQUE INDEX `post_id__digest` ON `files` (`post_id`,`digest`);

COMMIT;
