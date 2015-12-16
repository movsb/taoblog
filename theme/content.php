<article class="post" itemscope itemtype="http://schema.org/Article">	
	<div class="title clearfix">
		<h1 itemprop="name"><?php echo $the->title; ?></h1>
	</div>

    <?php if($logged_in) { ?>
    <div><i class="fa fa-mr fa-pencil"></i><span><?php echo the_edit_link($the);?></span></div>
    <?php } ?>

	<div class="entry clearfix" itemprop="articleBody">
		<?php echo $the->content; ?>

        <?php if($the->type == 'post') { ?>
        <div class="meta clearfix">
            <span class="item author" itemprop="author" itemscope itemtype="http://schema.org/Person"><span itemprop="name"><?php echo $tbopt->get('author'); ?></span></span>
            发表于<?php echo the_meta_date();?>，分类：<?php echo the_meta_category();?>，标签：<span itemprop="keywords"><?php echo the_meta_tag();?></span>
        </div>
        <?php } ?>
	</div><!-- end entry -->

	<div class="related clearfix">
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
    <div id="comments" class="clearfix">
        <script src="/theme/scripts/comment.js"></script>
    </div>
    <!-- comments end -->
</article>

