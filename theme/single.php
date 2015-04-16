<?php
	$the = $tbquery->the();

	if(!preg_match('/[-]000/', $the->modified)){
		header('Last-Modified: '.$tbdate->mysql_local_to_http_gmt($the->modified));
	}

	$snjs = $tbsnjs->get_snjs($the->id, $the->taxonomy);

require('header.php');
require('content.php');
require('comments.php');
require('footer.php');

