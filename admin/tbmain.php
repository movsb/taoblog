<?php

class TB_Main {
    public function __construct() {
        global $tbopt;

        // 决定网站使用的协议
        $this->is_ssl = ($_SERVER['HTTPS'] ?? 'off') === 'on';

        // 网站主页地址（数据库里面不再保存使用协议）
        $this->home = ($this->is_ssl ? 'https://' : 'http://') . $tbopt->get('home');

    }
}

