<?php
	$the = $tbquery->the();

	if(!preg_match('/[-]000/', $the->modified) && !$logged_in){
		header('Last-Modified: '.$tbdate->mysql_local_to_http_gmt($the->modified));
	}
	if($logged_in) {
		header('Cache-Control: private');
	}

require('header.php');
require('content.php');
require('footer.php');

