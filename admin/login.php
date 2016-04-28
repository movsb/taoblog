<?php

require_once('login-auth.php');
require_once('die.php');
require_once(dirname(__FILE__).'/../setup/config.php');
require_once('db/dbbase.php');
require_once('db/options.php');

// 登录相关的请求全部在 https 下进行
if($_SERVER['SERVER_PORT'] != 443) {
	header('HTTP/1.1 302 Unauthorized access');
	header('Location: /');
	die(0);
}

function login_html($url='') { ?>
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8" />
	<title>登录 - TaoBlog</title>
	<style>
		body {
			background-color: #646464;
		}

		#wrap {
			margin: 12.5% auto;
			background-color: #F1F1F1;
			width: 300px;
		}

		#login {
			width: 300px;
			height: 188px;
			box-shadow: 0px 0px 10px;
		}

		#login .title {
			height: 32px;
			font-size: 1.5em;
			text-align: center;
			padding: 0.2em;
		}

		#login .input input[type="text"], #login .input input[type="password"] {
			border: 1px solid #ccc;
			padding: 6px;
		}

	</style>
</head>
<body>
<div id="wrap">
	<form method="post" id="login">
		<div class="title">
			登录
		</div>
		<div style="padding: 10px 20px 10px;">
			<div class="input" style="text-align: center; margin-bottom: 15px;">
				<input type="text" name="user" placeholder="用户名" style="margin-bottom: 10px; width: 248px;"/>
				<input type="password" name="passwd" placeholder="密码" style="width: 248px;"/>
			</div>	
			<div class="submit" style="text-align: right;">
				<input type="submit" value="登录" style="padding: 4px 6px;"/>
			</div>
		</div>
		<div class="hidden">
		<?php if($url) { ?>
			<input type="hidden" name="url" value="<?php echo $url; ?>" />
		<?php } ?>
		</div>
	</form>
</div>
</body>
</html>

<?php } 

if($_SERVER['REQUEST_METHOD'] === 'GET') :

$do = isset($_GET['do']) ? $_GET['do'] : '';
if($do === 'logout') {
	header('HTTP/1.1 302 Logged Out');
	setcookie('login','',time()-1, '/');
    // 转到登录页面以验证成功退出
    // 如果没有成功退出，那么会因认证成功而转到管理员页面
	header('Location: login.php');
	die(0);
} else {
	if(login_auth()){
		header('HTTP/1.1 302 Login Success');

		$url = isset($_GET['url']) ? $_GET['url'] : '/admin/';
		header('Location: '.$url);
		die(0);
	} else {
		$url = isset($_GET['url']) ? $_GET['url'] : '';
		login_html($url);
		die(0);
	}
}

else : // POST

if(!login_auth_passwd($_POST)) {
	login_html();
	die(0);
}

$url = isset($_POST['url']) ? $_POST['url'] : '/admin/';
header('HTTP/1.1 302 OK');
header('Location: '.$url);
login_auth_set_cookie($_SERVER['REMOTE_ADDR']);
die(0);

endif;

