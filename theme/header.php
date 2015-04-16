<?php
	$home = $tbopt->get('home');
	$blog_name = $tbopt->get('blog_name');

?><!DOCTYPE html> 
<html lang="zh-CN">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<title><?php
		if(!$tbquery->count) {
			echo '404';
		} else if($tbquery->is_singular()) {
			echo $the->title;
		} else if($tbquery->is_category()) {
			$names = $tbquery->category['name'];
			$names = array_reverse($names);
			echo implode(' - ', $names);
		}

		echo ' - ',$blog_name;
	?></title>
	<link rel="stylesheet" type="text/css" href="/theme/style.css" />
	<link rel="stylesheet" type="text/css" href="/theme/font-awesome-4.3.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="/admin/scripts/jquery-2.1.3.min.js"></script>
<?php if(!$tbquery->has()) {
} else if($tbquery->is_post()) {?>
	<script type="text/javascript">var _post_id = <?php echo $the->id; ?>;</script>
	<link rel="canonical" href="<?php echo $home,"/archives/$the->id.html"; ?>" /><?php
} else if($tbquery->is_page()) {?>
	<link rel="canonical" href="<?php echo $home,"/$the->slug"; ?>" /><?php
} ?>
<?php apply_hooks('tb_head'); ?>

</head>

<body>
<div id="wrapper">
	<header id="header">
		<section id="head">
			<h1 title="Do you like it ?" onclick="location.href = location.protocol + '//' + location.host;"><?php echo $blog_name; ?><sup> ♡</sup></h1>
			<div class="description" id="today_english">
				<p id="today_english_cn">心有花种，静候春光.</p>
				<p id="today_english_en">Your heart is full of fertile seeds, waiting to sprout.</p>
			</div>
		</section>
	</header>

	<section id="main">
		<div id="content">

