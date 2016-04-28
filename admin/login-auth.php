<?php

require_once(dirname(__FILE__).'/../setup/config.php');
require_once('die.php');
require_once('db/dbbase.php');

// 用于登录页面的验证
function login_auth_passwd($arg = []) {
	require_once('db/options.php');
	$opt = new TB_Options;

    $saved_login = explode(',', $opt->get('login'));
    if($saved_login === false || count($saved_login) != 2)
        return false;

    $saved_user     = $saved_login[0];
    $saved_passwd   = $saved_login[1];

	$user           = isset($arg['user']) ? $arg['user'] : '';
	$passwd         = isset($arg['passwd']) ? $arg['passwd'] : '';

	if($user === $saved_user && sha1($passwd) === $saved_passwd) {
		return true;
	}
    else {
        return false;
    }
}

// 用于通过cookie认证客户端
function login_auth($redirect=false) {
	require_once('db/options.php');

	$opt = new TB_Options;

    $is_ssl = $_SERVER['SERVER_PORT'] == 443;

	$ip = $_SERVER['REMOTE_ADDR'];
	$cookie_login = isset($_COOKIE['login']) ? $_COOKIE['login'] : '';

	$loggedin = $is_ssl && $cookie_login && $cookie_login === sha1($ip.$opt->get('login'));

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

// 用于在登录成功之后设置客户端认证的cookie
// 保存的是 sha1(ip + login)
function login_auth_set_cookie($ip) {
	$opt = new TB_Options;
	setcookie('login', sha1($ip.$opt->get('login')), 0, '/', '', true, true);
}

