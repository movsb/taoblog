<article class="post">	
	<div class="title">
		<h1><?php echo $the->title; ?></h1>
	</div>

	<div class="meta">
		<span class="item"><i class="fa fa-mr fa-user"></i>作者: <?php echo $tbopt->get('nickname'); ?></span>
		<span class="item"><i class="fa fa-mr fa-calendar"></i>日期: <?php echo preg_split('/ /', $the->date)[0]; ?></span>
		<span class="item category"><i class="fa fa-mr fa-folder"></i>分类: <?php 
			$taxes = $tbtax->tree_from_id($the->taxonomy);
			$links = $tbtax->link_from_slug($taxes);

			$link_anchors = [];
			foreach($taxes['name'] as $i=>$n) {
				$link_anchors[] = '<a target="_blank" href="'.$links[$i].'">'.$n.'</a>';
			}

			echo implode(',', $link_anchors);

			?></span>
		<span class="item font-sizing">
			<i class="fa fa-mr fa-font"></i><span>字号: </span><?php
			?><span class="dec"><i class="fa fa-minus"></i></span><?php
			?><span class="inc"><i class="fa fa-plus"></i></span>
		</span>
		<?php if($logged_in) { ?>
		<span class="item edit-post">
			<i class="fa fa-mr fa-pencil"></i><span><?php echo the_edit_post_link($the);?></span>
		</span>
		<?php } ?>
		<script>
			$('.font-sizing').click(function(e){
				var post = $('.entry');
				var cl = e.target.classList;
				if(cl.contains('inc') || cl.contains('fa-plus')) {
					post.css('font-size', parseFloat(post.css('font-size')) * 1.2 + 'px');
				} else if(cl.contains('dec') || cl.contains('fa-minus')) {
					post.css('font-size', Math.max(8, parseFloat(post.css('font-size')) / 1.2) + 'px');
				}
			});
		</script>
	</div>

	<div class="entry">
		<?php echo $the->content; ?>
	</div><!-- end entry -->
</article>

