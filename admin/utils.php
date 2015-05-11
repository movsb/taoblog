<?php

function &tb_parse_args(&$def, &$arg) {
	foreach($def as $a => $v){
		if(!isset($arg[$a])){
			$arg[$a] = $v;
		}
	}

	return $arg;
}

function &parse_query_string($q, $dk=true, $dv=true){
	$segs = explode('&', $q);
	$r = [];
	foreach($segs as $s){
		$p = explode('=', $s);
		if(count($p) && $p[0]){
			$k = $dk ? urldecode($p[0]) : $p[0];
			$v = isset($p[1]) ? ($dv ? urldecode($p[1]) : $p[1]) : '';
			$r[$k] = $v;
		}
	}
	
	return $r;
}

