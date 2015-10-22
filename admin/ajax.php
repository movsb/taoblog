<?php

function ajax_die($code, $str) {
	$j = [
		'errno' => $code,
		'error' => $str,
		];
	
	header('HTTP/1.1 200 OK');
	header('Content-Type: application/json; charset=utf-8');
	echo json_encode($j);
	die(0);
}

function ajax_done($arg){
	header('HTTP/1.1 200 OK');
	header('Content-Type: application/json; charset=utf-8');
	echo json_encode($arg);
	die(0);
}

require('load.php');

function ajax_get_cat_posts() {
    global $tbpost;
    $cid = isset($_REQUEST['cid']) ? (int)$_REQUEST['cid']: 0;
    $posts = $tbpost->get_cat_posts($cid);
    if(!is_array($posts)) {
        ajax_die('err','获取失败');
    }

    ajax_done([
        'errno'     => 'ok',
        'posts'     => $posts,
        ]);
}

function ajax_get_date_posts() {
    global $tbpost;
    $yy = isset($_REQUEST['yy']) ? (int)$_REQUEST['yy'] : 0;
    $mm = isset($_REQUEST['mm']) ? (int)$_REQUEST['mm'] : 0;

    if ($yy < 1970                  // 我假装我的博客也是诞生于大纪元1970年，有问题么:-)
        || ($mm < 1 || $mm > 12)    // 听说公司发15个月的工资，我很happy！
    ){
        ajax_die(400, '你我不在同一个世界？');
    }

    $posts = $tbpost->get_date_posts($yy, $mm);
    if(!is_array($posts)){
        ajax_die(400, '获取失败');
        return false;
    }

    ajax_done([
        'errno'     => 'ok',
        'posts'     => $posts,
        ]);
}

if(!isset( $_REQUEST['do'])){
	ajax_die(400, '未指定动作！');
}

$do = str_replace('-','_','ajax-'.$_REQUEST['do']);

if(!function_exists($do)){
	ajax_die(400, "未找到Ajax函数($do)！");
}

$do();

die(0);

