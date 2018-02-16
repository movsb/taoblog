<?php

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once 'admin.php';

function tagmanage_admin_head()
{
?>

<style>

table, td, th {
    border: 1px solid gray;
    border-collapse: collapse;
}

th, td {
    padding: 8px;
}

td.name {
    max-width: 300px;
}

#edit-box {
    display: none;
    position: fixed;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    background-color: gray;
}

</style>

<?php }

add_hook('admin_head', 'tagmanage_admin_head');

admin_header();

?>

<table class="table">
<thead>
<tr>
<th>编号</th>
<th>名字</th>
<th>别名</th>
<th>文章数</th>
<th>编辑</th>
</tr>
</thead>
<tbody>
<?php
$tags = $tbtag->list_all_tags(-1, false);
foreach ($tags as $t) {
    echo '<tr data-id="',$t->id,'">',
        '<td>',$t->id,'</td>',
        '<td class="name">',htmlspecialchars($t->name),'</td>',
        '<td class="alias">',$t->alias,'</td>',
        '<td>',$t->size,'</td>',
        '<td><button class="edit">编辑</button></td>',
        '</tr>'
    ;
}
?>
</tbody>
</table>

<script>
var cur_edit_tr = null;
$('.table').click(function(e) {
    if(e.target.className == "edit") {
        var tr = $(e.target).parent().parent();
        var id = tr.attr('data-id');
        var name = tr.find('.name');
        var alias = tr.find('.alias');
        cur_edit_tr = tr;
        showEditBox(id, name.text(), alias.text());
    }
});
</script>

<div id="edit-box" class="admin-wrap">
<form id="edit-form">
    <div>
        <label for="idbox">编号</label>
        <input name="idbox" type="text" disabled value="" />
        <input name="id" type="hidden" value="" />
    </div>
    <div>
        <label for="name">名字</label>
        <input name="name" type="text" value="" />
    </div>
    <div>
        <label for="alias">别名</label>
        <input name="alias" type="text" value="" />
    </div>
    <div>
        <input type="hidden" name="do" value="update" />
        <input class="submit" type="submit" value="保存" />
        <input class="cancel" type="button" value="取消" />
    </div>
</form>
<script>
$('#edit-form .submit').click(function(){
    $.post('', $('#edit-form').serialize(),
        function(data) {
            if (data.errno != 'ok') {
                alert(data.error);
                return;
            }
            var f = $('#edit-box');
            var t = cur_edit_tr;
            t.find('.name').text(f.find('input[name=name]').val());
            t.find('.alias').text(f.find('input[name=alias]').val());
            f.fadeOut();
        },
        'json'
    );
    return false;
});
$('#edit-form .cancel').click(function() {
    $('#edit-box').fadeOut();
});
function showEditBox(id, name, alias) {
    var f = $('#edit-box');
    f.find('input[name=idbox]').val(id);
    f.find('input[name=id]').val(id);
    f.find('input[name=name]').val(name);
    f.find('input[name=alias]').val(alias);
    f.fadeIn();
}
</script>
</div>

<?php
admin_footer();

die(0);

else : // POST

function tag_die_json($arg)
{
    header('HTTP/1.1 200 OK');
    header('Content-Type: application/json');

    echo json_encode($arg, JSON_UNESCAPED_UNICODE);
    die(0);
}

require_once 'login-auth.php';

if (!login_auth()) {
    tag_die_json([
        'errno' => 'unauthorized',
        'error' => '需要登录后才能进行该操作！',
        ]);
}

require_once('load.php');

$do = $_POST['do'] ?? '';
if ($do == 'update') {
    $id = (int)$_POST['id'];
    $name = (string)$_POST['name'];
    $alias = (int)$_POST['alias'];

    $r = $tbtag->updateTag($id, $name, $alias);

    tag_die_json([
        'errno' => $r ? 'ok' : 'error',
        'error' => $tbtag->error,
    ]);
}

endif;

