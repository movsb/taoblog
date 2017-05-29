<?php

if($_SERVER['REQUEST_METHOD'] == 'GET') :




else :

function fm_die_json($arg) {
	header('HTTP/1.1 200 OK');
	header('Content-Type: application/json');

	echo json_encode($arg, JSON_UNESCAPED_UNICODE);
	die(0);
}

function fm_error($msg) {
    fm_die_json([
        'errno' => 'error',
        'error' => $msg,
        ]);
}

require_once('login-auth.php');

if(!login_auth()) {
	fm_error('需要登录后才能进行该操作！');
}

require_once('load.php');

function fm_upload() {
    $pid = (int)($_POST['pid'] ?? 0);
    if($pid <= 0) {
        fm_error('无效文章编号。');
    }

    $root = TBPATH.'/'.FILE_DIR.'/'.$pid;
    if(!is_dir($root)) {
        if(!@mkdir($root, 0777, true)) {
           fm_error('无法创建上传文件保存目录。');
        }
    }

    $count = 0;

    foreach($_FILES['files']['error'] as $index => $error) {
        $count++;
        if($error === 0) {
            $tmp_name = $_FILES['files']['tmp_name'][$index];
            $name = basename($_FILES['files']['name'][$index]);

            $path = "$root/$name";

            if(@move_uploaded_file($tmp_name, $path)) {
                $count--;
            }
        }
    }

    if($count != 0) {

    }

    fm_die_json([
        'errno' => 'ok',
    ]);
}

function fm_list() {
    $pid = (int)($_POST['pid'] ?? 0);
    $root = TBPATH.'/'.FILE_DIR.'/'.$pid;
    
    $files = [];

    if($handle = opendir($root)) {
        while(($file = readdir($handle)) !== false) {
            if($file[0] != '.') {
                $files[] = $file;
            }
        }
    }

    fm_die_json([
        'errno' => 'ok',
        'files' => $files,
    ]);
}

function fm_delete() {
    $pid = (int)($_POST['pid'] ?? 0);
    $root = TBPATH.'/'.FILE_DIR.'/'.$pid;
    $name = basename($_POST['name'] ?? '');
    $path = "$root/$name";

    $R = @unlink($path);

    fm_die_json([
        'errno' => $R ? 'ok' : 'error',
    ]);
}

$do = $_POST['do'] ?? '';

if($do === 'upload') fm_upload();
else if($do === 'list') fm_list();
else if($do === 'delete') fm_delete();

else error('error');

endif;

