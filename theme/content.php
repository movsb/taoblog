<article class="post">	
	<div class="title">
		<h1><?php echo $the->title; ?></h1>
	</div>

	<div class="meta">
		<span><i class="fa fa-mr fa-user"></i>作者: 女孩不哭</span>
		<span><i class="fa fa-mr fa-calendar"></i>日期: <?php echo preg_split('/ /', $the->date)[0]; ?></span>
		<span class="category"><i class="fa fa-mr fa-folder"></i>分类: <?php 
			$taxes = $tbtax->tree_from_id($the->taxonomy);
			$links = $tbtax->link_from_slug($taxes);

			$link_anchors = [];
			foreach($taxes['name'] as $i=>$n) {
				$link_anchors[] = '<a target="_blank" href="'.$links[$i].'">'.$n.'</a>';
			}

			echo implode(',', $link_anchors);

			?></span>
	</div>

	<div class="entry">
		<?php echo $the->content; ?>
	</div><!-- end entry -->
</article>

