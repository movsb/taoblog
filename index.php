<?php

require('admin/load.php');

if($tbquery->query() === false){
	tb_die(200, '未定义的查询！');
}

if(!$tbquery->have()){
	if(!$tbquery->is_query_modification) {
		require('theme/404.php');
		die(0);
	} else {
		header('HTTP/1.1 304 Not Modified');
		die(0);
	}
}

if($tbquery->type === 'home') {
	require('theme/index.php');
	die(0);
} 

if(in_array($tbquery->type,['post','page'])) {
	require('theme/single.php');
	die(0);
}

