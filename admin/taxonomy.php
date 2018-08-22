<?php

$admin_url = 'taxonomy.php';

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once('admin.php');

function new_tax_html() { ?>
<style>
    .tax-lv-0 {
        padding-left: 0.5em;
    }

    .tax-lv-1 {
        padding-left: 2em;
    }

    .tax-lv-2 {
        padding-left: 4em;
    }

    .tax-lv-3 {
        padding-left: 6em;
    }

    .tax-lv-4 {
        padding-left: 8em;
    }

    #parent option {
        padding-top: 0.3em;
        padding-bottom: 0.3em;
    }

    #tax-list tbody tr {
        height: 60px;
    }

    #tax-list tbody td {
        vertical-align: top;
    }


</style>

<div id="new-tax" style="float: left; padding: 1em; background-color: orange; box-shadow: 2px 2px 4px #E07080;">
<form method="post" id="new-tax-form">
    <div>
        <span>名字: </span>
        <input type="text" name="name" />
    </div>
    <div>
        <span>别名: </span>
        <input type="text" name="slug" />
    </div>
    <div>
        <span>父类: </span>
        <select id="parent" name="parent">
            <option class="tax-lv-0" id="tax-id-0" lv="0" value="0">--- 无 ---</option>
        </select>
    </div>
    <div style="margin: 1em; text-align: right;">
        <span id="add-new-tax" type="submit" class="btn" style="background-color: purple; color: white; padding: 8px 12px; cursor: pointer;">增加</span>
    </div>
    <div style="display: none;">
        <input type="hidden" name="do" value="add-new" />
    </div>
    <script><?php
        global $tbtax;
        echo 'var taxes = JSON.parse(\''.json_encode($tbtax->get_hierarchically(), JSON_UNESCAPED_UNICODE).'\');';
        ?>
    </script>
</form>
</div>
<div id="tax-list-div" style="margin-left: 450px; padding: 1em; background-color: palevioletred; box-shadow: 2px 2px 4px purple;">
<table id="tax-list" style="width: 100%;">
    <thead style="text-align: left;">
        <tr>
            <th>名字</th>
            <th>别名</th>
        </tr>
    </thead>
    <tfoot style="text-align: left;">
        <tr>
            <th>名字</th>
            <th>别名</th>
        </tr>
    </tfoot>
    <tbody>
        <tr id="edit-tax" style="display: none;" colspan="2">
            <td>
                <p>编辑分类</p>
                <form>
                    <span>名字：</span><input name="name" type="text"><br>
                    <span>别名：</span><input name="slug" type="text">
                    <div>
                        <button class="cancel">取消</button>
                        <input class="submit" type="submit" value="更新" />
                    </div>
                    <input type="hidden" name="do" value="" />
                    <input type="hidden" name="id" value="" />
                    <!--input type="hidden" name="parent" value="" /-->
                </form>
            </td>
        </tr>
        <tr id="tax-list-0" style="display: none;"><td></td><td></td></tr>
    </tbody>
</table>
<script src="scripts/taxonomy.js">
</script>
</div>
<?php }

admin_header();

new_tax_html();

admin_footer();


else : // POST

function tax_die_json($arg) {
    header('HTTP/1.1 200 OK');
    header('Content-Type: application/json');

    echo json_encode($arg, JSON_UNESCAPED_UNICODE);
    die(0);
}

require_once('login-auth.php');

if(!login_auth()) {
    tax_die_json([
        'errno' => 'unauthorized',
        'error' => '需要登录后才能进行该操作！',
        ]);
}

require_once('load.php');

function tax_new_tax() {
    

}

function tax_get_all() {
    echo json_encode($tbtax->get(), JSON_UNESCAPED_UNICODE);
    die(0);
}

function tax_update(&$arg) {
    global $tbtax;

    // TODO removed
    $r = $tbtax->update($arg);
    if(!$r) {
        tax_die_json([
            'errno' => 'error',
            'error' => $tbtax->error
            ]);
    }

    tax_die_json([
        'errno' => 'success'
        ]);
}

function tax_delete(&$arg) {
    global $tbtax;

    tax_die_json([
        'errno' => 'error',
        'error' => '当前不支持分类的删除。',
        ]);
}

$do = $_POST['do'] ?? '';
if($do === 'get-all') {
    tax_get_all();
} else if($do === 'add-new') {
    $id = $tbtax->add($_POST);
    if($id === false) {
        echo json_encode([
            'errno' => 'error',
            'error' => $tbtax->error,
            ], JSON_UNESCAPED_UNICODE);
        die(0);
    } else {
        echo json_encode([
            'errno' => 'success',
            'tax' => $tbtax->get($id),
            ], JSON_UNESCAPED_UNICODE);
        die(0);
    }
} else if($do === 'update') {
    tax_update($_POST);
} else if($do === 'delete') {
    tax_delete($_POST);
}


endif;

