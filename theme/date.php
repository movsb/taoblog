<?php

require('header.php');
?>
<div class="query date-query">
	<ul class="item-list">
<?php
while($tbquery->has()){
	$the = $tbquery->the();
?>
	<li class="item cat-item"><h2><a target="_blank" href="<?php 
			echo the_link($the, false);
			?>"><?php echo $the->title;?></a><span class="thedate"><?php
				$dd = preg_split('/-/', preg_split('/ /', $the->date)[0]);

				echo '(';
				if(!$tbquery->date->yy) echo $dd[0],'年';
				if(!$tbquery->date->mm) echo $dd[1],'月';

				echo $dd[2],'日)';
			?></span></h2></li>
<?php
} ?>
	</ul>
</div>
<div class="pagination">
<?php theme_gen_pagination();

require('footer.php');

