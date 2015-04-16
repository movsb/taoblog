<?php

require('header.php');
?>
<div class="categories">
	<ul><?php
while($tbquery->has()){
	$the = $tbquery->the();
?>
	<li class="cat-item">
		<h2><a target="_blank" href="<?php 
			if($the->type === 'post') {
				echo $home,'/',implode('/',$tbtax->tree_from_id($the->taxonomy)['slug']),'/',$the->slug,'.html';
			} else if($the->type === 'page') {
				echo $hoem,'/',$the->slug;
			}
			?>"><?php echo $the->title;?></a></h2>
	</li>
<?php
} ?>
	</ul>
</div><?php

require('footer.php');

