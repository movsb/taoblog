<?php
/**
 * Hooks of admin page
 */

defined('TBPATH') or die("Silence is golden.");

add_hook('admin_left', 'adminLeftAll');

/**
 * Hooks of admin left column
 * 
 * @return nothing
 */
function adminLeftAll()
{
    echo '<li><a href="/">首页</a></li>';
    echo '<li><a href="post.php">发表文章</a></li>';
    echo '<li><a href="post.php?type=page">发表页面</a></li>';
    echo '<li><a href="post-manage.php">文章管理</a></li>';
    echo '<li><a href="taxonomy.php">分类管理</a></li>';
    echo '<li><a href="tag-manage.php">标签管理</a></li>';
    echo '<li><a href="settings.php">设置</a></li>';
    echo '<li><a href="login.php?do=logout">退出</a></li>';
}
