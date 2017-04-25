<?php
	$the = $tbquery->the();

	if(!preg_match('/[-]000/', $the->modified) && !$logged_in){
		header('Last-Modified: '.$tbdate->mysql_local_to_http_gmt($the->modified));
	}

    // 由于博客程序经常修改，所以光靠文章的修改日期来定304不完全靠谱
    // 所以这里生成一个简单的实体标签，基于：博客标签+修改时间时间戳
    if(!$logged_in)
        header('Etag: "'. TB_VERSION .'-'.$tbdate->mysql_local_to_timestamp($the->modified).'"');

	if($logged_in) {
		header('Cache-Control: private');
	}

require('header.php');
require('content.php');
require('footer.php');

