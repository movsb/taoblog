<?php

$start_time = microtime();

require('admin/load.php');
require('theme/functions.php');


if($tbquery->query() === false){
	tb_die(200, '未定义的查询！');
}

// https://css-tricks.com/snippets/php/count-script-excecution-time/
$execution_time = sprintf('%.5f', microtime() - $start_time);

if(!$tbquery->have() && !$tbquery->is_home() && !$tbquery->is_sitemap()){
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
} else if($tbquery->is_date()) {
	require('theme/date.php');
} else if($tbquery->is_tag()) {
	require('theme/tag.php');
} else if($tbquery->is_feed()) {
	require('theme/feed.php');
} else if($tbquery->is_sitemap()) {
	require('theme/sitemap.php');
}

