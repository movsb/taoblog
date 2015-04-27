<?php

function feed_header() {
	global $tbopt;

	header('HTTP/1.1 200 OK');
	header('Content-Type: application/xml');
?><?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title><?php echo htmlspecialchars($tbopt->get('blog_name')); ?></title>
		<link><?php echo $tbopt->get('home'); ?></link>
		<description></description>
<?php
}

function feed_body() {
	global $tbquery;
	global $tbdate;

	while($tbquery->has()) {
		$the = $tbquery->the();
		$link = the_post_link($the);
?>		<item>
			<title><?php echo htmlspecialchars($the->title); ?></title>
			<link><?php echo htmlspecialchars($link); ?></link>
			<description>&lt;p style="color: red;"&gt;&lt;a href="<?php echo $link;?>"&gt;=== 阅读原文 ===&lt;/a&gt;&lt;/p&gt;<?php echo htmlspecialchars($the->content); ?></description>
			<pubDate><?php echo $tbdate->the_feed_date($the->date);?></pubDate>
		</item>
<?php
	}
}

function feed_footer() {?>
	</channel>
</rss>
<?php } 

feed_header();
feed_body();
feed_footer();

