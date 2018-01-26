<?php

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once('admin.php');

function postmanage_admin_head() {
?>

<style>

table, td, th {
    border: 1px solid gray;
    border-collapse: collapse;
}

th, td {
    padding: 8px;
}

td.title {
    max-width: 300px;
}

</style>

<?php }

add_hook('admin_head', 'postmanage_admin_head');

admin_header();

?>

<table>
<thead>
<tr>
<th>编号</th>
<th>标题</th>
<th>发表日期</th>
<th>修改日期</th>
<th>浏览量</th>
<th>评论数</th>
<th>源类型</th>
<th>操作</th>
</tr>
</thead>
<tbody>
<?php
$posts = $tbpost->get_all_posts_for_manage();
foreach($posts as $p) {
    echo '<tr>',
        '<td>',$p->id,'</td>',
        '<td class="title">',htmlspecialchars($p->title),'</td>',
        '<td>',$p->date,'</td>',
        '<td>',$p->modified,'</td>',
        '<td>',$p->page_view,'</td>',
        '<td>',$p->comments,'</td>',
        '<td>',htmlspecialchars($p->source_type),'</td>',
        '<td>',the_edit_link($p, true, true),'</td>',
        '</tr>'
    ;
}
?>
</tbody>
</table>

<?php
admin_footer();

die(0);

else : // POST

die(0);

endif;

