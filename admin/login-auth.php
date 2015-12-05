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

function login_auth_ip() {
	require_once('db/options.php');

	$opt = new TB_Options;

	$ipauth = false;
	$ip = $_SERVER['REMOTE_ADDR'];
	$auth_ips = explode(',', $opt->get('auth_ips'));
	foreach($auth_ips as &$aip) {
		if($aip && preg_match($aip, $ip)) {
			$ipauth = true;
			break;
		}
	}

	return $ipauth;
}

function login_auth($redirect=false) {
	require_once('db/options.php');

	$opt = new TB_Options;

	$ipauth = login_auth_ip();

	$ip = $_SERVER['REMOTE_ADDR'];
	$hash = isset($_COOKIE['login']) ? $_COOKIE['login'] : '';
	$loggedin = $hash && $hash === sha1(md5($ip).$opt->get('login'));

	$loggedin = $loggedin && $ipauth;
	if(!$loggedin) {
		if($redirect) {
            $is_ssl = $_SERVER['SERVER_PORT'] == 443;
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
	setcookie('login', sha1(md5($ip).$opt->get('login')), 0, '/', '', false, true);
}

