<?php

if(!isset($_SERVER['HTTP_USER_AGENT'])) {
    header('HTTP/1.1 400 Bad Request');
    die(-1);
}

if(!file_exists('theme/') || !file_exists('theme/index.php')) {
    header('HTTP/1.1 503');
    echo 'Error: theme is not available.';
    die(-1);
}

require('admin/load.php');
require('theme/functions.php');

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
    tb_die(404, '404 页面未找到');
}

if(!$tbquery->have() 
    && !$tbquery->is_home() 
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

if($tbquery->is_home())             require('theme/index.php');
else if($tbquery->is_singular())    require('theme/single.php');
else if($tbquery->is_tag())         require('theme/tag.php');
else if($tbquery->is_archive())     require('theme/archive.php');
