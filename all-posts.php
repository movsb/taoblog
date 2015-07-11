<?php

require('admin/load.php');

function get_all_posts() {
	global $tbdb;

	$sql = "SELECT id,date,title,content FROM posts WHERE type='post' AND status='public' ORDER BY date DESC";

	$results = $tbdb->query($sql);

	if($results) {
		while($row = $results->fetch_object()) {
			echo '<li><a target="_blank" href="/',$row->id,'/">',$row->title,'</a></li>',"\n";
		}
	}
}

?><!doctype html>
<html>
<head>
<meta charset="UTF-8" />
<title>所有文章 - <?php echo $tbopt->get('blog_name'); ?></title>
</head>
<body>
<h1>所有文章：</h1>
<div style="padding: 0px 2em;">
<ul>
<?php get_all_posts(); ?>
</ul>
</div>
</body>
</html>
<?php


