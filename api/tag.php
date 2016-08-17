<?php

namespace taoblog\api\tag;

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
    global $tbtag;
    global $tbpost;

    $pid = (int)$_REQUEST['pid'];
    if($pid == 0 || !$tbpost->have($pid)) {
        api_die([
            "ret" => -1,
            "msg" => "no such post",
        ]);
    }
} 

if($api->method == 'update') {
    $pid    = (int)check_arg('pid');
    $tags   = (string)check_arg('tags');

    $r = $tbtag->update_post_tags($pid, $tags) ? 0 : -1;

    api_die([
        "ret" => $r,
        "msg" => $tbtag->error,
    ]);
}
else if($api->method == 'get') {
    $pid = (int)check_arg('pid');
    check_existence();

    $r = $tbtag->get_post_tag_names($pid);

    api_die([
        "ret" => 0,
        "data" => implode(',', $r),
    ]);
}
else {
    api_bad_method();
}

