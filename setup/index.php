<?php

// 请在安装成功之后取消注释下面一条语句
die('Setup: Silence is golden.');

require_once(dirname(__FILE__).'/../admin/die.php');
require_once('config.php');
require_once(dirname(__FILE__).'/../admin/db/dbbase.php');

$my = $tbdb;

// 创建表 options
$sql = "CREATE TABLE IF NOT EXISTS `options` (
	`id` INT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	`name` VARCHAR(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`value` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	PRIMARY KEY(`id`)
	);";

if(!$my->query($sql)){
	tb_die(200, '无法创建表：options - '.$my->error);
}

// 创建表 posts

$sql = "CREATE TABLE IF NOT EXISTS `posts` (
	`id` INT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	`date` DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00',
	`modified` DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00',
	`title` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`content` LONGTEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`slug` VARCHAR(128) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`type` VARCHAR(16) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`taxonomy` INT(20) UNSIGNED NOT NULL DEFAULT 1,
	`status` VARCHAR(16) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL, 
	`comment_status` INT(1) UNSIGNED DEFAULT 1,
	`password` CHAR(40) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
	PRIMARY KEY(`id`)
	);";

if(!$my->query($sql)){
	tb_die(200, '无法创建表：posts - '.$my->error);
}

// 创建表 comments
$sql = "CREATE TABLE IF NOT EXISTS `comments` (
	`id` INT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	`post_id` INT(20) UNSIGNED NOT NULL,
	`author` TINYTEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`email` VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`url` VARCHAR(200) CHARACTER SET utf8 COLLATE utf8_general_ci,
	`ip` VARCHAR(16) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`date` DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00',
	`content` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`agent` VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
	`status` VARCHAR(16) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT 'approved',
	`parent` INT(20) UNSIGNED NOT NULL,
	`ancestor` INT(20) UNSIGNED NOT NULL,
	PRIMARY KEY(`id`)
	);";

if(!$my->query($sql)){
	tb_die(200, '无法创建表：comments - '.$my->error);
}

// 创建表 taxonomies
$sql = "CREATE TABLE IF NOT EXISTS taxonomies (
	`id` INT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	`name` VARCHAR(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`slug` VARCHAR(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`parent` INT(20) UNSIGNED NOT NULL,
	`ancestor` INT(20) UNSIGNED NOT NULL,
	PRIMARY KEY(`id`)
	);";
if(!$my->query($sql)){
	tb_die(200, '无法创建表：taxonomies - '.$my->error);
}

// 创建表 /分类/文章样式和JavaScript/关键字表 post_metas
$sql = "CREATE TABLE IF NOT EXISTS post_metas (
	`id` INT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	`type` ENUM('post','page','tax'),
	`tid` INT(20) UNSIGNED NOT NULL,
	`header` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`footer` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	`keywords` TEXT CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
	PRIMARY KEY(`id`)
	);";
if(!$my->query($sql)) {
	tb_die(200, '无法创建表：post_metas - '.$my->error);
}

tb_die(200, '操作成功！');

