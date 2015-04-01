<?php

require_once(dirname(__FILE__).'/../setup/config.php');
require_once('die.php');
require_once('db/dbbase.php');

function login_auth($redirect=false) {
	require_once('db/options.php');

	$opt = new TB_Options;

	$ip = $_SERVER['REMOTE_ADDR'];
	$auth_ips = $opt->get('auth_ips');
	$ipauth = in_array($ip, explode(',', $auth_ips)) === true;
	$hash = isset($_COOKIE['login']) ? $_COOKIE['login'] : '';
	$loggedin = $hash && $hash === sha1(md5($ip).$opt->get('login'));
	$loggedin = $loggedin && $ipauth;
	if(!$loggedin) {
		if($redirect) {
			$home = $opt->get('home');
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

