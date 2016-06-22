<?php

if(!isset($_SERVER['HTTP_USER_AGENT'])) {
	header('HTTP/1.1 400 Bad Request');
	die(-1);
}

$start_time = microtime();

require('admin/load.php');
require('theme/functions.php');

// maintenance mode
// https://yoast.com/http-503-site-maintenance-seo/
if(!login_auth() && file_exists('MAINTENANCE')) {
    header('HTTP/1.1 503 In Maintenance');
    header('Content-Type: text/plain; charset=utf-8');
    header('Retry-After: 300');
    echo '网站维护中，请稍后再访问...';
    die(-1);
}


if($tbquery->query() === false){
	tb_die(400, '未定义的查询！');
}

// https://css-tricks.com/snippets/php/count-script-excecution-time/
$execution_time = sprintf('%.5f', microtime() - $start_time);

if(!$tbquery->have() 
    && !$tbquery->is_home() 
    && !$tbquery->is_sitemap()
    && !$tbquery->is_archive()
    && !$tbquery->is_shuoshuo()
){
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
} else if($tbquery->is_archive()) {
    require('theme/archive.php');
} else if($tbquery->is_shuoshuo()) {
    require('theme/shuoshuo.php');
}

