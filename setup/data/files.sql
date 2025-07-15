BEGIN;

-- 设计上按文件 hash 唯一存储。
-- 所以：不能有相同内容而不同文件名的文件（更不能有多个空文件）。

CREATE TABLE IF NOT EXISTS `files` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `post_id` INTEGER NOT NULL,
    `digest` TEXT NOT NULL,
    `data` BLOB NOT NULL
);

CREATE UNIQUE INDEX `post_id__digest` ON `files` (`post_id`,`digest`);

COMMIT;
