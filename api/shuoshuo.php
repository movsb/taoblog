<?php

namespace api\shuoshuo;

defined('TBPATH') or die('Silence is golden.');

function check_arg($arg) {
    if(!isset($_REQUEST[$arg]) ) {
        api_die([
            "ret" => -1,
            "msg" => "expect argument `$arg'",
        ]);
    }
    else { return $_REQUEST[$arg]; } }

function check_existence() {
    global $tbshuoshuo;

    $id = (int)$_REQUEST['id'];
    if($id == 0 || !$tbshuoshuo->has($id)) {
        api_die([
            "ret" => -1,
            "msg" => "doesn't exist.",
        ]);
    }
}

if($api->method == 'get') {
    check_arg('id');
    check_existence();

    $id = (int)$_REQUEST['id'];
    $ss = $tbshuoshuo->get($id);
    api_die([
        "ret" => 0,
        "data" => $ss,
    ]);
}
elseif($api->method == 'del') {
    check_arg('id');
    check_existence();

    $r = $tbshuoshuo->del($_REQUEST['id']) ? 0 : -1;

    api_die([
        "ret" => $r,
    ]);
}
elseif($api->method == 'get-latest') {
    $count = 10;
    if(isset($_REQUEST['n'])) {
        $count = (int)$_REQUEST['n'];
    }

    if($count < 0) {
        api_die([
            "ret" => -1,
            "msg" => "bad arguments.",
        ]);
    }

    $ss = $tbshuoshuo->get_latest($count);

    api_die([
        "ret" => 0,
        "data" => $ss,
    ]);
}
elseif($api->method == 'update') {
    $id = (int)check_arg('id');
    $content = check_arg('content');

    check_existence();

    $r = $tbshuoshuo->update($id, $content) ? 0 : -1;

    api_die([
        "ret" => $r,
    ]);
}
elseif($api->method == 'post') {
    $content = check_arg('content');
    $date = $tbdate->mysql_datetime_gmt();

    $r = $tbshuoshuo->post($content);

    if($r === false) {
        api_die([
            "ret" => -1,
        ]);
    }
    else {
        api_die([
            "ret" => 0,
            "data" => [
                "id" => $r,
            ],
        ]);
    }
}
else{
    api_bad_method();
}

