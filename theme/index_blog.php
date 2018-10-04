<?php
global $tbopt;
$blog_name = $tbopt->get('blog_name');

function the_recent_posts() {
    global $tbquery;

    $posts = $tbquery->objs;
    if(is_array($posts) && count($posts)) {
        echo '<h2>近期文章</h2>',PHP_EOL;
        echo '<ul>';
        foreach($posts as &$p) {
            $link = the_link($p, false);
            echo '<li><a href="'.$link.'">',htmlspecialchars($p->title),'</a></li>',"\n";
        }
        echo '</ul>';
    }
    echo '<p style="padding-left:3em;"><a href="/all-posts.html">所有文章...</a></p>';
}

function the_recent_comments() {
    global $tbpost;
    global $tbopt;

    $admin_email = $tbopt->get('email');

    $cmts = Invoke('/posts!recentComments', 'json', null, false);
    $cmts = json_decode($cmts);

    if(is_array($cmts) && count($cmts)) {
        echo '<h2>近期评论</h2>',PHP_EOL;
        echo '<ul>';
        foreach($cmts as $c) {
            $title = $tbpost->get_vars('title',"id=$c->PostID")->title;
            $author = strcasecmp($c->EMail, $admin_email) == 0 ? '博主' : $c->Author;

            echo '<li style="margin-bottom: 8px; overflow: hidden;"><b>', htmlspecialchars($author),'</b>: ',htmlspecialchars($c->Content),
                '<span style="float: right;">《','<a href="/',$c->PostID,'/#comments">',htmlspecialchars($title),'</a>》</span>','</li>',PHP_EOL;
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

#content > ul, .status > ul {
    list-style: none;
}

/* TODO 这个最大宽度应该是变量 */
@media screen and (max-width: 850px) {
#content > ul, .status > ul {
    padding: 0px;
}
}
</style>
<?php }
add_hook('tb_head', 'tb_head');

require('header.php');

the_recent_posts();
the_recent_comments();

?>

<div class="status">
    <h2>状态</h2>
    <ul>
    <li>服务器开始运行于2014年12月24日，已运行 <span id="server-run-time">?</span> 天。</li>
    <li>博客归档：<?php
        echo '文章数：', $tbopt->get('post_count', '?');
        echo '，页面数：', $tbopt->get('page_count', '?');
        echo '，评论数：', $tbopt->get('comment_count', '?');
        echo '。';
    ?></li>
    </ul>
</div>

<?php

require('footer.php');

