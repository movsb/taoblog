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
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <title>登录 - TaoBlog</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1, maximum-scale=1" />
    <style>
body {
    padding: 0px;
    margin: 0px;
}

#wrapper {
    position: fixed;
    left: 0;
    top: 0;
    right: 0;
    bottom: 0;
}

#login-form {
    position: absolute;
    left: 0;
    top: 50%;
    right: 0;
    width: 220px;
    margin: auto;
    border: 1px solid gray;
    transform: translateY(-50%);
}

#title {
    font-size: 20px;
    text-align: center;
    height: 3em;
    line-height: 3em;
}

#input-wrapper {
    padding: 0 1.5em 1em 1.5em;
}

.input {
    width: 100%;
    box-sizing: border-box;
    padding: 0.5em;
    line-height: 1em;
    margin-bottom: 0.5em;
    border: 1px solid gray;
}

.btn {
    line-height: 1.8em;
    font-size: 1em;   
}

    </style>
</head>
<body>
<div id="wrapper">
    <form method="post" id="login-form">
        <div id="title">登录</div>
        <div id="input-wrapper">
            <div>
                <input class="input" type="text" name="user" placeholder="用户名" />
                <input class="input" type="password" name="passwd" placeholder="密码" />
            </div>	
            <div class="submit" style="text-align: right;">
                <input class="btn" type="submit" value="登录" />
                <input class="btn" type="button" value="取消" onclick="location.href='/';" />
            </div>
        </div>
        <div class="hidden">
        <?php if($url) { ?>
            <input type="hidden" name="url" value="<?php echo htmlspecialchars($url); ?>" />
        <?php } ?>
        </div>
    </form>
</div>
</body>
</html>

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

