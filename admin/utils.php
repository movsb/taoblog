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

function make_query_string($fields) {
    $defs = [
        'select'    => '',
        'from'      => '',
        'where'     => null,
        'groupby'   => null,
        'having'    => null,
        'orderby'   => null,
        'limit'     => null,
        'offset'    => null,
    ];

    $f = tb_parse_args($defs, $fields);

    $sql = 'SELECT ' . $f['select'] . ' FROM ' . $f['from'];

    if($f['where']) {
        if(is_string($f['where'])) {
            if(strlen($f['where'])) {
                $sql .= ' WHERE ' . $f['where'];
            }
        }
        else if(is_array($f['where'])) {
            $cond = '';

            foreach($f['where'] as $v) {
                if(is_string($v)) {
                    $cond .= ' AND (' . $v . ')';
                }
            }

            if($cond) {
                $sql .= ' WHERE 1' . $cond;
            }
        }
    }

    if($f['groupby']) {
        $sql .= ' GROUP BY ' . $f['groupby'];
    }

    if($f['having']) {
        $sql .= ' HAVING ' . $f['having'];
    }

    if($f['orderby']) {
        $sql .= ' ORDER BY ' . $f['orderby'];
    }

    if($f['limit']) {
        if($f['offset']) {
            $sql .= ' LIMIT ' . $f['limit'] . ',' . $f['offset'];
        }
        else {
            $sql .= ' LIMIT ' . $f['limit'];
        }
    }

    return $sql;
}

