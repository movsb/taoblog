BEGIN;

-- 创建表 options
CREATE TABLE IF NOT EXISTS `options` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(64)  NOT NULL UNIQUE,
    `value` TEXT  NOT NULL
);

-- 创建表 posts
CREATE TABLE IF NOT EXISTS `posts` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `date` DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00',
    `modified` DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00',
    `title` TEXT  NOT NULL,
    `content` LONGTEXT  NOT NULL,
    `slug` VARCHAR(128)  NOT NULL,
    `type` VARCHAR(16)  NOT NULL,
    `category` INTEGER  NOT NULL DEFAULT 1,
    `status` VARCHAR(16),
    `page_view` INTEGER  NOT NULL DEFAULT 0,
    `comment_status` INTEGER  DEFAULT 1,
    `comments` INTEGER  NOT NULL DEFAULT 0,
    `metas` TEXT  NOT NULL,
    `source` LONGTEXT ,
    `source_type` VARCHAR(16)
);

-- 创建表 comments
CREATE TABLE IF NOT EXISTS `comments` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `post_id` INTEGER  NOT NULL,
    `author` TINYTEXT  NOT NULL,
    `email` VARCHAR(100)  NOT NULL,
    `url` VARCHAR(200) ,
    `ip` VARCHAR(16)  NOT NULL,
    `date` DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00',
    `source_type` varchar(16) NOT NULL,
    `source` text NOT NULL,
    `content` TEXT  NOT NULL,
    `parent` INTEGER  NOT NULL,
    `root` INTEGER  NOT NULL
);

-- 创建表 文章分类 categories
CREATE TABLE IF NOT EXISTS categories (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(32)  NOT NULL,
    `slug` VARCHAR(32)  NOT NULL,
    `parent_id` INTEGER  NOT NULL,
    `path` VARCHAR(256)  NOT NULL,
    UNIQUE (`path`,`slug`)
);

-- 创建表 tag标签/post_tags
CREATE TABLE IF NOT EXISTS tags (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(32)  NOT NULL,
    `alias` INTEGER  NOT NULL DEFAULT 0,
    UNIQUE(`name`)
);

-- 创建表 文章标签表
CREATE TABLE IF NOT EXISTS post_tags (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `post_id` INTEGER  NOT NULL,
    `tag_id` INTEGER  NOT NULL,
    UNIQUE (`post_id`,`tag_id`)
);

COMMIT;
