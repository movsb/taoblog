<?php

namespace api\post;

defined('TBPATH') or die('Silence is golden.');

if ($tbapi->method == 'get_tag_posts') {
    $tag = $tbapi->expected('tag');
    if (!strlen($tag)) {
        $tbapi->err(-1, "");
    }

    $posts = $tbpost->query_by_tags($tag, false);
    if (!is_array($posts)) {
        $tbapi->err(-1, "");
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
} else {
    $tbapi->bad();
}

