<?php

$tbhooks = new stdClass;

function the_hook_object($f, $p){
	$o = new stdClass;
	$o->func = $f;
	$o->priority = $p;
	return $o;
}

function add_hook($flt, $func, $pri=10){
	global $tbhooks;

	if(!isset($tbhooks->$flt))
		$tbhooks->$flt = array();
	
	$flt = &$tbhooks->$flt;

	$flt[count($flt)] = the_hook_object($func, $pri);
}

function apply_hooks($flt, ...$args) {
	global $tbhooks;
	if(!isset($tbhooks->$flt)){
		if(count($args)){
			return $args[0];
		} else {
			return false;
		}
	}

	$fs = $tbhooks->$flt;
	foreach($fs as $f){
		if(count($args)) {
			if(!is_array($args[0])) {
				$args[0] = call_user_func_array($f->func, $args);
			}
		} else {
			$fn = $f->func;
			$fn();
		}
	}

	if(count($args))
		return $args[0];
}

function &get_hooks($flt) {
	global $tbhooks;

	$tmp = [];
	if(!isset($tbhooks->$flt)){
		return $tmp;
	}

	return $tbhooks->$flt;
}

