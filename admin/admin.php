<?php 

require_once('login-auth.php');

login_auth(true);

require_once('load.php');

function the_header() {
    header('HTTP/1.1 200 OK');
    header('Cache-Control: no-cache, must-revalidate, max-age=0');
    header('Pragma: no-cache');
    header('Expires: Wed, 11 Jan 1984 05:00:00 GMT');
}


function admin_header($arg=[]) { 
    the_header(); 
?>
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <link rel="stylesheet" type="text/css" href="admin.css" />
    <meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1, maximum-scale=1" />
    <script type="text/javascript" src="//blog-10005538.file.myqcloud.com/jquery.min.js"></script>
    <?php apply_hooks('admin_head'); ?>
</head>
<body>
<div>
    <div id="admin-left">
        <ul>
            <li><a href="/">首页</a></li>
            <li><a href="post.php">发表文章</a></li>
            <li><a href="post.php?type=page">发表页面</a></li>
            <li><a href="post-manage.php">文章管理</a></li>
            <li><a href="taxonomy.php">分类管理</a></li>
            <li><a href="tag-manage.php">标签管理</a></li>
            <li><a href="login.php?do=logout">退出</a></li>
        </ul>
    </div>
    <div id="admin-wrap">
<?php }


function admin_footer($arg=[]) { ?>
    </div><!-- admin-wrap -->
</div><!-- admin-main -->
<?php apply_hooks('admin_footer'); ?>
</body>
</html>
<?php } 
