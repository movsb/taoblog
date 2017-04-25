<?php

$path = urldecode($_SERVER['QUERY_STRING']);
$url = 'http://www.gravatar.com/avatar/'.$path;

$headers = [];
if(isset($_SERVER['HTTP_IF_MODIFIED_SINCE']))
	$headers[] = 'If-Modified-Since: '.$_SERVER['HTTP_IF_MODIFIED_SINCE'];
if(isset($_SERVER['HTTP_IF_NONE_MATCH']))
	$headers[] = 'If-None-Match: '.$_SERVER['HTTP_IF_NONE_MATCH'];

$ch = curl_init($url);
if(count($headers)) {
	curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
	$headers = [];
}

curl_setopt($ch, CURLOPT_HEADERFUNCTION,
	function($handle, $line) use (&$headers) {
		$headers[] = substr($line, 0, -2);
		if(strlen($line) === 2) {
			for($i=0,$c=count($headers)-1; $i<$c; $i++) {
				header($headers[$i]);
			}
		}
		return strlen($line);
	});

curl_exec($ch);
curl_close($ch);

