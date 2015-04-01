<?php

function tb_parse_args(&$def, &$arg) {
	foreach($arg as $a => $v){
		if(!isset($def[$a])){
			unset($arg[$a]);
		}
	}

	foreach($def as $a => $v){
		if(!isset($arg[$a])){
			$arg[$a] = $v;
		}
	}

	return $arg;
}

function sanitize_uri($uri){
	return $uri;
}

function parse_query_string($q){
	$segs = explode('&', $q);
	$r = [];
	foreach($segs as $s){
		$p = explode('=', $s);
		if(count($p) && $p[0]){
			$r[$p[0]] = isset($p[1]) ? $p[1] : '';
		}
	}
	
	return $r;
}

