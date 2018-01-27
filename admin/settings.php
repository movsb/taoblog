<?php

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once('admin.php');

admin_header();

?>
<h2>修改密码</h2>
<form id="password" method="post">
    <input type="text" name="user" placeholder="新用户名" /><br/>
    <input type="password" name="passwd_old" placeholder="原密码" /><br/>
    <input type="password" name="passwd_new" placeholder="新密码" /><br/>
    <input type="submit" value="保存" />
<script>
    var form_password = $('#password');
    form_password.submit(function() {
        $.post(
            '',
            form_password.serialize() + '&do=password',
            function(data) {
                alert(data.error);
            },
            'json'            
        );
        return false;
    });
</script>
</form>

<?php
admin_footer();

die(0);

else : // POST

function settings_die_json($arg) {
    header('HTTP/1.1 200 OK');
    header('Content-Type: application/json');

    echo json_encode($arg, JSON_UNESCAPED_UNICODE);
    die(0);
}

require_once('login-auth.php');

function auth() {
    if(!login_auth()) {
        settings_die_json([
            'errno' => 'unauthorized',
            'error' => '需要登录后才能进行该操作！',
            ]);
    }
}

require_once('load.php');

auth();

$do = $_POST['do'] ?? '';

if($do === 'password') {
    $user = $_POST['user'] ?? '';
    $passwd_old = $_POST['passwd_old'] ?? '';
    $passwd_new = $_POST['passwd_new'] ?? '';

    if(!$user || !$passwd_old || !$passwd_new) {
        settings_die_json([
            'errno' => 'fail',
            'error' => '请输入用户名/密码。',
        ]);
    }

    $_POST['passwd'] = $passwd_old;

    if(!login_auth_passwd($_POST)) {
        settings_die_json([
            'errno' => 'fail',
            'error' => '原密码不正确。',
        ]);
    }

    if($tbopt->set('login', $user.','.sha1($passwd_new))) {
        login_auth_set_cookie();
        settings_die_json([
            'errno' => 'succ',
            'error' => '修改成功。',
        ]);
    }
    else {
        settings_die_json([
            'errno' => 'succ',
            'error' => '修改失败。',
        ]);
    }
}


die(0);
endif;

