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

