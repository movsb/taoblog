<?php
$home = $tbopt->get('home');
$blog_name = $tbopt->get('blog_name');

function the_recent_posts() {
	global $tbpost;

	$q = ['pageno' => 1, 'no_content'=>true];
	$posts = $tbpost->query($q);
	if(is_array($posts) && count($posts)) {
		echo '<h2>近期文章: </h2>',PHP_EOL;
		echo '<ul style="font-size: 1.2em; list-style: none;">';
		foreach($posts as &$p) {
			$link = $p->type == 'post' ? the_post_link($p) : the_page_link($p);
			echo '<li><a href="'.$link.'">',$p->title,'</a></li>',"\n";
		}
		echo '</ul>';
	}
}

require('header.php');

the_recent_posts();

?>

<div>
	<h2>又换风格了</h2>
	<p>改来改去，还是觉得不喜欢现在的风格。怎么办？</p>
</div>

<?php

require('footer.php');

