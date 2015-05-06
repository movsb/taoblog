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
			echo "第{$tbquery->pageno}页 - ";
			$names = $tbquery->category['name'];
			$names = array_reverse($names);
			echo implode(' - ', $names);
		} else if($tbquery->is_date()) {
			echo "第{$tbquery->pageno}页 - ";
			if($tbquery->date->yy >= 1970) echo $tbquery->date->yy,'年';
			if($tbquery->date->mm >= 1 && $tbquery->date->mm <= 12) echo $tbquery->date->mm,'月';
		}

		echo ' - ',$blog_name;
	?></title>
	<link rel="alternate" type="application/rss+xml" title="<?php echo htmlspecialchars($blog_name);?>" href="<?php echo $home,'/rss';?>" />
	<link rel="stylesheet" type="text/css" href="/theme/style.css" />
	<link rel="stylesheet" type="text/css" href="/theme/font-awesome-4.3.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="/admin/scripts/jquery-2.1.3.min.js"></script>
<?php if(!$tbquery->count) {

} else if($tbquery->is_singular()) {?>
	<script type="text/javascript">var _post_id = <?php echo $the->id; ?>;</script>
<?php } 

	apply_hooks('tb_head'); ?>

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
			<div class="social" style="font-size: 2em;">
				<span><a title="RSS" target="_blank" href="/rss"><i class="fa fa-rss"></i></a></span>
				<span><a title="GitHub" target="_blank" href="https://github.com/movsb"><i class="fa fa-github"></i></a></span>
				<span><a title="QQ" target="_blank" href="http://sighttp.qq.com/authd?IDKEY=b19745b9da616a000d2db5731672dd06b575204bf1bbf9c2"><i class="fa fa-qq" style="font-size: 0.8em;"></i></a></span>
			</div>
		</section>
	</header>

	<section id="main">
		<div id="content">

