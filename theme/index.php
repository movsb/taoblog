<!DOCTYPE html> 
<html lang="zh-CN">
<head>
	<meta charset="UTF-8" />
	<title><?php echo $tbopt->get('blog_name');?></title>
	<meta name="description" content="<?php echo $tbopt->get('blog_desc'); ?>" />
	<link rel="stylesheet" type="text/css" href="/theme/style.css" />
	<link rel="stylesheet" type="text/css" href="/theme/font-awesome-4.3.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="/admin/scripts/jquery-2.1.3.min.js"></script>
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
		while($tbquery->have()){
			$the = $tbquery->the();
			require('content.php');
		}
	?>

	<div class="pagination">
		<?php
			$post_count = $tbpost->get_count();
			$pageno = $tbquery->pageno;
			$pagenum = (int)ceil($post_count / $tbquery->posts_per_page);

			$start = max(1, $pageno-5);
			$end = min($pagenum, $pageno + 5);

			if($pageno > 1) echo '<a href="/page/'.($pageno-1).'" class="page-number">上一页</a>';

			for($i=$start; $i < $pageno; $i++) {
				echo '<a href="/page/'.$i.'" class="page-number">'.$i.'</a>';
			}
			echo '<span class="current">'.$pageno.'</span>';
			for($i=$pageno+1; $i <= $end; $i++) {
				echo '<a href="/page/'.$i.'" class="page-number">'.$i.'</a>';
			}

			if($pageno < $pagenum) echo '<a href="/page/'.($pageno+1).'" class="page-number">下一页</a>';
		?>
	</div>
	</div>

<?php require_once('footer.php'); 

