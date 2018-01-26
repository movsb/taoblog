<?php

function the_meta_date() {
	global $the;

	$dd = preg_split('/-/', preg_split('/ /', $the->date)[0]);;
    $tt = sprintf('%d年%d月%d日', $dd[0], $dd[1], $dd[2]);

	$DD = preg_split('/-/', preg_split('/ /', $the->modified)[0]);;
    $TT = sprintf('%d年%d月%d日', $DD[0], $DD[1], $DD[2]);

	return '<span class="value" title="发表时间：'.$tt."\n".'修改时间：'.$TT.'">'.$tt.'</span>';
}

function the_meta_tag() {
	global $the;

	$tags = &$the->tag_names;
	$as = [];

	foreach($tags as &$t) {
		$as[] = '<a href="/tags/'.htmlspecialchars(urlencode($t)).'">'.htmlspecialchars($t).'</a>';
	}

    $ts = join(' · ', $as);
    
    return $ts ? '<span class="value">'.$ts.'</span>' : '';
}

