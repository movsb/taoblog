<?php 
header('HTTP/1.1 404 Not Found');
header('Content-Type: text/html; charset=utf-8');

if($tbquery->is_page()) {
    echo '404 页面未找到';
} else if($tbquery->is_post()) {
    echo '404 文章未找到';
} else if($tbquery->is_tag()) {
    echo '404 标签下不存在文章';
}
