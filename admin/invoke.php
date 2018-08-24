<?php

function Invoke($path, $type='json', $body=null, $post=true) {
    $url = 'http://127.0.0.1:2564/v1'.$path;
    $type_header = 'text/plain';
    if ($type === 'json') {
        $type_header = 'application/json';
    } else if ($type === 'form') {
        $type_header = 'application/x-www-form-urlencoded';
    }

    $auth = $_COOKIE["login"] ?? '';
    $ua = $_SERVER['HTTP_USER_AGENT'];

    // use key 'http' even if you send the request to https://...
    $options = array(
        'http' => array(
            'header'  => "User-Agent: $ua\r\nContent-type: ".$type_header."\r\nCookie: login=".$auth."\r\n",
            'method'  => $post ? 'POST':'GET',
            'content' => $body,
        )
    );
    $context  = stream_context_create($options);
    return file_get_contents($url, false, $context);
}
