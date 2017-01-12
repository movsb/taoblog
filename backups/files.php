<?php
	$whats = ['file', 'style', 'common'];
	$type = '';
	$file = '';
	foreach($whats as $what){
		if(isset($_GET[$what])){
			$type = $what;
			$file = $_GET[$type];
			break;
		}
	}

	if($type == ''){
		exit;
	}

	$hdr = array();

	function header_callback($ch, $str){
		global $hdr;

		if(count($hdr) == 0){
			//$hdr['Status'] = preg_split('/\s/',trim($str));
			$hdr['Status'] = trim($str);
		}
		else if(strlen($str) <= 2){
			$h = preg_split('/\s/', $hdr['Status'], 3);
			if($h[1] !=='200'){
				header($hdr['Status']);
				echo '<h1>'.$h[1].' '. $h[2].'</h1>';
				return 0;
			}
			else{
				write_header();
			}
		}
		else{
			$a = preg_split('/:/', $str, 2);
			$a[1] = trim($a[1]);
			$hdr[$a[0]] = $a[1];
		}
		return strlen($str);
	}
	
	function write_header(){
		global $hdr;
		header($hdr['Status']);

		$wanthdr = array(
			'Last-Modified',
			'ETag',
			'Content-Length',
			'Content-Type'
		);
		foreach($wanthdr as $h){
			if(isset($hdr[$h])){
				header($h . ': ' . $hdr[$h]);
			}
		}
	}

	$osshost = 'twofei-wordpress.oss-cn-hangzhou.aliyuncs.com';
	if($type === 'file') $osshost .= '/files/';
	else if($type === 'style') $osshost .= '/styles/';
	else if($type === 'common') $osshost .= '/common/';

	$ch = curl_init($osshost . $file);
	curl_setopt($ch,CURLOPT_HEADER, false);
	curl_setopt($ch, CURLOPT_HEADERFUNCTION, 'header_callback');

	$req_hdr = [
		['HTTP_USER_AGENT','User-Agent'],
		['HTTP_IF_NONE_MATCH','If-None-Match'],
		['HTTP_IF_MODIFIED_SINCE','If-Modified-Since'],
	];
	$req_hdr_real = array();
	foreach($req_hdr as $h){
		if(isset($_SERVER[$h[0]])){
			$req_hdr_real[count($req_hdr_real)] = $h[1] . ': ' . $_SERVER[$h[0]];
		}
	}
	curl_setopt($ch, CURLOPT_HTTPHEADER, $req_hdr_real);
	curl_exec($ch);
	curl_close($ch);
	die(0);

