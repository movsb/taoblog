<?php

defined('TBPATH') or die("Silence is golden.");

add_hook('admin_left', 'admin_home_page');

function admin_home_page() {
?><li><a href="/">首页</a></li>
<?php 
}

add_hook('admin_left', 'shuoshuo_admin_left');

function shuoshuo_admin_left() {
?><li><a href="shuoshuo.php">发表说说</a></li>
<?php
}

add_hook('admin_left', 'post_admin_left');

function post_admin_left() {
?><li><a href="post.php">发表文章</a></li>
<?php 
}

add_hook('admin_left', 'page_admin_left');

function page_admin_left() {
?><li><a href="post.php?type=page">发表页面</a></li>
<?php 
}

add_hook('admin_left', 'post_manage_admin_left');

function post_manage_admin_left() {
?><li><a href="post-manage.php">文章管理</a></li>
<?php
}

add_hook('admin_left', 'tax_admin_left');

function tax_admin_left() {
?><li><a href="taxonomy.php">分类管理</a></li>
<?php 
}

add_hook('admin_left', 'admin_settings');

function admin_settings() {
?><li><a href="settings.php">设置</a></li>
<?php
}

add_hook('admin_left', 'admin_logout');

function admin_logout() {
?><li><a href="login.php?do=logout">退出</a></li>
<?php
}

