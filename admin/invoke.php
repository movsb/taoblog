<?php

function Invoke($path, $type='json', $body=null, $post=true, $ver='/v1') {
    $url = 'http://127.0.0.1:'.TB_API_PORT.$ver.$path;
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

function get_tax_tree() {
    $cats = Invoke('/categories!tree', 'json', null, false);
    return json_decode($cats);
}

function get_all_tags(int $limit, bool $merge) {
    $tags = Invoke('/tags!count?limit='.$limit.'&merge='.($merge?1:0), 'json', null, false);
    return json_decode($tags);
}

function get_opt(string $name, string $def='') {
    $value = Invoke('/options/'.urlencode($name), 'json', null, false, '/.v1');
    $value = json_decode($value);
    return $value ?? $def;
}

function get_date_archives() {
    $d = Invoke('/archives/dates', 'json', null, false);
    $d = json_decode($d);
    return $d;
}
