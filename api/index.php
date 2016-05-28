<?php

$api = new stdClass();

require_once(dirname(__FILE__).'/../admin/load.php');

$api->loggedin = $logged_in;

function api_die($arr) {
    header('HTTP/1.1 200 OK');
    header('Content-Type: application/json');
    echo json_encode($arr);
    die(0);
}

function api_bad_method() {
    global $api;

    api_die([
        'ret' => -1,
        'msg' => "unknown method `{$api->method}' of module `{$api->module}'.",
    ]);
}

if(preg_match('~^/api/([^/]+)/([^/?]+)~', $_SERVER['REQUEST_URI'], $matches)) {
    $api->module = $matches[1];
    $api->method = $matches[2];

    require_once($api->module.'.php');
    die(0);
}
else {
    die('bad request.');
}

