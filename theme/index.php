<?php
$home = $tbopt->get('home');
$blog_name = $tbopt->get('blog_name');

function the_recent_posts() {
	global $tbpost;

	$q = ['pageno' => 1, 'no_content'=>true];
	$posts = $tbpost->query($q);
	if(is_array($posts) && count($posts)) {
		echo '<h2>近期文章</h2>',PHP_EOL;
		echo '<ul style="list-style: none;">';
		foreach($posts as &$p) {
			$link = the_link($p);
			echo '<li><a href="'.$link.'">',$p->title,'</a></li>',"\n";
		}
		echo '</ul>';
	}
}

function the_recent_comments() {
	global $tbcmts;
	global $tbpost;

	$cmts = $tbcmts->get_recent_comments();
	if(is_array($cmts) && count($cmts)) {
		echo '<h2>近期评论</h2>',PHP_EOL;
		echo '<ul style="list-style: none;">';
		foreach($cmts as $c) {
			$title = $tbpost->get_vars('title',"id=$c->post_id")->title;
			echo '<li style="margin-bottom: 8px;"><span style="color: green">',$c->author,'</span>: ',htmlspecialchars($c->content),
				' --- 《','<a href="/',$c->post_id,'/">',$title,'</a>》','</li>',PHP_EOL;
		}
		echo '</ul>';
	}
}

require('header.php');

today_english();
the_recent_posts();
the_recent_comments();

?>

<div>
	<h2>又换风格了</h2>
	<p>改来改去，还是觉得不喜欢现在的风格。怎么办？</p>
</div>

<div>
	<h2>状态</h2>
	<p>服务器开始运行于2015年4月1日，已经运行 <span id="server-run-time">?</span> 天。</p>
</div>

<?php

require('footer.php');

