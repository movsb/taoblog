<?php

$admin_url = 'post.php';

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once('admin.php');

function new_post_html($p=null){ ?>
<div id="admin-post">
	<form method="POST">
		<div>
			<div>
				<h2>标题</h2>
				<input type="text" name="title" value="<?php
					if($p) {
						echo htmlspecialchars($p->title);
					}
				?>" />
			</div>
			<div>
				<h2>别名</h2>
				<input type="text" name="slug" value="<?php
					if($p) {
						echo htmlspecialchars($p->slug);
					}
				?>" />
			</div>
			<div>
				<h2>内容</h2>
				<textarea name="content" wrap="off" style="width: 500px; height: 300px;"><?php
					if($p) {
						echo htmlspecialchars($p->content);
					}
				?></textarea>
			</div>
			<div>
				<input type="reset" value="清空" />
				<input type="submit" value="发表" />
			</div>
			<div>
				<input type="hidden" name="do" value="new-post" />
			</div>
		</div>
	</form>
</div><!-- admin-post -->
<?php } 

$do = isset($_GET['do']) ? $_GET['do'] : '';
if(!$do) {
	admin_header();
	new_post_html();
	admin_footer();
	die(0);
} else if($do === 'edit') {
	$id = intval($_GET['id']);
	$arg = ['p' => $id, 'noredirect'=>true];
	$post = $tbpost->query($arg);
	if($post === false || empty($post)){
		tb_die(200, '没有这篇文章！');
	}

	admin_header();
	new_post_html($post[0]);
	admin_footer();
	die(0);
}


die(0);

/* GET */ else :

require_once('load.php');

function post_new_post() {
	global $tbdb;
	global $tbpost;
	global $tbopt;

	$title = $_POST['title'];
	$content = $_POST['content'];
	$slug = $_POST['slug'];

	$p = compact('title', 'content', 'slug');

	if(($id=$tbpost->insert($p))){
		header('HTTP/1.1 302 Found');
		header('Location: '.$tbopt->get('home').'/admin/post.php?do=edit&id='.$id);
		die(0);
	} else {
		$j = [ 'errno' => 'failed',];
		tb_die(400, $tbdb->error);
	}
}

if(!isset( $_POST['do'])){
	tb_die(400, '未指定动作！');
}

$do = str_replace('-','_','post-'.$_REQUEST['do']);

if(!function_exists($do)){
	tb_die(400, "未找到函数($do)！");
}

$do();

die(0);

endif; // POST

