<?php

add_hook('the_content', 'gf_content_hook');

$gf_files_dir = dirname(__FILE__).'/git-files/';

function gf_content_hook($content, $id) {
	global $gf_files_dir;
	return preg_replace_callback('/<!-- git-file -->/',
		function ($matches) use ($gf_files_dir,$id) {
			return @file_get_contents($gf_files_dir.$id.'.html');
		},
		$content
		);
}

