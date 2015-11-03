<?php 

require_once(dirname(__FILE__).'/../setup/config.php');

require_once('die.php');
require_once('utils.php');
require_once('datetime.php');

require_once('login-auth.php');

require_once('hooks.php');
require_once('plugin.php');

require_once('db/dbbase.php');
require_once('db/options.php');
require_once('db/posts.php');
require_once('db/taxonomies.php');
require_once('db/comments.php');
require_once('db/tags.php');

require_once('db-hooks.php');

require_once('query.php');
require_once('canonical.php');

apply_hooks('tb_load');

$tbopt = new TB_Options;
$tbpost = new TB_Posts;
$tbtax = new TB_Taxonomies;
$tbcmts = new TB_Comments;
$tbquery = new TB_Query;
$tbdate = new TB_DateTime;
$tbtag = new TB_Tags;

date_default_timezone_set('Asia/Chongqing');

$logged_in = login_auth_cookie();

