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
    `date` INTEGER NOT NULL,
    `modified` INTEGER NOT NULL,
    `title` TEXT  NOT NULL,
    `content` TEXT  NOT NULL,
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
    `source_type` TEXT NOT NULL,
    `source` TEXT NOT NULL,
    `content` TEXT NOT NULL,
    `parent` INTEGER NOT NULL,
    `root` INTEGER NOT NULL
);

-- 创建表 文章分类 categories
CREATE TABLE IF NOT EXISTS categories (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `name` TEXT NOT NULL,
    `slug` TEXT NOT NULL,
    `parent_id` INTEGER NOT NULL,
    `path` TEXT NOT NULL,
    UNIQUE (`path`,`slug`)
);

-- 创建表 tag标签/post_tags
CREATE TABLE IF NOT EXISTS tags (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` TEXT NOT NULL,
    `alias` INTEGER NOT NULL,
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
