<!DOCTYPE html> 
<html lang="zh-CN">
<head>
	<meta charset="UTF-8" />
	<title><?php 
		if($tbquery->count == 1) {
			echo $tbquery->objs[0]->title;
		}
	?></title>
	<link rel="stylesheet" type="text/css" href="/theme/style.css" />
	<link rel="stylesheet" type="text/css" href="/theme/font-awesome-4.3.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="/admin/scripts/jquery-2.1.3.js"></script>
	<style>
		body {
			margin-left: 250px;
			transition: margin-left 1s;
		}
	</style>
</head>

<body>
<div id="wrapper">
	<header id="header-single">

	</header>

	<section id="main">
	<div id="panel" style="box-shadow: 3px 0px 3px;background-color: purple; position: fixed; left: 0px; top: 0px; bottom: 0px; width: 250px; transition: all 1s;">
		<div class="avatar" style="width: 100px; height: 100px; position: absolute; right: 75px; top: 25px;">
			<a href="/"><img src="/favicon.ico" style="width: 100px; height: 100px; border-radius: 50px;" /></a>
		</div>
	</div>
	<div id="content-full">
		<?php 
			$the = $tbquery->the();
			require('theme/content.php');
			require('theme/comments.php');
		?>
	</div>
<?php require_once('footer.php'); 

