<?php

require_once(dirname(__FILE__).'/../admin/load.php');


if(preg_match('~^/api/([^/]+)/([^/?]+)~', $_SERVER['REQUEST_URI'], $matches)) {
    $tbapi->init($matches[1], $matches[2]);
    include_once($tbapi->module.'.php');
}
else {
    $tbapi->err(-1, 'bad request.');
}

die(0);

