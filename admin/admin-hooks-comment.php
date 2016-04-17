<?php

defined('TBPATH') or die("Silence is golden.");

function ahc_update_count($id, $post) {
	global $tbopt;
    global $tbcmts;

    $count = $tbcmts->get_count_of_comments();
    $tbopt->set('comment_count', $count);
}

add_hook('comment_posted', 'ahc_update_count');

