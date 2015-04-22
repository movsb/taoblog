<?php

require('header.php');
?>
<div class="date-query">
	<ul><?php
while($tbquery->has()){
	$the = $tbquery->the();
?>
	<li class="cat-item">
		<h2><a target="_blank" href="<?php 
			if($the->type === 'post') {
				echo the_post_link($the);
			} else if($the->type === 'page') {
				echo the_page_link($the);
			}
			?>"><?php echo $the->title;?></a></h2>
	</li>
<?php
} ?>
	</ul>
</div>
<div class="pagination">
<?php
$post_count = $tbpost->get_count();
$pageno = $tbquery->pageno;
$pagenum = (int)ceil($post_count / $tbquery->posts_per_page);

$start = max(1, $pageno-5);
$end = min($pagenum, $pageno + 5);

if($pageno > 1) echo '<a href="/page/'.($pageno-1).'" class="page-number">上一页</a>';

	for($i=$start; $i < $pageno; $i++) {
		echo '<a href="/page/'.$i.'" class="page-number">'.$i.'</a>';
	}
	echo '<span class="current">'.$pageno.'</span>';
	for($i=$pageno+1; $i <= $end; $i++) {
		echo '<a href="/page/'.$i.'" class="page-number">'.$i.'</a>';
	}

	if($pageno < $pagenum) echo '<a href="/page/'.($pageno+1).'" class="page-number">下一页</a>';
?>
</div>
<?php

require('footer.php');

