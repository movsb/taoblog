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

</style>

<div id="new-tax" style="float: left; width: 400px; padding: 1em; background-color: orange; box-shadow: 2px 2px 4px #E07080;">
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
	<div style="margin: 1em;">
		<span id="add-new-tax" type="submit" class="btn" style="background-color: purple; padding: 1em;">增加</span>
	</div>
	<div style="display: none;">
		<input type="hidden" name="do" value="add-new" />
	</div>
	<script><?php
		global $tbtax;
		echo 'var taxes = JSON.parse(\''.json_encode($tbtax->get()).'\');';
		?>
	</script>
</form>
</div>
<div id="tax-list" style="margin-left: 450px; padding: 1em; background-color: palevioletred; box-shadow: 2px 2px 4px purple;">
<table style="width: 100%;">
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
		<tr id="tax-list-0" style="display: none;"><td></td><td></td></tr>
	</tbody>
</table>
<script src="ta.js">
</script>
</div>
<?php }

admin_header();

new_tax_html();

admin_footer();


else : // POST

function tax_new_tax() {
	

}

function tax_get_all() {
	require_once(dirname(__FILE__).'/../setup/config.php');
	require_once('db/dbbase.php');
	require_once('db/taxonomies.php');

	$GLOBALS['tbdb'] = $tbdb;

	$tbtax = new TB_Taxonomies;
	echo json_encode($tbtax->get());
	die(0);
}

$do = isset($_POST['do']) ? $_POST['do'] : '';
if($do === 'get-all') {
	tax_get_all();
} else if($do === 'add-new') {
	require_once(dirname(__FILE__).'/../setup/config.php');
	require_once('db/dbbase.php');
	require_once('db/taxonomies.php');

	$GLOBALS['tbdb'] = $tbdb;

	$tbtax = new TB_Taxonomies;
	$name = $_POST['name'];
	$slug = $_POST['slug'];
	$parent = $_POST['parent'];
	$ancestor = $tbtax->get_ancestor($parent);
	$id = $tbtax->add($name, $slug, $parent, $ancestor);
	if($id === false) {
		echo json_encode([
			'errno' => 'error',
			'error' => '添加分类失败！'
			]);
		die(0);
	} else {
		echo json_encode([
			'errno' => 'success',
			'tax' => $tbtax->get($id),
			]);
		die(0);
	}
}


endif;

