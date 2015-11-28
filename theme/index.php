<?php
$home = $tbopt->get('home');
$blog_name = $tbopt->get('blog_name');

function the_baidu_search() {
?>
<div>
<h2>站内搜索</h2>
<div style="padding: 0px 2em;">
<script type="text/javascript">(function(){document.write(unescape('%3Cdiv id="bdcs"%3E%3C/div%3E'));var bdcs = document.createElement('script');bdcs.type = 'text/javascript';bdcs.async = true;bdcs.src = 'http://znsv.baidu.com/customer_search/api/js?sid=6111287814214412158' + '&plate_url=' + encodeURIComponent(window.location.href) + '&t=' + Math.ceil(new Date()/3600000);var s = document.getElementsByTagName('script')[0];s.parentNode.insertBefore(bdcs, s);})();</script>
</div>
</div>
<?php
}
function the_recent_shuoshuos() {
    global $tbshuoshuo;

    $sss = $tbshuoshuo->get_latest(10);
    if(count($sss) == 0) return false;

    echo '<h2>近期说说</h2>',PHP_EOL;
    echo '<ul>';
    foreach($sss as &$ss) {
        echo '<li>',
            '<p>',$ss->content,
            ' <span>(',substr($ss->date,5,11),')</span>',
            '</p>',
            '</li>';
    }
    echo '</ul>';


}

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

the_baidu_search();
the_recent_shuoshuos();
the_recent_posts();
the_recent_comments();

?>

<div>
    <h2>文章归档</h2>
    <p><a href="/archives">全部文章的归档页面，按标签、按分类、按日期。</a></p>
</div>

<div>
	<h2>状态</h2>
	<p>服务器开始运行于2014年12月24日，已运行 <span id="server-run-time">?</span> 天。</p>
</div>

<?php

require('footer.php');

