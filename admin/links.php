<?php

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once('admin.php');

admin_header();

global $tbopt;

$links = $tbopt->get('links');
if(empty($links)) $links = '[]';?>
<script type="text/javascript">
	var links = JSON.parse('<?php echo $links; ?>');
</script>
<div class="link-list" id="link-list">
<h2>已有链接</h2>
<ul style="list-style: none;">
</ul>
<div class="add-new-link" id="add-new-link" style="margin-top: 1em;">
	<h2>新增</h2>
	<label for="name">名字</label>&nbsp;<input type="text" name="name" value="" />
	<label for="name">说明</label>&nbsp;<input type="text" name="title" value="" />
	<label for="href">链接</label>&nbsp;<input type="text" name="href" value="" />
	<button class="add append">增加到后面</button><button class="add prepend">增加到前面</button>
	<script type="text/javascript">
		$('#add-new-link').click(function(e) {
			if(e.target.classList.contains('add')) {
				var name = $('#add-new-link input[name=name]').val();
				var title = $('#add-new-link input[name=title]').val();
				var href = $('#add-new-link input[name=href]').val();

				if(e.target.classList.contains('append')) add_new_link(name, title, href, 1);
				if(e.target.classList.contains('prepend')) add_new_link(name, title, href, 0);
			}
		});
	</script>
</div>
<script type="text/javascript">
	$('#link-list').click(function(e){
		if(e.target.classList.contains('del-link')) {
			var li = e.target.parentNode;
			$(li).remove();
		}
	});

	function add_new_link(name, title, href, after) {
		var html = '<li>';
		html += '<label for="name">名字</label>&nbsp;<input type="text" name="name" value="'+ name +'"/>&nbsp&nbsp;';
		html += '<label for="title">说明</label>&nbsp;<input type="text" name="title" value="'+ title +'"/>&nbsp&nbsp;';
		html += '<label for="href">链接</label>&nbsp;<input type="text" name="href" value="'+ href +'"/>';
		html += '&nbsp;&nbsp&nbsp;<button class="del-link">删除</button>';
		html += '</li>';

		if(after) $('#link-list ul').append(html);
		else $('#link-list ul').prepend(html);
	}

	for(var i=0; i<links.length; i++) {
		add_new_link(links[i].name, links[i].title, links[i].href, 1);
	}

</script>
</div>
<form method="post" style="margin-top: 3em;" id="form">
<div>
	<input type="submit" id="submit" value="保存修改" />
	<input type="hidden" name="do" value="update" />
	<input type="hidden" name="links" value="" />

	<script type="text/javascript">
		$('#submit').click(function() {
			var links = [];
			var lis = $('#link-list ul')[0].children;
			for(var i=0; i<lis.length; i++) {
				var li = lis[i];
				var name = $(li).find('input[name=name]').val();
				var title = $(li).find('input[name=title]').val();
				var href = $(li).find('input[name=href]').val();
				links.push({
					name: name,
					title: title,
					href: href
				});
			}

			links = JSON.stringify(links);
			$('#form input[name=links]').val(links);
			return true;
		});
	</script>
</div>
</form>
<?php

admin_footer();

die(0);

else : // POST

function links_die_json($arg) {
	header('HTTP/1.1 200 OK');
	header('Content-Type: application/json');

	echo json_encode($arg);
	die(0);
}

require_once('login-auth.php');

if(!login_auth()) {
	links_die_json([
		'errno' => 'unauthorized',
		'error' => '需要登录后才能进行该操作！',
		]);
}

require_once('load.php');


$do = isset($_POST['do']) ? $_POST['do'] : '';

if($do == 'update') {
	$tbopt->set('links', $_POST['links']);

	header('HTTP/1.1 302 OK');
	header('Location: /admin/links.php');
}

die(0);

endif;

