<?php 

define('TB_VERSION', '1.1.13');

require_once dirname(__FILE__).'/../setup/config.php';

require_once 'utils/die.php';
require_once 'utils/datetime.php';

require_once 'utils/hooks.php';

require_once 'models/base.php';
require_once 'models/posts.php';

require_once 'canonical.php';

require_once 'invoke.php';

apply_hooks('tb_load');

$tbpost         = new TB_Posts;
$tbquery        = new TB_Query;
$tbdate         = new TB_DateTime;

date_default_timezone_set('Asia/Chongqing');

$logged_in = login_auth();
