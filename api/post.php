<?php

namespace api\post;

defined('TBPATH') or die('Silence is golden.');

function check_arg($arg) {
    if(!isset($_REQUEST[$arg]) ) {
        api_die([
            "ret" => -1,
            "msg" => "expect argument `$arg'",
        ]);
    }
    else {
        return $_REQUEST[$arg];
    }
}

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

if($api->method == 'update') {
    $id         = check_arg('id');
    $content    = check_arg('content');

    $r = $tbpost->tmp_update_content((int)$id, $content) ? 0 : -1;

    api_die([
        "ret" => $r,
        "msg" => $tbpost->error,
    ]);
}
elseif($api->method == 'get') {
    $id = (int)check_arg('id');
    $posts = $tbpost->query(compact('id'));

    if($posts === false || !count($posts)) {
        api_die([
            "ret" => -1,
            "msg" => "",
        ]);
    }

    api_die([
        "ret" => 0,
        "data" => $posts[0],
    ]);
}
elseif($api->method == 'get_id') {
    api_die([
        "ret" => 0,
        "data" => [
            "id" => $tbpost->the_next_id(),
        ],
    ]);
}
else {
    api_bad_method();
}

