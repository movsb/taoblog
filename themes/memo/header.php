<?php
	$blog_name = $tbopt->get('blog_name');

?><!DOCTYPE html> 
<html lang="zh-CN">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1, maximum-scale=1" />
	<title><?php
		if($tbquery->is_home()) {

        } else if($tbquery->is_archive()) {
            echo '文章归档';
		} else if($tbquery->is_singular()) {
			echo htmlspecialchars($the->title);
		} else if($tbquery->is_category()) {
			$names = $tbquery->category['name'];
			$names = array_reverse($names);
			echo htmlspecialchars(implode(' - ', $names));
		} else if($tbquery->is_date()) {
			if($tbquery->date->yy >= 1970) echo $tbquery->date->yy,'年';
			if($tbquery->date->mm >= 1 && $tbquery->date->mm <= 12) echo $tbquery->date->mm,'月';
		} else if($tbquery->is_tag()) {
			echo htmlspecialchars($tbquery->tags);
        } else if($tbquery->is_memory()) {
            echo '说说';
		} else if(!$tbquery->count) {
			echo '404';
        }

        if(!$tbquery->is_home()) echo ' - ';
		echo htmlspecialchars($blog_name);
	?></title>
	<?php if($tbquery->is_home()) {
		echo '<meta name="keywords" content="', htmlspecialchars($tbopt->get('keywords')), '" />', PHP_EOL;
} ?>
	<link rel="alternate" type="application/rss+xml" title="<?php echo htmlspecialchars($blog_name);?>" href="<?php echo '/rss';?>" />
	<link rel="stylesheet" type="text/css" href="/themes/memo/style.css" />
    <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.5.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="//blog-10005538.file.myqcloud.com/jquery.min.js"></script>
<?php if(!$tbquery->count) {

} else if($tbquery->is_singular()) {?>
    <link rel="canonical" href="<?php echo the_link($the);?>" />
    <base href="<?php echo the_id_link($the);?>" />
    <script type="text/javascript">var _post_id = <?php echo $the->id; ?>;</script>
<?php } 

	apply_hooks('tb_head'); ?>

</head>

<body>
<div id="wrapper">
    <!-- 头部 -->
	<header id="header">
        <div class="content">
            <div class="padding">
                <h2 class="sitename"><a href="/"><?php echo htmlspecialchars($blog_name); ?></a></h2>
                <p class="motto">不忘初心，方得始终</p>
                <div class="nav">
                    <ol>
                        <li><a href="/admin/">管理后台</a></li>
                    </ol>
                </div>
            </div>
        </div>

	</header>

	<section id="main">
		<div id="content">

