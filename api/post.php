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
    check_login();

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
    // TODO 使用 tbquery 的查询功能
    $posts = $tbpost->query_by_id($id,'');

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
    check_login();

    api_die([
        "ret" => 0,
        "data" => [
            "id" => $tbpost->the_next_id(),
        ],
    ]);
}
elseif($api->method == 'get_tag_posts') {
    $tag = check_arg('tag');
    if(!strlen($tag)) {
        api_die([
            'ret' => -1,
            'msg' => '空标签',
        ]);
    }

    $posts = $tbpost->get_tag_posts($tag);
    if(!is_array($posts)) {
        api_die([
            'ret' => -1,
            'msg' => '未知错误',
        ]);
    }

    api_die([
        'ret'  => 0,
        'posts' => $posts,
    ]);
}
elseif($api->method == 'get_date_posts') {
    $yy = (int)check_arg('yy');
    $mm = (int)check_arg('mm');

    if ($yy < 1970 || ($mm < 1 || $mm > 12)){
        api_die([
            'ret' => -1,
            'msg' => '你我不在同一个世界？',
        ]);
    }

    $posts = $tbpost->get_date_posts($yy, $mm);
    if(!is_array($posts)){
        api_die([
            'ret' => -1,
            'msg' => '未知错误',
        ]);
    }

    api_die([
        'ret'  => 0,
        'posts' => $posts,
    ]);
}
elseif($api->method == 'get_cat_posts') {
    $cid = (int)check_arg('cid');
    $posts = $tbpost->get_cat_posts($cid);
    if(!is_array($posts)) {
        api_die([
            'ret' => -1,
            'msg' => '未知错误',
        ]);
    }

    api_die([
        'ret' => 0,
        'posts' => $posts,
    ]);
}
else {
    api_bad_method();
}

