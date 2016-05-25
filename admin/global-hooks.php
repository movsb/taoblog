<?php

namespace global_hooks; // 有史以来第1次写 namespace

defined('TBPATH') or die("Silence is golden.");

function before_query_posts($_, $sql) {
    global $logged_in;

    if($logged_in) {

    }
    else {
        $sql['where'][] = "status='public'";
    }

    return $sql;
}

add_hook('before_query_posts', 'global_hooks\before_query_posts');

