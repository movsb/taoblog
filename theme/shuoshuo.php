<?php

require('admin/load.php');

function get_all_shuoshuos() {
	global $tbdb;

	$sql = "SELECT id,content FROM shuoshuo WHERE 1 ORDER BY date DESC";

	$results = $tbdb->query($sql);

	if($results) {
		while($row = $results->fetch_object()) {
			echo '<li>', $row->content, '</li>',"\n";
		}
	}
}

?><!doctype html>
<html>
<head>
<meta charset="UTF-8" />
<title>所有说说 - <?php echo $tbopt->get('blog_name'); ?></title>
</head>
<body>
<h1>所有说说：</h1>
<div style="padding: 0px 2em;">
<ul>
<?php get_all_shuoshuos(); ?>
</ul>
</div>
</body>
</html>
<?php

