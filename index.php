<?php

if(!isset($_SERVER['HTTP_USER_AGENT'])) {
	header('HTTP/1.1 400 Bad Request');
	die(-1);
}

function theme_file($f)
{
    return 'themes/'.TB_THEME.'/'.$f;
}

require('admin/load.php');
require(theme_file('functions.php'));

// 是否对外开放
if(TB_PRIVATE === TRUE) {
    login_auth(true);
}

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

if(!$tbquery->have() 
    && !$tbquery->is_home() 
    && !$tbquery->is_sitemap()
    && !$tbquery->is_archive()
    && !$tbquery->is_memory()
){
	if(!$tbquery->is_query_modification) {
		require(theme_file('404.php'));
		die(0);
	} else {
		header('HTTP/1.1 304 Not Modified');
		die(0);
	}
}

if($tbquery->is_home())             require(theme_file('index.php'));
else if($tbquery->is_singular())    require(theme_file('single.php'));
else if($tbquery->is_category())    require(theme_file('category.php'));
else if($tbquery->is_date())        require(theme_file('date.php'));
else if($tbquery->is_tag())         require(theme_file('tag.php'));
else if($tbquery->is_feed())        require(theme_file('feed.php'));
else if($tbquery->is_sitemap())     require(theme_file('sitemap.php'));
else if($tbquery->is_archive())     require(theme_file('archive.php'));
else if($tbquery->is_memory())      require(theme_file('memory.php'));

