<?php

require_once(dirname(__FILE__).'/../admin/load.php');

$api = new stdClass();
$api->loggedin = $logged_in;

function api_die($arr) {
    header('HTTP/1.1 200 OK');
    header('Content-Type: application/json');
    echo json_encode($arr);
    die(0);
}

function api_bad_method() {
    global $api;

    api_die([
        'ret' => -1,
        'msg' => "unknown method `{$api->method}' of module `{$api->module}'.",
    ]);
}

function check_login() {
    global $api;

    if($api->loggedin) {
        return;
    }

    if($api->module != 'login' || $api->method != 'auth') {
        api_die([
            "ret" => -1,
            "msg" => "login please.",
        ]);
    }

    $user = $_REQUEST['user'] ?? '';
    $passwd = $_REQUEST['passwd'] ?? '';

    $arg = compact('user', 'passwd');

    $ok = login_auth_passwd($arg);

    if($ok) {
        api_die([
            "ret" => 0,
            "data" => [
                "login" => login_gen_cookie(),
            ],
        ]);
    }
    else {
        api_die([
            "ret" => -1,
            "msg" => "login auth failed.",
        ]);
    }
}

if(preg_match('~^/api/([^/]+)/([^/?]+)~', $_SERVER['REQUEST_URI'], $matches)) {
    $api->module = $matches[1];
    $api->method = $matches[2];

    check_login();

    @include_once($api->module.'.php');

    die(0);
}
else {
    api_die([
        "ret" => -1,
        "msg" => "bad request.",
    ]);
}

