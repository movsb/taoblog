<?php

function feed_header() {
    global $tbdate;
    global $tbopt;

    header('HTTP/1.1 200 OK');
    header('Content-Type: application/xml');

    $last_post_time = $tbopt->get('last_post_time');
    if($last_post_time) {
        $last_post_time_hg = $tbdate->mysql_local_to_http_gmt($last_post_time);
        header('Last-Modified: '.$last_post_time_hg);
    }

?><?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <title><?php echo htmlspecialchars($tbopt->get('blog_name')); ?></title>
        <link><?php echo home(); ?></link>
        <description></description>
<?php
}

function feed_body() {
    global $tbquery;
    global $tbdate;

    while($tbquery->has()) {
        $the = $tbquery->the();
        $title = $the->title;
        $link = the_link($the);
        $content = '<![CDATA[' . str_replace(']]>', ']]]]><![CDATA[>', $the->content) . ']]>';

?>		<item>
            <title><?php echo htmlspecialchars($the->title); ?></title>
            <link><?php echo htmlspecialchars($link); ?></link>
            <description><?php echo $content; ?></description>
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

