<?php

require_once('login-auth.php');
require_once('die.php');
require_once(dirname(__FILE__).'/../setup/config.php');
require_once('db/dbbase.php');
require_once('db/options.php');

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
			box-shadow: 3px 3px 10px;
		}

		#login .title {
			height: 32px;
			font-size: 1.5em;
			text-align: center;
			padding: 0.2em;
		}

		#login .input {
			padding: 0.5em;
		}

		#login .input div {
			margin-top: 0.2em;
			margin-bottom: 0.2em;
		}

		#login .input input[type="text"], #login .input input[type="password"] {
			padding: 0.3em;
		}

		#login .input .submit {
			text-align: center;
			padding: 0.8em;
		}

	</style>
</head>
<body>
<div id="wrap">
	<form method="post" id="login">
		<div class="title">
			登录
		</div>
		<div class="input">
			<div>
				<label>用户</label>
				<input type="text" name="user" />
			</div>
			<div>
				<label>密码</label>
				<input type="password" name="passwd" />
			</div>
			<div class="submit">
				<input type="submit" value="登录" />
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

$opt1 = new TB_Options;

$do = isset($_GET['do']) ? $_GET['do'] : '';
if($do === 'logout') {
	header('HTTP/1.1 200 Logged Out');
	setcookie('login');
	//header('Location: '.$opt1->get('home').'/admin/login.php');
	login_html();
	die(0);
} else {
	if(login_auth()){
		header('HTTP/1.1 302 Login Success');

		$url = isset($_GET['url']) ? $_GET['url'] : $opt1->get('home').'/admin/';
		header('Location: '.$url);
		die(0);
	} else {
		$url = isset($_GET['url']) ? $_GET['url'] : '';
		login_html($url);
		die(0);
	}
}

else : // POST

$user = isset($_POST['user']) ? $_POST['user'] : '';
$passwd = isset($_POST['passwd']) ? $_POST['passwd'] : '';

$opt1 = new TB_Options;

if($user!=='twofei' ||  sha1(md5($_SERVER['REMOTE_ADDR']).sha1(md5($passwd).sha1($passwd))) !== sha1(md5($_SERVER['REMOTE_ADDR']).$opt1->get('login'))) {
	login_html();
	die(0);
}

$url = isset($_POST['url']) ? $_POST['url'] : $opt1->get('home').'/admin/';
header('HTTP/1.1 302 OK');
header('Location: '.$url);
setcookie('login', sha1(md5($_SERVER['REMOTE_ADDR']).$opt1->get('login')));
die(0);

endif;

