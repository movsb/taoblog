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
    
    return '<span class="value">'.($ts ? $ts : "（没有）").'</span>';
}

