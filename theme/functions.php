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

	return implode(',', $link_anchors);
}

function the_meta_date() {
	global $the;

	$dd = preg_split('/-/', preg_split('/ /', $the->date)[0]);;

	$link  = '<a target="_blank" href="/date/'.$dd[0].'/">'.$dd[0].'</a>-';
	$link .= '<a target="_blank" href="/date/'.$dd[0].'/'.$dd[1].'/">'.$dd[1].'</a>-';
	$link .= $dd[2];

	return $link;
}

function today_english() {
	$today = [];

	$ch = curl_init("http://xue.youdao.com/w?method=tinyEngData&date=" . date("Y-m-d"));
	curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
	$html = curl_exec($ch);
	curl_close($ch);

	$doc = new DOMDocument();
	$libxml_previous_state = libxml_use_internal_errors(true);
	$doc->loadHTML('<?xml version="1.0" encoding="UTF-8" ?>' . $html);
	libxml_clear_errors();
	libxml_use_internal_errors($libxml_previous_state);

	$have = false;

	$divlist = $doc->getElementsByTagName('div');
	if($divlist && $divlist->length){
		$div = $divlist->item(0);
		if($div->hasAttributes()){
			$example_english = $div->attributes->getNamedItem('class');
			if($example_english && $example_english->nodeValue == 'example english'){
				$nodes = $div->childNodes;
				$image = $nodes->item(1)->childNodes->item(0)->attributes->getNamedItem('src')->nodeValue;
				$sentence = $nodes->item(3)->childNodes->item(1)->childNodes->item(0)->nodeValue;
				$translate= $nodes->item(3)->childNodes->item(3)->childNodes->item(0)->wholeText;
				$today = compact('sentence', 'translate');
				$have = true;
			}
		}
	}

	if($have === false){
		$today = [
			'sentence' => 'Praising what is lost makes the remembrance dear.',
			'translate' => '缅怀已经失去的，将使回忆变得亲切。'
		];
	}
?>
<div class="today-english">
	<h2>今日英语</h2>
	<div style="text-indent: 2em;">
		<p><?php echo $today['sentence'];?></p>
		<p><?php echo $today['translate'];?></p>
	</div>
</div>
<?php
}
