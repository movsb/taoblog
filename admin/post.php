<?php

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once('admin.php');

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

function post_widget_metas($p=null) {
    $metas = str_replace(['\\','\''], ['\\\\','\\\''], $p ? $p->metas_raw : '{}');
    $title = '自定义';
    $classname = 'metas';
    $content = <<< DOM
<label>类型：</label>
<select class="keys">
    <option>&lt;新建&gt;</option>
</select>
<span class="new">
    <input class="key" type="text" style="width: 100px;" />
    <button class="ok">添加</button>
</span>
<textarea class="content" style="margin-top: 10px; height: 200px; display: block;"></textarea>

<input type="hidden" name="metas" value="" />

<script>
(function() {
    var keys = $('.widget-metas .keys');
    var metas = JSON.parse('$metas');
    var newf = $('.widget-metas .new');
    var content = $('.widget-metas .content');

    $('.widget-metas input[name=metas]').val('$metas');

    for(var key in metas) {
        keys.append($('<option>', {value: key, text: key}));
    }

    var prev_key = '';

    function save_prev() {
        if(prev_key) {
            metas[prev_key] = content.val();
        }
    }

    content.on('blur', function() {
        save_prev();
        $('.widget-metas input[name=metas]').val(JSON.stringify(metas));
    });

    keys.on('change', function() {
        var i = this.selectedIndex;


        if(i == 0) {
            newf.css('visibility', 'visible');
            prev_key = '';
            content.val('');
        }
        else {
            newf.css('visibility', 'hidden');
            prev_key = keys[0].options[i].value;
            content.val(metas[prev_key]);
        }

    });

    $('.widget-metas .new .ok').on('click', function() {
        var key = $('.widget-metas .new .key').val().trim();
        if(key == '' || metas.hasOwnProperty(key)) {
            alert('为空或已经存在。');
            return false;
        }

        keys.append($('<option>', {value: key, text: key}));
        keys.val(key);
        prev_key = key;
        content.focus();
        newf.css('visibility', 'hidden');

        return false;
    });
})();
</script>
DOM;

    return compact('title', 'content', 'classname');
}

add_hook('post_widget', 'post_widget_metas');

function post_widget_tax_add(&$taxes, $tax=1) {
	$s = '';
	foreach($taxes as $t) {
		$s .= '<li style="margin-bottom: 4px;"><label><input type="radio" style="margin-right: 6px" name="taxonomy" value="'.$t->id.'"'.
			($tax==$t->id?' checked="checked"':'').'/>'.htmlspecialchars($t->name)."</label>\n";
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

function post_admin_head() {
    $post = $GLOBALS['__p__'] ?? null;

    echo '<title>', $post ? '【编辑文章】'.htmlspecialchars($post->title) : '新文章','</title>';

?>

<script src="scripts/marked.js"></script>

<script>
    var renderer = new marked.Renderer();
    renderer.code = function(code, lang) {
        var beg = '<pre class="code" lang="' + (lang === undefined ? '' : lang) + '">\n';
        var text = code.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        var end = '\n</pre>';
        return beg + text + end;
    }
    renderer.hr = function() {
        return '<hr/>';
    }
    renderer.br = function() {
        return '<br/>';
    }
    marked.setOptions({renderer: renderer});
</script>

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

	#source {
		max-height: 2000px;
		height: 70vh;
		min-height: 300px;
		width: 100%;
		padding: 4px;
		box-sizing: border-box;
	}

#form-post {
    display: flex;
}

.post {
    flex: 1;
}
.sidebar-right {
    flex: 1;
}

@media screen and (min-width: 851px) {
    .sidebar-right {
        width: 280px;
        max-width: 280px;
        min-width: 280px;
    }
    .post {
        margin-right: 1em;
    }
}
@media screen and (max-width: 850px) {
    #form-post {
        flex-direction: column;
    }
}
</style>
<?php
    apply_hooks('admin:post:head');
}

add_hook('admin_head', 'post_admin_head');

function post_admin_footer() { ?>
	<script type="text/javascript">
		$('.widget h3').click(function(e) {
			var div = e.target.nextElementSibling;
			$(div).toggle();
		});
	</script>
<?php
    apply_hooks('admin:post:foot');
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
		$w->classname = $w->classname ?? 'widget';

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
		$widget->pos = $w->position ?? 'right';
		$widget->types = $w->types ?? 'post,page';

		$widgets[] = $widget;
	}

	$type = $p ? $p->type : ($_GET['type'] ?? '');
	if(!in_array($type, ['post','page']))
		$type = 'post';

?><div id="admin-post">
	<form method="POST" id="form-post">
		<div class="post">
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
					<h2>内容</h2>
					<div class="textarea-wrap">
                        <input type="hidden" id="content" name="content"/>
						<textarea id="source" name="source" wrap="off"><?php
							if($p) {
								echo htmlspecialchars($p->source ? $p->source : $p->content);
							}
						?></textarea>
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
		<div class="sidebar sidebar-right">
			<div class="widget widget-post">
				<h3>发表</h3>
				<div class="widget-content">
                    <?php if($p) { ?>
                    <input type="button" onclick="do_preview()" value="预览" />
                    <script>
                        function do_preview() {
                            var form = $('#form-post');
                            form.attr('target', '_blank');
                            form.attr('action', '/<?php echo $p->id; ?>/');
                            form.find('input[name=do]').val('preview');
                            form.submit();
                            form.attr('target', '')
                            form.attr('action', '')
                            form.find('input[name=do]').val('update');
                        }
                    </script>
                    <?php } ?>
					<input type="submit" value="发表" />
                    <select name="status">
                        <option value="public"<?php if($p && $p->status == 'public') echo ' selected'; ?>>公开</option>
                        <option value="draft"<?php if($p && $p->status == 'draft') echo ' selected'; ?>>草稿</option>
                    </select>
                    <select name="source_type">
                        <?php if($p && $p->source_type === '') $p->source_type = 'html'; ?>
                        <option value="markdown"<?php if($p && $p->source_type == 'markdown') echo ' selected'; ?>>Markdown</option>
                        <option value="html"<?php if($p && $p->source_type == 'html') echo ' selected'; ?>>HTML</option>
                    </select>
				</div>
			</div>
			<?php foreach($widgets as &$widget) {
				if(in_array($type, explode(',',$widget->types)) && $widget->pos == 'right') {
					echo $widget->dom;
				}
			} ?>
		</div><!-- sidebar right -->
        <script>
            // TODO 临时代码，在切换源的类型时切换编辑器语法
            $('select[name="source_type"]').change(function() {
                console.log('源类型切换为：', this.value);
                if(typeof codemirror == 'object') {
                    var mode = '';
                    var value = this.value;

                    if(value == 'markdown')
                        mode = 'markdown';
                    else if(value == 'html')
                        mode = 'htmlmixed';

                    codemirror.setOption('mode', mode);
                }
            }).change();

            $('#form-post').submit(function() {
                var source = $('#source');
                var content = $('input[name="content"]');
                var type = $('select[name="source_type"]').val();
                if(type == 'html') {
                    content.val(source.val());
                }
                else if(type == 'markdown') {
                    var result = marked(source.val());
                    content.val(result);
                }
                else {
                    alert('bad source type.');
                    return false;
                }
            });
        </script>
	</form>
	<script type="text/javascript">
		document.getElementsByTagName('form')[0].reset();
	</script>
</div><!-- admin-post -->
<?php } 

$do = $_GET['do'] ?? '';
if(!$do) {
	admin_header();
	new_post_html();
	admin_footer();
	die(0);
} else if($do === 'edit') {
	$id = intval($_GET['id']);
	$post = $tbpost->query_by_id($id,'');
	if($post === false || empty($post)){
		tb_die(200, '没有这篇文章！');
	}

	// 输出编辑内容之前过滤
	if(isset($post[0]->content))
		$post[0]->content = apply_hooks('edit_the_content', $post[0]->content, $post[0]->id);

    // 罪过，使用全局变量了
    $GLOBALS['__p__'] = $post[0];

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

$do = $_POST['do'];

if($do == 'new') {
	post_new_post();
} else if($do == 'update') {
	post_update();
}

die(0);

endif; // POST

