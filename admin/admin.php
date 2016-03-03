<?php 

require_once('login-auth.php');

login_auth(true);

require_once('load.php');

require_once('admin-hooks.php');

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
	<link rel="stylesheet" type="text/css" href="styles/admin.css.php" />
	<script type="text/javascript" src="//ajax.lug.ustc.edu.cn/ajax/libs/jquery/2.1.4/jquery.min.js"></script>
	<?php apply_hooks('admin_head'); ?>
</head>
<body>
<div>
	<div id="admin-top">
		<?php apply_hooks('admin_top'); ?>
	</div>
	<div id="admin-left">
		<ul>
<?php 
			apply_hooks('admin_left'); ?>
		</ul>
	</div>
	<div id="admin-wrap">
<?php }


function admin_footer($arg=[]) { ?>
	<div id="thanksbar" style="display: none; color: #B070B0; text-shadow: 1px 1px 4px; margin-top: 3em; font-style: italic; height: 1.5em;">
		<p style="float: left;">感谢使用 TaoBlog 进行创作，您将感受到什么叫作自由。</p>
		<p style="float: right;">Version: 0.0.0</p>
	</div>
	</div><!-- admin-wrap -->
</div><!-- admin-main -->
<?php apply_hooks('admin_footer'); ?>
</body>
</html>
<?php } 

