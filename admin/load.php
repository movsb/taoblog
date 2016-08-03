<?php 

define('TB_VERSION', '1.0.3');

require_once(dirname(__FILE__).'/../setup/config.php');

require_once('die.php');
require_once('utils.php');
require_once('datetime.php');

require_once('login-auth.php');

require_once('hooks.php');
require_once('plugin.php');

require_once('models/base.php');
require_once('models/options.php');
require_once('models/posts.php');
require_once('models/taxonomies.php');
require_once('models/comments.php');
require_once('models/tags.php');
require_once('models/shuoshuo.php');

require_once('tbmain.php');

require_once('models-hooks.php');
require_once('global-hooks.php');

require_once('query.php');
require_once('canonical.php');

apply_hooks('tb_load');

$tbopt          = new TB_Options;
$tbpost         = new TB_Posts;
$tbtax          = new TB_Taxonomies;
$tbcmts         = new TB_Comments;
$tbquery        = new TB_Query;
$tbdate         = new TB_DateTime;
$tbtag          = new TB_Tags;
$tbshuoshuo     = new TB_Shuoshuo;
$tbsscmt        = new TB_ShuoshuoComments;

$tbmain         = new TB_Main;

date_default_timezone_set('Asia/Chongqing');

$logged_in = login_auth();

