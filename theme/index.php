<?php
global $tbopt;
$blog_name = $tbopt->get('blog_name');

function the_recent_shuoshuos() {
    global $tbshuoshuo;
    global $tbsscmt;

    $sss = $tbshuoshuo->get_latest(10);
    if(!is_array($sss) || count($sss) == 0) return false;

    echo '<div id="shuoshuo">';
    echo '<h2>近期说说</h2>',PHP_EOL;
    echo '<ul>';
    foreach($sss as &$ss) {
        echo '<li id="shuoshuo-', $ss->id, '">';
        echo '<p>', $ss->content, ' <span>(', substr($ss->date,5,11), ')</span>',
            '<i title="发表评论" style="margin-left: 4px; cursor: pointer;" class="fa fa-pencil-square-o post-shuoshuo-comment" data-id="',$ss->id,'"></i>','</p>';
        echo '<div>';
        // 读取评论列表
        echo '<ul class="comment-list">';
        $cmts = $tbsscmt->get($ss->id);
        foreach($cmts as &$cmt) {
            echo '<li>';
            echo '<span class="author">',htmlspecialchars($cmt->author),'</span>: ';
            echo '<span class="content">',htmlspecialchars($cmt->content),'</span>';
            echo '</li>';
        }
        echo '</ul>';
        echo '</div>';
        echo '</li>';
    }
    echo '</ul>';
    echo '<p style="padding-left: 2em;"><a href="/shuoshuo">查看全部说说。</a></p>';
?>
<form id="shuoshuo-comment-form" method="post" action="/admin/shuoshuo.php" style="display: none;">
    <input type="submit" value="评论" style="display: none;"/>
    <input type="hidden" name="do" value="post-comment" />
    <input type="hidden" name="sid" value="" />
    <input type="text" name="author" required placeholder="昵称" style="width: 100px;"/>
    <input type="text" name="content" required placeholder="内容" />
    <input type="button" id="cancel-shuoshuo-comment" value="取消" />
</form>
    <script type="text/javascript">
        $('#shuoshuo-comment-form input[name=author]').val(localStorage.getItem('shuoshuo-author'));

        $('#cancel-shuoshuo-comment').on('click', function() {
            var form = $('#shuoshuo-comment-form');
            form.find('input[name=content]').val('');
            form.hide();
        });

        $('.post-shuoshuo-comment').on('click', function() {
            var sid     = $(this).attr('data-id');
            var from    = '#shuoshuo-comment-form';
            var to      = '#shuoshuo-'+sid+' .comment-list';
            $(from + '> input[name=sid]').val(sid);
            $(from).detach().appendTo(to).show();
        });
        
        $('#shuoshuo-comment-form').on('submit', function() {
            var self = $(this);

            if(self.attr('data-busy') == '1')
                return false;

            self.attr('data-busy', '1');

            var sid = $(this).find('input[name=sid]').val();

            $.post(self.attr('action'),
                self.serialize(),
                function(data) {
                    if(data.errno == 'ok') {
                        var s = '<li>'
                            + '<span class="author">' + data.author + '</span>: '
                            + '<span class="content">' + data.content + '</span>'
                            + '</li>';
                        $('#shuoshuo-'+sid+' .comment-list').append(s);
                        self.find('input[name=content]').val('');
                        localStorage.setItem('shuoshuo-author', self.find('input[name=author]').val());
                    }
                    else {
                        alert(data.error);
                    }
                },
                'json'
            )
            .fail(function(xhr, sta, e) {
                alert('未知错误！');
            })
            .always(function() {
                $('#shuoshuo-comment-form').hide();
                self.attr('data-busy','');
            });

            return false;
        });
    </script>
</div>
<!-- #shuoshuo end -->
<?php
}

function the_recent_posts() {
	global $tbpost;

	$q = ['pageno' => 1, 'no_content'=>true];
	$posts = $tbpost->query($q);
	if(is_array($posts) && count($posts)) {
		echo '<h2>近期文章</h2>',PHP_EOL;
		echo '<ul style="list-style: none;">';
		foreach($posts as &$p) {
			$link = the_link($p, false);
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
			echo '<li style="margin-bottom: 8px;"><b>',$c->author,'</b>: ',htmlspecialchars($c->content),
				' --- 《','<a href="/',$c->post_id,'/">',$title,'</a>》','</li>',PHP_EOL;
		}
		echo '</ul>';
	}
}

function tb_head() {?>
<style>
    #shuoshuo .post-shuoshuo-comment {
        visibility: hidden;
    }
    #shuoshuo li:hover .post-shuoshuo-comment {
        visibility: visible;
    }

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

the_recent_shuoshuos();
the_recent_posts();
the_recent_comments();

?>

<div>
    <h2>文章归档</h2>
    <p style="padding-left: 2em;"><a href="/archives">全部文章的归档页面，按标签、按分类、按日期。</a></p>
</div>

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

