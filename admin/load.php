<?php 

require_once(dirname(__FILE__).'/../setup/config.php');

require_once('die.php');
require_once('utils.php');

require_once('hooks.php');

require_once('db/dbbase.php');
require_once('db/options.php');
require_once('db/posts.php');
require_once('db/taxonomies.php');
require_once('db/comments.php');

require_once('query.php');

$tbopt = new TB_Options;
$tbpost = new TB_Posts;
$tbtax = new TB_Taxonomies;
$tbcmts = new TB_Comments;
$tbquery = new TB_Query;

require_once(dirname(__FILE__).'/../plugins/content-from-file/content-from-file.php');

