ALTER DATABASE taoblog COLLATE utf8mb4_unicode_ci;

USE taoblog;

-- option
ALTER TABLE `options` CHANGE `value` `value` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;

-- posts
ALTER TABLE `posts` CHANGE `title` `title` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `posts` CHANGE `content` `content` LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `posts` CHANGE `slug` `slug` VARCHAR(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `posts` CHANGE `metas` `metas` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `posts` CHANGE `source` `source` LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- comments
ALTER TABLE `comments` CHANGE `author` `author` TINYTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `comments` CHANGE `content` `content` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;

-- taxonomies
ALTER TABLE `taxonomies` CHANGE `name` `name` VARCHAR(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `taxonomies` CHANGE `slug` `slug` VARCHAR(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;

-- tags
ALTER TABLE `tags` CHANGE `name` `name` VARCHAR(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;

-- shuoshuo
ALTER TABLE `shuoshuo` CHANGE `content` `content` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `shuoshuo` CHANGE `source` `source` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- shuoshuo_comments
ALTER TABLE `shuoshuo_comments` CHANGE `author` `author` TINYTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `shuoshuo_comments` CHANGE `content` `content` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;
