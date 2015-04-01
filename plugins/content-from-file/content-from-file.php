<?php

add_hook('the_content', 'cff_content_hook');

$cff_blog_dir = '/home/twofei/Desktop/blogcontent/';

function cff_content_cb ($matches) {
	global $cff_blog_dir;
	return file_get_contents($cff_blog_dir.$matches[1]);
}

function cff_content_hook($content) {
	return preg_replace_callback(
		'#\[file\](.*)\[\/file\]#U', 
		'cff_content_cb',
		$content
		);
}

