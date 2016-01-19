<?php

require_once(dirname(__FILE__).'/../setup/config.php');
require_once('die.php');
require_once('db/dbbase.php');

function login_auth_passwd($arg = []) {
	require_once('db/options.php');
	$opt = new TB_Options;

	$user = isset($arg['user']) ? $arg['user'] : '';
	$passwd = isset($arg['passwd']) ? $arg['passwd'] : '';
	$ip = $_SERVER['REMOTE_ADDR'];

	if($user === 'twofei' && $passwd === sha1(md5($ip).$opt->get('login'))) {
		return true;
	}

	return false;
}

// 后期将取消非HTTPS登录许可，也就不再需要验证IP
function login_auth_ip() {
    $is_ssl = $_SERVER['SERVER_PORT'] == 443;

    return $is_ssl;
}

function login_auth($redirect=false) {
	require_once('db/options.php');

	$opt = new TB_Options;

    $is_ssl = $_SERVER['SERVER_PORT'] == 443;

	$ipauth = login_auth_ip();

	$ip = $_SERVER['REMOTE_ADDR'];
	$hash = isset($_COOKIE['login']) ? $_COOKIE['login'] : '';
	$loggedin = $hash && $hash === sha1(md5($ip).$opt->get('login'));

	$loggedin = $loggedin && $ipauth;
	if(!$loggedin) {
		if($redirect) {
			$home = ($is_ssl ? 'https://' : 'http://') . $opt->get('home');
			$url = $home.'/admin/login.php?url='.urlencode($_SERVER['REQUEST_URI']);

			header('HTTP/1.1 302 Login Required');
			header('Location: '.$url);
			die(0);
		}

		return false;
	} else {
		return true;
	}
}

function login_auth_cookie($redirect=false) {
	return login_auth($redirect);
}

function login_auth_set_cookie($ip) {
	$opt = new TB_Options;
	setcookie('login', sha1(md5($ip).$opt->get('login')), 0, '/', '', true, true);
}

