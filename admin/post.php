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
		'title'		=> '分类',
		'content'	=> $content,
		'classname'	=> 'category',
		'types' => 'post',
		];
}

add_hook('post_widget', 'post_widget_tax');

function post_widget_page_parents($p=null) {
    global $tbpost;
    if($p) {
        $v = $tbpost->get_the_parents_string($p->id);
        if($v) {
            $v = substr($v, 1);
            $v = implode(',', explode('/',$v));
        }
    }
    $content = '<input type="text" name="page_parents" value="'.($p ? $v : '').'" />';

    return [
        'title' => '父页面',
        'content' => $content,
        'types' => 'page',
    ];
}

add_hook('post_widget', 'post_widget_page_parents');

function post_widget_tag($p=null) {
	$tag = $p ? join(',', $p->tag_names) : '';

	$title = '标签';
	$classname = 'tags';
	$types = 'post';
	$content = <<<EOD
<input type="text" name="tags" value="$tag" />
EOD;

	return compact('title', 'classname', 'types', 'content');
}

add_hook('post_widget', 'post_widget_tag');

function post_widget_slug($p=null) {
	return [
		'title' => '别名',
		'content' => '<input type="text" name="slug" value="'.($p ? htmlspecialchars($p->slug) : '').'" />',
		];
}

add_hook('post_widget', 'post_widget_slug');

function post_widget_date($p=null) {
	global $tbdate;

	$title = '日期';
	$content = '<input type="text" name="date" value="'.($p ? $p->date : '').'"/><br>'
		.'<input type="text" name="modified" value="'.($p ? '-' : '').'" />';

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

	.widget-category .widget-content {
		max-height: 200px;
		overflow: auto;
	}

	.widget-content input[type=text], .widget-content textarea {
		padding: 4px;
		width: 100%;
		box-sizing: border-box;
	}

	#content {
		max-height: 2000px;
		height: 500px;
		min-height: 300px;
		width: 100%;
		padding: 4px;
		box-sizing: border-box;
	}

	.fullscreen {
		position: fixed;
		left: 0px;
		top: 0px;
		width: 100%;
		height: 100%;
		z-index: 1000;
		padding: 1px;
		box-sizing: border-box;
	}

	.fullscreen #content {
		width: 100%;
		height: 100%;
		max-width: 100%;
		max-height: 100%;
		box-sizing: border-box;
		padding: 6px;
	}

</style>
<?php }

add_hook('admin_head', 'post_admin_head');

function post_admin_footer() { ?>
	<script type="text/javascript">
		$('.widget h3').click(function(e) {
			var div = e.target.nextElementSibling;
			$(div).toggle();
		});
	</script>
<?php
}

add_hook('admin_footer', 'post_admin_footer');

function new_post_html($p=null){
	global $tbpost;

	// 先生成所有的挂件对象
	// 因为分布在不同地方（hook对象无法保存这些分布）
	$widgets = [];

	$widget_objs = get_hooks('post_widget');
	foreach($widget_objs as $wo) {
		$fn = $wo->func;
		$w = (object)$fn($p);
		$w->classname = isset($w->classname) ? $w->classname : 'widget';

		$dom = <<< DOM
<div class="widget widget-$w->classname">
	<h3>$w->title</h3>
	<div class="widget-content">
		$w->content
	</div>
</div> 
DOM;
		$widget = new stdClass;
		$widget->dom = $dom;
		$widget->pos = isset($w->position) ? $w->position : 'right';
		$widget->types = isset($w->types) ? $w->types : 'post,page';

		$widgets[] = $widget;
	}

	$type = $p ? $p->type : (isset($_GET['type']) ? $_GET['type'] : '');
	if(!in_array($type, ['post','page']))
		$type = 'post';

?><div id="admin-post">
	<form method="POST" id="form-post">
		<div class="post" style="float: left; width: 100%; max-width: 75%;">
			<div class="post-area">
				<div style="margin-bottom: 1em;">
					<h2>标题</h2>
					<div>
					<input style="padding: 8px; width: 100%; box-sizing: border-box;" type="text" name="title" value="<?php
						if($p) {
							echo htmlspecialchars($p->title);
						}
					?>" />
					</div>
				</div>
				<?php if($p) {
					$link = the_link($p);
				?>
				<div class="permanlink" style="margin-bottom: 1em;">
					<span>固定链接：</span>
					<a id="permalink" href="<?php echo $link; ?>"><?php echo $link; ?></a>
					<script type="text/javascript">
						var new_window = null;
						$('#permalink').click(function() {
							if(!new_window || new_window.closed) {
								new_window = window.open($('#permalink').prop('href'));
							} else {
								new_window.location.href = $('#permalink').prop('href');
							}
							return false;
						});
					</script>
				</div>
				<?php } else {
					$next_id = $tbpost->the_next_id();
				?>
				<div class="permalink_id" style="margin-bottom: 1em;">
					<span>文章ID：</span>
					<span><?php echo $next_id; ?></span>
				</div>
				<?php } ?>
				<div>
					<div><h2 style="float: left; margin-right: 10px;">内容</h2><span id="msg" style="position: relative; top: 6px;"></span></div>
					<div class="textarea-wrap">
						<textarea id="content" name="content" wrap="off"><?php
							if($p) {
								echo htmlspecialchars($p->content);
							}
						?></textarea>
						<script>
							// 无法使用 keyup，因为 keyup 是在默认事件调用之后才被调用
							$('#content').on('keydown', function(e) {
								if(e.keyCode == 122) { // f11
									$('.textarea-wrap').toggleClass('fullscreen');
									e.preventDefault();
									return false;
								} else if(e.keyCode == 83) { // s
									if(e.originalEvent.ctrlKey == true) { // ctrl + s
										localStorage.setItem('post_content', $('#content').val());
										$('#msg').html('(<span style="color: green;">草稿已保存</span>)').show();
										setTimeout(function(){$('#msg').fadeOut();},1000);
										e.preventDefault();
										return false;
									}
								}
							});
						</script>
					</div>
				</div>
				<div>
					<input type="hidden" name="do" value="<?php echo $p ? 'update' : 'new'; ?>" />
					<input type="hidden" name="type" value="<?php echo $p ? $p->type : $type; ?>" />
					<?php if($p) { ?><input type="hidden" name="id" value="<?php echo $p->id; ?>" /><?php } ?>
				</div>
			</div><!-- post-area -->
			<div class="sidebar sidebar-left">
				<?php foreach($widgets as &$widget) {
					if(in_array($type, explode(',',$widget->types)) && $widget->pos == 'left') {
						echo $widget->dom;
					}
				} ?>
			</div>
		</div><!-- post -->
		<div class="sidebar sidebar-right" style="float: right;">
			<div class="widget">
				<h3>发表</h3>
				<div class="widget-content">
					<input type="reset" value="清空" onclick="return confirm('确定清空吗？');" />
					<input id="btn-preview" type="button" value="预览" />
					<input type="submit" value="发表为" />
                    <select name="status">
                        <option value="public"<?php if($p && $p->status == 'public') echo 'selected'; ?>>公开</option>
                        <option value="draft"<?php if($p && $p->status == 'draft') echo 'selected'; ?>>草稿</option>
                    </select>
					<script>
						$('#btn-preview').click(function() {
							var form = $('#form-post');
							var ido = $('input[name=do]');
							var doval = ido.val();
							ido.val('preview');
							form.attr('target', '_blank');
							form.submit();
							form.attr('target', '');
							ido.val(doval);
						});
					</script>
				</div>
			</div>
			<?php foreach($widgets as &$widget) {
				if(in_array($type, explode(',',$widget->types)) && $widget->pos == 'right') {
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

	// 输出编辑内容之前过滤
	if(isset($post[0]->content))
		$post[0]->content = apply_hooks('edit_the_content', $post[0]->content, $post[0]->id);

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
require_once('admin-hooks-post.php');

function post_new_post() {
	global $tbpost;
	global $tbmain;

	if(($id=$tbpost->insert($_POST))){
		apply_hooks('post_posted', $id, $_POST);
		header('HTTP/1.1 302 Found');
		header('Location: '.$tbmain->home.'/admin/post.php?do=edit&id='.$id);
		die(0);
	} else {
		$j = [ 'errno' => 'error', 'error' => $tbpost->error];
		post_die_json($j);
	}
}

function post_update() {
	global $tbpost;
    global $tbmain;

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
	header('Location: '.$tbmain->home.'/admin/post.php?do=edit&id='.$id);
	die(0);
}

function post_preview() {
	@include TBPATH.'theme/preview.php';
	die(0);
}

$do = $_POST['do'];

if($do == 'new') {
	post_new_post();
} else if($do == 'update') {
	post_update();
} else if($do == 'preview') {
	post_preview();
}

die(0);

endif; // POST

