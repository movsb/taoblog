<?php

/*
 * 该程序用来获取所有公开的文章，按日期排序，独立使用
 * 临时使用，不久的将来（一定）会删除掉
 *
 */

require('admin/load.php');

function get_all_posts() {
	global $tbdb;

	$sql = "SELECT id,date,title,content FROM posts WHERE type='post' AND status='public' ORDER BY date DESC";

	$results = $tbdb->query($sql);

	if($results) {
		while($row = $results->fetch_object()) {
			echo '<li><a target="_blank" href="/',$row->id,'/">',htmlspecialchars($row->title),'</a></li>',"\n";
		}
	}
}

?><!doctype html>
<html>
<head>
<meta charset="UTF-8" />
<title>所有文章 - <?php echo htmlspecialchars($tbopt->get('blog_name')); ?></title>
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

