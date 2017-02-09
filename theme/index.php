<?php
global $tbopt;
$blog_name = $tbopt->get('blog_name');

function the_recent_posts() {
	global $tbpost;

	$q = ['pageno' => 1, 'no_content'=>true];
	$posts = $tbpost->query($q);
	if(is_array($posts) && count($posts)) {
		echo '<h2>近期文章</h2>',PHP_EOL;
		echo '<ul style="list-style: none;">';
		foreach($posts as &$p) {
			$link = the_link($p, false);
			echo '<li><a href="'.$link.'">',htmlspecialchars($p->title),'</a></li>',"\n";
		}
		echo '</ul>';
	}
}

function the_recent_comments() {
	global $tbcmts;
	global $tbpost;
    global $tbopt;

    $admin_email = $tbopt->get('email');

	$cmts = $tbcmts->get_recent_comments();
	if(is_array($cmts) && count($cmts)) {
		echo '<h2>近期评论</h2>',PHP_EOL;
		echo '<ul style="list-style: none;">';
		foreach($cmts as $c) {
			$title = $tbpost->get_vars('title',"id=$c->post_id")->title;
            $author = strcasecmp($c->email, $admin_email) == 0 ? '博主' : $c->author;

			echo '<li style="margin-bottom: 8px; overflow: hidden;"><b>', htmlspecialchars($author),'</b>: ',htmlspecialchars($c->content),
				'<span style="float: right;">《','<a href="/',$c->post_id,'/">',htmlspecialchars($title),'</a>》</span>','</li>',PHP_EOL;
		}
		echo '</ul>';
	}
}

function tb_head() {?>
<style>
#main a {
  text-decoration: none;
  color: #005782; }
#main a:hover {
    text-decoration: underline; }
#main  a:visited {
    color: #924646; }
#main  a:focus {
    outline: none; }

a, input, textarea, button {
  transition: all 0.2s ease-out; }

button, input[type=submit] {
  border: none; }
  button::-moz-focus-inner, input[type=submit]::-moz-focus-inner {
    border: 0; }

acronym, abbr {
  border-bottom: 1px dashed #ccc; }

input:focus, a:focus, textarea:focus {
  outline: none; }

input, textarea, button {
      border: 1px solid #ccc;
        padding: 5px 7px; }
</style>
<?php }
add_hook('tb_head', 'tb_head');

require('header.php');

the_recent_posts();
the_recent_comments();

?>

<div>
	<h2>状态</h2>
	<p style="padding-left: 2em;">服务器开始运行于2014年12月24日，已运行 <span id="server-run-time">?</span> 天。</p>
    <p style="padding-left: 2em;">博客归档：<?php
        echo '文章数：', $tbopt->get('post_count', '?');
        echo '，页面数：', $tbopt->get('page_count', '?');
        echo '，评论数：', $tbopt->get('comment_count', '?');
        echo '。';
    ?></p>
</div>

<?php

require('footer.php');

