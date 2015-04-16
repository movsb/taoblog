<?php

require('admin/load.php');

if($tbquery->query() === false){
	tb_die(200, '未定义的查询！');
}

if(!$tbquery->have() && !$tbquery->is_home()){
	if(!$tbquery->is_query_modification) {
		require('theme/404.php');
		die(0);
	} else {
		header('HTTP/1.1 304 Not Modified');
		die(0);
	}
}

if($tbquery->is_home()) {
	require('theme/index.php');
} else if($tbquery->is_singular()) {
	require('theme/single.php');
} else if($tbquery->is_category()) {
	require('theme/category.php');
} else if($tbquery->is_archive()) {
	require('theme/date.php');
}


