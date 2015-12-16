<?php

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once('admin.php');

admin_header();

$id = isset($_GET['id']) ? (int)$_GET['id'] : 0;
$content = $id > 0 ? $tbshuoshuo->get($id) : '';
?>
<form method="post" style="margin-bottom: 2em;">
<h2>发表说说</h2>
<textarea name="content" style="display: block; min-width: 400px; min-height: 150px;"><?php echo $content;?></textarea>
<input type="submit" value="发表" />
<input type="hidden" name="do" value="<?php echo $id > 0 ? 'update' : 'new';?>" />
<input type="hidden" name="id" value="<?php echo $id;?>" />
</form>

<h2>近期说说</h2>
<?php
    $sss = $tbshuoshuo->get_latest(10);
    if(count($sss) == 0) return false;

    echo '<ul id="shuoshuos" style="list-style: none;">';
    foreach($sss as &$ss) {
        echo '<li data-id="',$ss->id,'"><p>',$ss->content,'</p>','<span>',$ss->date,'</span>';
        echo '<button class="edit">编辑</button><button class="delete">删除</button></li>';
    }
    echo '</ul>';
?>
<script>
    $('#shuoshuos').on('click', function(e) {
        var cls = e.target.classList;
        if(cls.contains('edit')) {
            var id = $(e.target.parentNode).attr('data-id');
            location.href = '/admin/shuoshuo.php?id=' + id;
            e.preventDefault();
            e.stopPropagation();
            return false;
        }
        else if(cls.contains('delete')) {
            var id = $(e.target.parentNode).attr('data-id');
            if(!confirm('确定删除吗？'))
                return false;

            $.post('/admin/shuoshuo.php',
                'do=delete&id=' + id,
                function(data) {
                    if(data.errno == 'ok') {
                        $('#shuoshuos > li[data-id=' + id + ']').remove();
                    }
                    else {
                        alert(data.error);
                    }
                },
                'json'
            );
        }
    });
</script>

<?php
admin_footer();

die(0);

else : // POST

function shuoshuo_die_json($arg) {
	header('HTTP/1.1 200 OK');
	header('Content-Type: application/json');

	echo json_encode($arg);
	die(0);
}

require_once('login-auth.php');

if(!login_auth()) {
	shuoshuo_die_json([
		'errno' => 'unauthorized',
		'error' => '需要登录后才能进行该操作！',
		]);
}

require_once('load.php');


$do = isset($_POST['do']) ? $_POST['do'] : '';

if($do == 'new') {
    $r = $tbshuoshuo->post($_POST['content']);
    if($r === false)
        shuoshuo_die_json([
            'errno' => 'failed',
            'error' => 'failed',
            ]);

    header('HTTP/1.1 302 Found');
    header('Location: /admin/shuoshuo.php');
    die(0);
}
else if($do == 'update') {
    $r = $tbshuoshuo->update((int)$_POST['id'], $_POST['content']);
    if($r === false)
        shuoshuo_die_json([
            'errno' => 'failed',
            'error' => 'failed',
            ]);

    header('HTTP/1.1 302 Found');
    header('Location: /admin/shuoshuo.php');
    die(0);
}
else if($do == 'delete') {
    $r = $tbshuoshuo->del((int)$_POST['id']);
    if($r === false)
        shuoshuo_die_json([
            'errno' => 'failed',
            'error' => 'failed',
            ]);

    else
        shuoshuo_die_json([
            'errno' => 'ok',
            ]);
    die(0);
}

die(0);

endif;

