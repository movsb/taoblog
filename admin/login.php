<?php

require_once 'login-auth.php';
require_once 'utils/die.php';
require_once dirname(__FILE__).'/../setup/config.php';
require_once 'models/base.php';

// 登录相关的请求全部在 https 下进行
if(($_SERVER['HTTPS'] ?? 'off') !== 'on') {
    header('HTTP/1.1 302 Unauthorized access');
    header('Location: /');
    die(0);
}

function login_html($url='') { ?>

<?php } 

if($_SERVER['REQUEST_METHOD'] === 'GET') :

$do = $_GET['do'] ?? '';
if($do === 'logout') {
    header('HTTP/1.1 302 Logged Out');
    setcookie('login','',time()-1, '/');
    // 转到登录页面以验证成功退出
    // 如果没有成功退出，那么会因认证成功而转到管理员页面
    header('Location: login.php');
    die(0);
} else {
    if(login_auth()){
        header('HTTP/1.1 302 Login Success');

        $url = $_GET['url'] ?? '/admin/';
        header('Location: '.$url);
        die(0);
    } else {
        $url = $_GET['url'] ?? '';
        login_html($url);
        die(0);
    }
}

else : // POST

if(!login_auth_passwd($_POST)) {
    login_html();
    die(0);
}

$url = $_POST['url'] ?? '/admin/';
header('HTTP/1.1 302 OK');
header('Location: '.$url);
login_auth_set_cookie();
die(0);

endif;

