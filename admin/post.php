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

function post_widget_snjs($p=null) {
	global $tbsnjs;

	if($p) $snjs = $tbsnjs->get_snjs($p->id);

	$title = 'SnJS';
	$content = '<textarea wrap="off" name="snjs_header">'.($p ? htmlspecialchars($snjs->post->header) : '').'</textarea><br>'
		.'<textarea wrap="off" name="snjs_footer">'.($p ? htmlspecialchars($snjs->post->footer) : '').'</textarea>';

	return compact('title', 'content');

}

add_hook('post_widget', 'post_widget_snjs');

function post_admin_head() { ?>
<style>
	.sidebar-right {

	}

	.sidebar-right input[type="text"] {
		padding: 4px;
	}

	.sidebar-right .widget {
		background-color: white;
		border: 1px solid #ccc;
		margin-bottom: 20px;
	}

	.sidebar-right .widget h3 {
		padding: 4px 6px;
		border-bottom: 1px solid #ccc;
	}

	.sidebar-right .widget-content {
		padding: 10px;
	}

	.sidebar-right .widget ul {
		list-style: none;
	}
</style>
<?php }

add_hook('admin_head', 'post_admin_head');

function new_post_html($p=null){ ?>
<div id="admin-post">
	<form method="POST">
		<div class="post" style="float: left;">
			<div>
				<h2>标题</h2>
				<input type="text" name="title" value="<?php
					if($p) {
						echo htmlspecialchars($p->title);
					}
				?>" />
			</div>
			<!--div>
				<h2>别名</h2>
				<input type="text" name="slug" value="<?php
					if($p) {
						echo htmlspecialchars($p->slug);
					}
				?>" />
			</div-->
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
				<input type="hidden" name="do" value="<?php echo $p ? 'update' : 'new'; ?>" />
				<?php if($p) { ?><input type="hidden" name="id" value="<?php echo $p->id; ?>" /><?php } ?>
			</div>
		</div><!-- post -->
		<div class="sidebar-right" style="float: right;">
			<?php
				$widgets = get_hooks('post_widget');
				foreach($widgets as $wo) {
					$fn = $wo->func;
					$w = (object)$fn($p);
					?>
<div class="widget">
	<h3><?php echo $w->title; ?></h3>
	<div class="widget-content">
		<?php echo $w->content; ?>
	</div>
</div><?php } ?>
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
	global $tbsnjs;

	if(($id=$tbpost->insert($_POST))){
		$tbsnjs->insert($id);
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
	global $tbsnjs;

	$r = $tbpost->update($_POST);
	if(!$r) {
		post_die_json([
			'errno' => 'error',
			'error' => $tbpost->error
			]);
	}

	$tbsnjs->update((int)$_POST['id']);

	header('HTTP/1.1 302 Updated');
	header('Location: '.$tbopt->get('home').'/admin/post.php?do=edit&id='.$_POST['id']);
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

