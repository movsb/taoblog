<?php

function the_post_link(&$p, $home=true) {
	global $tbopt;
	global $tbtax;
	
	$home = $home ? $tbopt->get('home') : '';
	$cats = implode('/', $tbtax->tree_from_id($p->taxonomy)['slug']);
	$slug = $p->slug;

	return $home.'/'.$cats.'/'.$slug.'.html';
}

function the_page_link(&$p, $home=true) {
	global $tbopt;

	$home = $home ? $tbopt->get('home') : '';

	return $home.'/'.$p->slug;
}

function the_edit_post_link(&$p, $ret_anchor = true) {
	$link = '/admin/post.php?do=edit&id='.(int)$p->id;

	return $ret_anchor
		? '<a href="'.$link.'">编辑</a>'
		: $link;
}

