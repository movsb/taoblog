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
				echo the_post_link($the);
			} else if($the->type === 'page') {
				echo the_page_link($the);
			}
			?>"><?php echo $the->title;?></a></h2>
	</li>
<?php
} ?>
	</ul>
</div><?php

require('footer.php');

