<?php

namespace api\post;

defined('TBPATH') or die('Silence is golden.');

function check_existence() {
    global $tbapi;
    global $tbshuoshuo;

    $id = (int)$_REQUEST['id'];
    if($id == 0 || !$tbshuoshuo->has($id)) {
        $tbapi->err(-1, "doesn't exist");
    }
} 

if($tbapi->method == 'update') {
    $tbapi->auth();

    $id         = (int)$tbapi->expected('id');
    $content    = (string)$tbapi->expected('content');

    $r = $tbpost->tmp_update_content($id, $content) ? 0 : -1;

    $tbapi->err($r, $tbpost->error);
}
elseif($tbapi->method == 'get') {
    $id = (int)$tbapi->expected('id');
    // TODO 使用 tbquery 的查询功能
    $posts = $tbpost->query_by_id($id,'');

    if($posts === false || !count($posts)) {
        $tbapi->err(-1,"");
    }

    $tbapi->done($posts[0]);
}
elseif($tbapi->method == 'get_id') {
    $tbapi->auth();

    $tbapi->done(["id"=>$tbpost->the_next_id()]);
}
elseif($tbapi->method == 'get_tag_posts') {
    $tag = $tbapi->expected('tag');
    if(!strlen($tag)) {
        $tbapi->err(-1,"");
    }

    $posts = $tbpost->get_tag_posts($tag);
    if(!is_array($posts)) {
        $tbapi->err(-1,"");
    }

    $tbapi->done($posts);
}
elseif($tbapi->method == 'get_date_posts') {
    $yy = (int)$tbapi->expected('yy');
    $mm = (int)$tbapi->expected('mm');

    if ($yy < 1970 || ($mm < 1 || $mm > 12)){
        $tbapi->err(-1, "你我不在同一个世界？");
    }

    $posts = $tbpost->get_date_posts($yy, $mm);
    if(!is_array($posts)){
        $tbapi->err(-1,"");
    }

    $tbapi->done($posts);
}
elseif($tbapi->method == 'get_cat_posts') {
    $cid = (int)$tbapi->expected('cid');
    $posts = $tbpost->get_cat_posts($cid);
    if(!is_array($posts)) {
        $tbapi->err(-1,"");
    }

    $tbapi->done($posts);
}
else {
    $tbapi->bad();
}

