<?php

function Invoke($path, $type, $body) {
    $url = 'http://127.0.0.1:2564/v1'.$path;
    $type_header = 'text/plain';
    if ($type === 'json') {
        $type_header = 'application/json';
    } else if ($type === 'form') {
        $type_header = 'application/x-www-form-urlencoded';
    }

    $auth = $_COOKIE["login"];
    $ua = $_SERVER['HTTP_USER_AGENT'];

    // use key 'http' even if you send the request to https://...
    $options = array(
        'http' => array(
            'header'  => "User-Agent: $ua\r\nContent-type: ".$type_header."\r\nCookie: login=".$auth."\r\n",
            'method'  => 'POST',
            'content' => $body,
        )
    );
    $context  = stream_context_create($options);
    $result = file_get_contents($url, false, $context);
}
