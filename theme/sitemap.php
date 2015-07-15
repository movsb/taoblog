<?php

function sitemap_header() {
	header('HTTP/1.1 200 OK');
	header('Content-Type: application/xml');
?><?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<?php
}

function sitemap_body() {
	global $tbpost;
	global $tbopt;

	$home = $tbopt->get('home');
	$ids = $tbpost->get_all_posts_id();

	foreach($ids as $id) {
		echo '<url><loc>',$home,'/',$id,'/','</loc></url>',PHP_EOL;
	}
}

function sitemap_footer() {?>
</urlset>
<?php } 

sitemap_header();
sitemap_body();
sitemap_footer();

