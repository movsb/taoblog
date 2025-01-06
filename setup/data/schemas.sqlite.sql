BEGIN;

-- 创建表 options
CREATE TABLE IF NOT EXISTS `options` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(64)  NOT NULL UNIQUE COLLATE NOCASE,
    `value` TEXT  NOT NULL
);

-- 创建表 posts
CREATE TABLE IF NOT EXISTS `posts` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `date` INTEGER NOT NULL,
    `date_timezone` TEXT NOT NULL,
    `modified` INTEGER NOT NULL,
    `modified_timezone` TEXT NOT NULL,
    `last_commented_at` INTEGER NOT NULL,
    `title` TEXT  NOT NULL,
    `slug` TEXT NOT NULL,
    `type` TEXT NOT NULL,
    `category` INTEGER  NOT NULL,
    `status` TEXT NOT NULL,
    `page_view` INTEGER  NOT NULL,
    `comment_status` INTEGER NOT NULL,
    `comments` INTEGER  NOT NULL,
    `metas` TEXT  NOT NULL,
    `source` TEXT NOT NULL,
    `source_type` TEXT NOT NULL
);

-- 创建表 comments
CREATE TABLE IF NOT EXISTS `comments` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `post_id` INTEGER  NOT NULL,
    `author` TEXT  NOT NULL,
    `email` TEXT NOT NULL,
    `url` TEXT NOT NULL,
    `ip` TEXT NOT NULL,
    `date` INTEGER NOT NULL,
    `date_timezone` TEXT NOT NULL,
    `modified` INTEGER NOT NULL,
    `modified_timezone` TEXT NOT NULL,
    `source_type` TEXT NOT NULL,
    `source` TEXT NOT NULL,
    `parent` INTEGER NOT NULL,
    `root` INTEGER NOT NULL
);
-- 奇怪，为什么不能直接在 create table 的时候写？
-- sqlite3 好像没有创建普通索引的写法？
-- https://sqlite.org/syntax/column-constraint.html
CREATE INDEX `post_id` on `comments` (`post_id`);

-- 创建表 tag标签/post_tags
CREATE TABLE IF NOT EXISTS tags (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` TEXT NOT NULL UNIQUE COLLATE NOCASE,
    `alias` INTEGER NOT NULL
);

-- 创建表 文章标签表
CREATE TABLE IF NOT EXISTS post_tags (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `post_id` INTEGER  NOT NULL,
    `tag_id` INTEGER  NOT NULL,
    UNIQUE (`post_id`,`tag_id`)
);

CREATE TABLE IF NOT EXISTS logs (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `time` INTEGER NOT NULL,
    `type` TEXT NOT NULL COLLATE NOCASE,
    `sub_type` TEXT NOT NULL COLLATE NOCASE,
    `version` INTEGER NOT NULL,
    `data` TEXT NOT NULL
);

COMMIT;
