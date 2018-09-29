<?php 

define('TB_VERSION', '1.1.12');

require_once dirname(__FILE__).'/../setup/config.php';

require_once 'utils/die.php';
require_once 'utils/utils.php';
require_once 'utils/datetime.php';

require_once 'login-auth.php';

require_once 'utils/hooks.php';
require_once 'plugin.php';

require_once 'models/base.php';
require_once 'models/options.php';
require_once 'models/posts.php';
require_once 'models/taxonomies.php';
require_once 'models/tags.php';

require_once 'query.php';
require_once 'canonical.php';

require_once 'invoke.php';

apply_hooks('tb_load');

$tbopt          = new TB_Options;
$tbpost         = new TB_Posts;
$tbtax          = new TB_Taxonomies;
$tbquery        = new TB_Query;
$tbdate         = new TB_DateTime;
$tbtag          = new TB_Tags;

date_default_timezone_set('Asia/Chongqing');

$logged_in = login_auth();
