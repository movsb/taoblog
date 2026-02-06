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
    `user_id` INTEGER NOT NULL,
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
    `source_type` TEXT NOT NULL,
    `citations` TEXT NOT NULL
);

CREATE INDEX `idx_modified` on `posts` (`modified`);

-- 创建表 comments
CREATE TABLE IF NOT EXISTS `comments` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `post_id` INTEGER  NOT NULL,
    `user_id` INTEGER NOT NULL,
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

CREATE TABLE IF NOT EXISTS users (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `created_at` INTEGER NOT NULL,
    `updated_at` INTEGER NOT NULL,
    `nickname` TEXT NOT NULL,
    `email` TEXT NOT NULL,
    `password` TEXT NOT NULL,
    `otp_secret` TEXT NOT NULL,
    `credentials` TEXT NOT NULL,
    `google_user_id` TEXT NOT NULL,
    `github_user_id` INTEGER NOT NULL,
    `hidden` INTEGER NOT NULL,
    `avatar` TEXT NOT NULL,
    `bark_token` TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS acl (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `created_at` INTEGER NOT NULL,
    `user_id` INTEGER NOT NULL,
    `post_id` INTEGER NOT NULL,
    `permission` TEXT NOT NULL
);

CREATE TABLE categories (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `user_id` INTEGER NOT NULL,
    `parent_id` INTEGER NOT NULL,
    `name` TEXT  NOT NULL COLLATE NOCASE
);

CREATE UNIQUE INDEX `uix_cat_user_id__parent_id__name` ON `categories` (`user_id`,`parent_id`,`name`);

CREATE TABLE IF NOT EXISTS `files` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `parent_id` INTEGER NOT NULL,
    `created_at` INTEGER NOT NULL,
    `updated_at` INTEGER NOT NULL,
    `post_id` INTEGER NOT NULL,
    `path` TEXT NOT NULL,
    `mod_time` INTEGER  NOT NULL,
    `size` INTEGER  NOT NULL,
    `digest` TEXT NOT NULL,
    `meta` BLOB NOT NULL
);

CREATE UNIQUE INDEX `post_id__path` ON `files` (`post_id`,`path`);

CREATE INDEX `files_updated_at` on `files` (`updated_at`);

COMMIT;
