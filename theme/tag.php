<?php

require('header.php');
?>
<div class="query tag-query">
	<h2><?php
		echo '标签 `',htmlspecialchars($tbquery->tags),'` 下的归档（第',$tbquery->pageno,'页）：';
	?></h2>
	<ul class="item-list">
<?php
while($tbquery->has()){
	$the = $tbquery->the();
?>
	<li class="item cat-item"><h2><a target="_blank" href="<?php 
			echo the_link($the, false);
			?>"><?php echo htmlspecialchars($the->title);?></a><span class="thedate"><?php
				$dd = preg_split('/-/', preg_split('/ /', $the->date)[0]);
				echo '(', $dd[0],'年', $dd[1],'月', $dd[2],'日)';
			?></span></h2></li>
<?php
} ?>
	</ul>
</div>

<?php theme_gen_pagination();

require('footer.php');

