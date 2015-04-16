<?php
	$home = $tbopt->get('home');
	$the = $tbquery->the();

	if(!preg_match('/[-]000/', $the->modified)){
		header('Last-Modified: '.$tbdate->mysql_local_to_http_gmt($the->modified));
	}
?><!DOCTYPE html> 
<html lang="zh-CN">
<head>
	<meta charset="UTF-8" />
	<title><?php echo $the->title,' - ',$tbopt->get('blog_name');?></title>
	<link rel="stylesheet" type="text/css" href="/theme/style.css" />
	<link rel="stylesheet" type="text/css" href="/theme/font-awesome-4.3.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="/admin/scripts/jquery-2.1.3.min.js"></script>
	<script type="text/javascript">var _post_id = <?php echo $the->id; ?>;</script>
	<link rel="canonical" href="<?php echo $home,"/archives/$the->id.html"; ?>" />
<?php apply_hooks('tb_head'); ?>
</head>

<body>
<div id="wrapper">
	<header id="header">
		<section id="head">
			<h1 title="Do you like it ?" onclick="location.href = location.protocol + '//' + location.host;">Metal Max<sup> ♡</sup></h1>
			<div class="description" id="today_english">
				<p id="today_english_cn">心有花种，静候春光.</p>
				<p id="today_english_en">Your heart is full of fertile seeds, waiting to sprout.</p>
			</div>
		</section>
	</header>

	<section id="main">
		<div id="content">
			<?php 
				require('theme/content.php');
				require('theme/comments.php');
			?>
		</div>
	</section>
</div><!-- wrapper -->
<?php apply_hooks('tb_footer'); ?>
</body>
</html>

