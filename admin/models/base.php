<?php

/**
 * 建立全局的数据库连接
 *
 * 该文件被所有其它需要建立数据库连接的文件所包含
 * 在包含该文件之前需要先包含数据库配置文件（/setup/config.php）
 */

$tbdb = @new mysqli(DB_HOST, DB_USER, DB_PASSWORD, DB_NAME);
if ($tbdb->connect_error) {
    tb_die(503, '连接数据库失败：'.$tbdb->connect_error);
}

if (!$tbdb->set_charset("utf8mb4")) {
    tb_die(503, '无法设置字符集：'.$tbdb->connect_error);
}

