<?php

$tbdb = @new mysqli(DB_HOST, DB_USER, DB_PASSWORD, DB_NAME);
if($tbdb->connect_error){
	tb_die(200, '连接数据库失败：'.$tbdb->connect_error);
}

if(!$tbdb->set_charset("utf8")){
	tb_die(200, '无法设置字符集：'.$tbdb->connect_error);
}

