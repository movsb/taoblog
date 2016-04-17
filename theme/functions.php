<?php

function theme_gen_pagination() {
	global $tbquery;

	$pageno = $tbquery->pageno;
	$pagenum = $tbquery->pagenum;

	$start = max(1, $pageno-5);
	$end = min($pagenum, $pageno + 5);

	echo '<div class="pagination no-sel">',"\n";

	if($pageno > 1) echo '<a href="'.($pageno-1).'" class="page-number">上一页</a>';

		for($i=$start; $i < $pageno; $i++) {
			echo '<a pageno="'.$i.'" class="page-number">'.$i.'</a>';
		}
		echo '<span class="current">'.$pageno.'</span>';
		for($i=$pageno+1; $i <= $end; $i++) {
			echo '<a pageno="'.$i.'" class="page-number">'.$i.'</a>';
		}

		if($pageno < $pagenum) echo '<a pageno="'.($pageno+1).'" class="page-number">下一页</a>';
?>
		<script type="text/javascript">
			$('.pagination').click(function(e) {
				var cl = e.target.classList;
				if(cl.contains('page-number')) {
					var pageno = $(e.target).attr('pageno');
					var loc = location.pathname;
					if(!/page\/\d+$/.test(loc)) loc += 'page/1';
					loc = loc.replace(/page\/\d+$/,'page/'+pageno);
					location.pathname = loc;
				}
			});
		</script>
	</div>
<?php
}

function the_meta_category() {
	global $tbtax;
	global $the;

	$taxes = $tbtax->tree_from_id($the->taxonomy);
	$links = $tbtax->link_from_slug($taxes);

	$link_anchors = [];
	foreach($taxes['name'] as $i=>$n) {
		$link_anchors[] = '<a target="_blank" href="'.$links[$i].'">'.$n.'</a>';
	}

	return '<span class="value">'.implode(', ', $link_anchors).'</span>';
}

function the_meta_date() {
	global $the;

	$dd = preg_split('/-/', preg_split('/ /', $the->date)[0]);;

	$link  = '<a target="_blank" href="/date/'.$dd[0].'/">'.$dd[0].'</a>年';
	$link .= '<a target="_blank" href="/date/'.$dd[0].'/'.$dd[1].'/">'.$dd[1].'</a>月';
	$link .= $dd[2].'日';

	return '<span class="value">'.$link.'</span>';
}

function the_meta_tag() {
	global $the;

	$tags = &$the->tag_names;
	$as = [];

	foreach($tags as &$t) {
		$as[] = '<a target="_blank" href="/tags/'.urlencode($t).'">'.$t.'</a>';
	}

    $ts = join(', ', $as);
    
    return '<span class="value">'.($ts ? $ts : "（没有）").'</span>';
}

