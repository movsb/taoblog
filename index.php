<?php
if(!isset($_SERVER['HTTP_USER_AGENT'])) {
	header('HTTP/1.1 400 Good Request');
	die(-1);
}

if(preg_match('/MSIE|Trident/', $_SERVER['HTTP_USER_AGENT'])		// IE
	&& !preg_match('/spider/i', $_SERVER['HTTP_USER_AGENT'])		// Spider, like stupid 360 haosou
	&& !(isset($_GET['iexplore']) && $_GET['iexplore'] == 'true')	// ! still use ie
	){
	header('HTTP/1.1 503 IE was history');
	header('Content-Type: text/html');
?><!doctype html>
<html>
<head>
<meta charset="UTF-8" />
<title>IE was history</title>
</head>
<body>
<p style="font-size: 2em;"><del>Internet Explorer</del> was history.</p>
<p style="color: red;">Please consider using another web browser like <b>fx</b>.</p>
<p style="color: red;">抱歉，本网站站长由于能力过分有限，无法将网站支持在 IE 浏览器上。请考虑换用其它浏览器，然后重新访问。</p>
<p>不管了，依然使用 IE 访问：<a href="<?php
	echo $_SERVER['REQUEST_URI'];
	if(isset($_SERVER['QUERY_STRING']) && $_SERVER['QUERY_STRING']) {
		echo '&iexplore=true';
	} else {
		echo '?iexplore=true';
	}?>">点击访问</a></p>
</body>
</html>
<?php
	die(0);
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
	tb_die(200, '未定义的查询！');
}

// https://css-tricks.com/snippets/php/count-script-excecution-time/
$execution_time = sprintf('%.5f', microtime() - $start_time);

if(!$tbquery->have() 
    && !$tbquery->is_home() 
    && !$tbquery->is_sitemap()
    && !$tbquery->is_archive()
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
}

