<article class="post" itemscope itemtype="http://schema.org/Article">	
	<div class="title">
		<h1 itemprop="name"><?php echo $the->title; ?></h1>
	</div>

	<div class="meta no-sel">
		<?php if($the->type == 'post') { ?>
		<span class="item tag"><i class="fa fa-mr fa-tag"></i><span class="label">标签: </span><span itemprop="keywords"><?php echo the_meta_tag(); ?></span></span>
		<span class="item author" itemprop="author" itemscope itemtype="http://schema.org/Person"><i class="fa fa-mr fa-user"></i><span class="label">作者: </span><span itemprop="name"><?php echo $tbopt->get('author'); ?></span></span>
		<span class="item date"><i class="fa fa-mr fa-calendar"></i><span class="label">日期: </span><?php echo the_meta_date(); ?></span>
		<span class="item category"><i class="fa fa-mr fa-folder"></i><span class="label">分类: </span><?php echo the_meta_category(); ?></span>
		<?php } ?>
		<?php if($logged_in) { ?>
		<span class="item edit-post"><i class="fa fa-mr fa-pencil"></i><span><?php echo the_edit_link($the);?></span></span>
		<?php } ?>
	</div>

	<div class="entry" itemprop="articleBody">
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

    <!-- comments begin -->
    <div id="comments">
        <script src="/theme/scripts/comment.js"></script>
    </div>
    <!-- comments end -->
</article>

