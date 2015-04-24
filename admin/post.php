<?php

$admin_url = 'post.php';

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once('admin.php');

function post_widget_tax_add(&$taxes, $tax=1) {
	$s = '';
	foreach($taxes as $t) {
		$s .= '<li><label><input type="radio" style="position: relative; top: 2px;" name="taxonomy" value="'.$t->id.'"'.
			($tax==$t->id?' checked="checked" ':' ').' /> '.$t->name."</label>\n";
		if(isset($t->sons) && count($t->sons)) {
			$s .= '<ul class="children" style="margin-left: 14px;">';
			$s .= post_widget_tax_add($t->sons, $tax);
			$s .= "</ul>\n";
		}
		$s .= '</li>'."\n";
	}
	return $s;
}

function post_widget_tax($p=null) {
	global $tbtax;
	$taxes = $tbtax->get_hierarchically();

	$content = '<ul>'.post_widget_tax_add($taxes, ($p ? $p->taxonomy : 1)).'</ul>';

	return [
		'title' => '分类',
		'content' => $content,
		];
}

add_hook('post_widget', 'post_widget_tax');

function post_widget_slug($p=null) {
	return [
		'title' => 'Slug',
		'content' => '<input type="text" name="slug" value="'.($p ? htmlspecialchars($p->slug) : '').'" />',
		];
}

add_hook('post_widget', 'post_widget_slug');

function post_widget_date($p=null) {
	global $tbdate;

	$title = '日期';
	$content = '<input type="text" name="date" value="'.($p ? $p->date : $tbdate->mysql_datetime_local()).'"/><br>'
		.'<input type="text" name="modified" value="" />';

	return compact('title', 'content');
}

add_hook('post_widget', 'post_widget_date');

function post_admin_head() { ?>
<style>
	.sidebar {

	}

	.sidebar input[type="text"] {
		padding: 4px;
	}

	.sidebar .widget {
		background-color: white;
		border: 1px solid #ccc;
		margin-bottom: 20px;
	}

	.sidebar .widget h3 {
		padding: 4px 6px;
		border-bottom: 1px solid #ccc;
	}

	.sidebar .widget-content {
		padding: 10px;
	}

	.sidebar .widget ul {
		list-style: none;
	}

	.post-area {
		margin-bottom: 3em;
	}
</style>
<?php }

add_hook('admin_head', 'post_admin_head');

function new_post_html($p=null){
	// 先生成所有的挂件对象
	// 因为分布在不同地方（hook对象无法保存这些分布）
	$widgets = [];

	$widget_objs = get_hooks('post_widget');
	foreach($widget_objs as $wo) {
		$fn = $wo->func;
		$w = (object)$fn($p);

		$dom = <<< DOM
<div class="widget">
	<h3>$w->title</h3>
	<div class="widget-content">
		$w->content
	</div>
</div> 
DOM;
		$widget = new stdClass;
		$widget->dom = $dom;
		$widget->pos = isset($w->position) ? $w->position : 'right';

		$widgets[] = $widget;
	}
?><div id="admin-post">
	<form method="POST">
		<div class="post" style="float: left;">
			<div class="post-area">
				<div>
					<h2>标题</h2>
					<input type="text" name="title" value="<?php
						if($p) {
							echo htmlspecialchars($p->title);
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
					<input type="hidden" name="do" value="<?php echo $p ? 'update' : 'new'; ?>" />
					<?php if($p) { ?><input type="hidden" name="id" value="<?php echo $p->id; ?>" /><?php } ?>
				</div>
			</div><!-- post-area -->
			<div class="sidebar sidebar-left">
				<?php foreach($widgets as &$widget) {
					if($widget->pos == 'left') {
						echo $widget->dom;
					}
				} ?>
			</div>
		</div><!-- post -->
		<div class="sidebar sidebar-right" style="float: right;">
			<div class="widget">
				<h3>发表</h3>
				<div class="widget-content">
					<input type="reset" value="清空" />
					<input type="submit" value="发表" />
				</div>
			</div>
			<?php foreach($widgets as &$widget) {
				if($widget->pos == 'right') {
					echo $widget->dom;
				}
			} ?>
		</div><!-- sidebar right -->
	</form>
	<?php if(!$p) {?>
	<script type="text/javascript">
		document.getElementsByTagName('form')[0].reset();
	</script>
	<?php } ?>
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
	$arg = ['id' => $id];
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

function post_die_json($arg) {
	header('HTTP/1.1 200 OK');
	header('Content-Type: application/json');

	echo json_encode($arg);
	die(0);
}

require_once('login-auth.php');

if(!login_auth()) {
	post_die_json([
		'errno' => 'unauthorized',
		'error' => '需要登录后才能进行该操作！',
		]);
}

require_once('load.php');

function post_new_post() {
	global $tbdb;
	global $tbpost;
	global $tbopt;

	if(($id=$tbpost->insert($_POST))){
		apply_hooks('post_posted', $id, $_POST);
		header('HTTP/1.1 302 Found');
		header('Location: '.$tbopt->get('home').'/admin/post.php?do=edit&id='.$id);
		die(0);
	} else {
		$j = [ 'errno' => 'error', 'error' => $tbpost->error];
		post_die_json($j);
	}
}

function post_update() {
	global $tbdb;
	global $tbpost;
	global $tbopt;

	$r = $tbpost->update($_POST);
	if(!$r) {
		post_die_json([
			'errno' => 'error',
			'error' => $tbpost->error
			]);
	}

	$id = (int)$_POST['id'];
	apply_hooks('post_updated', $id, $_POST);

	header('HTTP/1.1 302 Updated');
	header('Location: '.$tbopt->get('home').'/admin/post.php?do=edit&id='.$id);
	die(0);
}

$do = $_POST['do'];

if($do == 'new') {
	post_new_post();
} else if($do == 'update') {
	post_update();
}


die(0);

endif; // POST

