<?php

require('header.php');
?>
<div class="query category-query">
	<h2><?php
		$names = $tbquery->category['name'];
		echo '分类 “',htmlspecialchars(implode(',', $names)),'” 下的归档：';
	?></h2>
	<ul class="item-list">
<?php
while($tbquery->has()){
	$the = $tbquery->the();
?>
	<li class="item cat-item"><h2><a target="_blank" href="<?php 
			echo the_link($the, false);
			?>"><?php echo htmlspecialchars($the->title);?></a></h2></li>
<?php
} ?>
	</ul>
</div>

<?php

require('footer.php');

