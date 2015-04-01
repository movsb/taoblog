<?php

function ajax_die($code, $str) {
	$j = [
		'errno' => $code,
		'error' => $str,
		];
	
	header('HTTP/1.1 200 OK');
	header('Content-Type: application/json');
	echo json_encode($j);
	die(0);
}

function ajax_done($arg){
	header('HTTP/1.1 200 OK');
	echo json_encode($arg);
	die(0);
}

require('load.php');

if(!isset( $_REQUEST['do'])){
	ajax_die(400, '未指定动作！');
}

$do = str_replace('-','_','ajax-'.$_REQUEST['do']);

if(!function_exists($do)){
	ajax_die(400, "未找到Ajax函数($do)！");
}

$do();

die(0);

