<?php

if($_SERVER['REQUEST_METHOD'] == 'GET') :




else :

require_once('load.php');

function cmt_make_public(&$cmts) {
    global $tbopt;
    global $logged_in;
    $admin_email = $tbopt->get('email');

    $flts = !$logged_in ? ['email', 'ip'] : [];

    for($i=0,$c=count($cmts); $i < $c; $i++) {
        $cmt = $cmts[$i];
        $cmt->avatar = md5(strtolower($cmt->email));
        $cmt->is_admin = strcasecmp($cmt->email, $admin_email)==0;
        foreach($flts as &$f) unset($cmt->$f);

        if(isset($cmt->children)) {
            for($x=0,$xc=count($cmt->children); $x < $xc; $x++) {
                $child = $cmt->children[$x];
                $child->avatar = md5(strtolower($child->email));
                $child->is_admin = strcasecmp($child->email, $admin_email)==0;
                foreach($flts as &$f) unset($child->$f);
            }
        }
    }
    return $cmts;
}

function cmt_post_cmt() {
    global $tbcmts;
    global $tbdb;

    $ret_cmt = (int)($_POST['return_cmt'] ?? '');
    
    $r = $tbcmts->insert($_POST);
    if(!$r) {
        header('HTTP/1.1 200 OK');
        header('Content-Type: application/json');

        echo json_encode([
            'errno' => 'error',
            'error' => $tbcmts->error
            ], JSON_UNESCAPED_UNICODE);
        die(-1);
    }

    header('HTTP/1.1 200 OK');
    header('Content-Type: application/json');

    ob_start();
    if($ret_cmt) {
        $c = ['id'=>$r];
        $cmts = cmt_make_public($tbcmts->get($c));

        echo json_encode([
            'errno' => 'success',
            'cmt'   => $cmts[0],
            ], JSON_UNESCAPED_UNICODE);
    } else {
        echo json_encode([
            'errno' => 'success',
            'id'    => $r,
            ], JSON_UNESCAPED_UNICODE);
    }
    header('Content-Length: '.ob_get_length());
    header('Connection: close');
    ob_end_flush();
    fastcgi_finish_request();

    apply_hooks('comment_posted', 0, $_POST);
    die(0);
}

$do = $_POST['do'] ?? '';

if($do == 'post-cmt')  cmt_post_cmt();

endif;

