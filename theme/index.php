<?php
	$home = $tbopt->get('home');
	$blog_name = $tbopt->get('blog_name');

?><!DOCTYPE html> 
<html lang="zh-CN">
<head>
	<meta charset="UTF-8" />
	<title><?php echo $tbopt->get('blog_name');?></title>
	<meta name="description" content="<?php echo $tbopt->get('blog_desc'); ?>" />
	<link rel="alternate" type="application/rss+xml" title="<?php echo htmlspecialchars($blog_name);?>" href="<?php echo $home,'/rss';?>" />
	<link rel="stylesheet" type="text/css" href="/theme/index.css" />
	<link rel="stylesheet" type="text/css" href="/theme/font-awesome-4.3.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="/admin/scripts/jquery-2.1.3.min.js"></script>
</head>

<body>
<div id="wrapper">
	<div style="margin: 2em;">
		<span style="font-size: 3em;"><?php echo $tbopt->get('blog_name'); ?></span>
	</div>
	<div class="recent-posts">
		<h3 style="display: inline-block; border-bottom: 1px solid; padding: 0px 2em;">近期文章</h3>
		<ul style="padding: 0px; margin: 0px; list-style: none;">
		<?php
			$q = ['pageno' => 1];
			$posts = $tbpost->query($q);
			if(is_array($posts) && count($posts)) {
				foreach($posts as $p) {
					$link = the_post_link($p);
					echo '<li><a href="'.$link.'">',$p->title,'</a></li>';
				}
			}
		?>
		</ul>
	</div>
</div>
</body>
</html>

