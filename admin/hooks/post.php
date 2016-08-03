<?php

defined('TBPATH') or die("Silence is golden.");

function ahp_last_post_time($id, $post) {
	global $tbopt;

	$last = $tbopt->get('last_post_time');
	$pdate = $post['date'];

	if(!$last || $pdate >= $last) {
		$tbopt->set('last_post_time', $pdate);
	}
}

add_hook('post_posted', 'ahp_last_post_time');

function ahp_update_post_count() {
    global $tbopt;
    global $tbpost;

    $post_count = $tbpost->get_count_of_type('post');
    $page_count = $tbpost->get_count_of_type('page');

    $tbopt->set('post_count', $post_count);
    $tbopt->set('page_count', $page_count);
}

add_hook('post_posted', 'ahp_update_post_count');


function ahp_before_query_posts($_, $sql) {
    global $logged_in;

    if($logged_in) {

    }
    else {
        $sql['where'][] = "status='public'";
    }

    return $sql;
}

add_hook('before_query_posts', 'ahp_before_query_posts');

