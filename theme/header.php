<?php
	$home = $tbopt->get('home');
	$blog_name = $tbopt->get('blog_name');

?><!DOCTYPE html> 
<html lang="zh-CN">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<title><?php
		if($tbquery->is_home()) {
			echo '首页';
		} else if(!$tbquery->count) {
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
		} else if($tbquery->is_tag()) {
			echo "第{$tbquery->pageno}页 - ";
			echo $tbquery->tags;
		}

		echo ' - ',$blog_name;
	?></title>
	<?php if($tbquery->is_home()) {
		echo '<meta name="keywords" content="', $tbopt->get('keywords'), '" />', PHP_EOL;
} ?>
	<link rel="alternate" type="application/rss+xml" title="<?php echo htmlspecialchars($blog_name);?>" href="<?php echo $home,'/rss';?>" />
	<link rel="stylesheet" type="text/css" href="/theme/style.css" />
	<link rel="stylesheet" type="text/css" href="/theme/font-awesome-4.3.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="/admin/scripts/jquery-2.1.3.min.js"></script>
<?php if(!$tbquery->count) {

} else if($tbquery->is_singular()) {?>
	<link rel="canonical" href="<?php echo the_link($the);?>" />
	<script type="text/javascript">var _post_id = <?php echo $the->id; ?>;</script>
<?php } 

	apply_hooks('tb_head'); ?>

</head>

<body>
<div id="wrapper">
	<header id="header">
		<div id="header-wrapper">
			<div class="desktop">
				<section id="head">
					<div class="center">
						<img class="me home-a" src="/theme/images/me.png" />
						<h1 class="home-a"><?php echo $blog_name; ?></h1>
						<script>
							$('.home-a').click(function() {
								location.href = location.protocol + '//' + location.host;
							});
						</script>
						<div class="social" style="font-size: 2em;">
							<span><a title="RSS" target="_blank" href="/rss"><i class="fa fa-rss"></i></a></span>
							<span><a title="GitHub" target="_blank" href="https://github.com/movsb"><i class="fa fa-github"></i></a></span>
							<span><a title="QQ" target="_blank" href="http://sighttp.qq.com/authd?IDKEY=b19745b9da616a000d2db5731672dd06b575204bf1bbf9c2"><i class="fa fa-qq" style="font-size: 0.8em; position: relative; top: -1px;"></i></a></span>
						</div>
					</div>
				</section>
				<div class="footer center" id="footer">
					<div class="footer-wrapper">
						<div class="column about">
							<h3>ABOUT</h3>
							<ul>
								<li><a href="/blog">关于博客</a></li>
								<li><a href="/echo">建议反馈</a></li>
								<li><a href="/about">关于我</a></li>
								<li><a target="_blank" href="<?php echo $tbopt->get('home'),'/rss'; ?>">订阅博客</a></li>
							</ul>
						</div>
						<div class="column links">
							<h3>LINKS</h3>
							<ul>
								<li><a title="自学去 - 一个免费的自学网站~" target="_blank" href="http://www.zixue7.com">自学去</a></li>
								<li><a title="小谢的博客" target="_blank" href="http://memorycat.com">小写adc</a></li>
								<li><a title="网事如风的博客" target="_blank" href="http://godebug.org">网事如风</a></li>
							</ul>
						</div>
					</div>
				</div><!-- footer -->
				<div class="copy center">
					<div class="copyright mb">
						<span>&copy; <?php echo date('Y'),' ',$tbopt->get('author'); ?></span>
					</div>
				</div>
			</div>
			<div class="mobile">
				<img class="me home-a" src="/theme/images/me.png" />
				<h1 class="home-a"><?php echo $blog_name; ?></h1>
				<div class="social" style="font-size: 2em;">
					<span><a title="RSS" target="_blank" href="/rss"><i class="fa fa-rss"></i></a></span>
					<span><a title="GitHub" target="_blank" href="https://github.com/movsb"><i class="fa fa-github"></i></a></span>
					<span><a title="QQ" target="_blank" href="http://sighttp.qq.com/authd?IDKEY=b19745b9da616a000d2db5731672dd06b575204bf1bbf9c2"><i class="fa fa-qq" style="font-size: 0.78em; position: relative; top: -2px;"></i></a></span>
				</div>
				<script>
					$('.home-a').click(function() {
						location.href = location.protocol + '//' + location.host;
					});
				</script>
			</div>
		</div>
	</header>

	<section id="main">
		<div id="content">

