<article class="post">	
	<div class="title">
		<h1><?php echo $the->title; ?></h1>
	</div>

	<div class="meta no-sel">
		<?php if($the->type == 'post') { ?>
		<span class="item tag"><i class="fa fa-mr fa-tag"></i><span class="label">标签: </span><?php echo the_meta_tag(); ?></span>
		<span class="item author"><i class="fa fa-mr fa-user"></i><span class="label">作者: </span><?php echo $tbopt->get('author'); ?></span>
		<span class="item date"><i class="fa fa-mr fa-calendar"></i><span class="label">日期: </span><?php echo the_meta_date(); ?></span>
		<span class="item category"><i class="fa fa-mr fa-folder"></i><span class="label">分类: </span><?php echo the_meta_category(); ?></span>
		<span class="item font-sizing"><i class="fa fa-mr fa-font"></i><span><span class="label">字号: </span></span><?php
			?><span class="dec"><i class="fa fa-minus"></i></span><?php
			?><span class="inc"><i class="fa fa-plus"></i></span></span>
		<?php } ?>
		<?php if($logged_in) { ?>
		<span class="item edit-post"><i class="fa fa-mr fa-pencil"></i><span><?php echo the_edit_link($the);?></span></span>
		<?php } ?>
	</div>

	<div class="entry">
		<?php echo $the->content; ?>
	</div><!-- end entry -->

	<div class="related">
		<?php if($the->type == 'post' && $tbquery->related_posts && count($tbquery->related_posts)) {
			echo '<h3><i class="fa fa-mr fa-book"></i>相关文章</h3>', PHP_EOL;
			echo '<ol>',PHP_EOL;

			$ps = &$tbquery->related_posts;
			foreach($ps as $p) {
				echo '<li><a href="/',$p->id,'/">', $p->title, '</a></li>', PHP_EOL;
			}

			echo '</ol>',PHP_EOL;
		} ?>
	</div><!-- end related -->
</article>

