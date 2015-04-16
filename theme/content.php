<article class="post">	
	<div class="title">
		<h1><?php echo $the->title; ?></h1>
	</div>

	<div class="meta">
		<span><i class="fa fa-mr fa-user"></i>女孩不哭</span>
		<span><i class="fa fa-mr fa-calendar"></i><?php echo preg_split('/ /', $the->date)[0]; ?></span>
	</div>

	<div class="entry">
		<?php echo $the->content; ?>
	</div><!-- end entry -->
</article>

