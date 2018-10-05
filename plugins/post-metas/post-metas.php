<?php

class TB_Post_Metas {
    public $error = '';
    public $allowed_types = ['post', 'page'];

    public function has($tid, $type = 'post') {
        global $tbdb;

        if(!in_array($type, $this->allowed_types)) {
            $this->error = '未知的meta类型';
            return false;
        }

        $tid = (int)$tid;
        $sql = "SELECT id FROM post_metas WHERE type='$type' AND tid=$tid LIMIT 1";
        $r = $tbdb->query($sql);
        if(!$r || !is_a($r, 'mysqli_result')) {
            $this->error = $tbdb->error;
            return false;
        }

        return $r->num_rows > 0;
    }

    public function get($tid, $type = 'post') {
        global $tbdb;

        if(!in_array($type, $this->allowed_types)) {
            $this->error = '未知的meta类型';
            return false;
        }

        if(!$this->has($tid, $type)) {
            $this->error = "类型为 $type 的 $tid 不存在";
            return false;
        }

        $tid = (int)$tid;

        $metas = new stdClass;

        $sql = "SELECT * FROM post_metas WHERE type='$type' AND tid=$tid LIMIT 1";
        if(($r = $tbdb->query($sql)) && is_a($r, 'mysqli_result') && $r->num_rows>0) {
            $metas = $r->fetch_object();
        }

        return $metas;
    }
}

function postmetas_head() {
    global $tbquery;

    if(!$tbquery->is_singular() || !$tbquery->count) {
        return false;
    }

    $tbpm = new TB_Post_Metas;

    global $the; // the post/page object

    $id = $the->id;
    $type = $the->type;

    $ispost = $type === 'post';

    $post_metas = $tbpm->get($id, $type);

    $metas = new stdClass;
    $metas->post = $post_metas;

    $GLOBALS['postmetas'] = $metas;

    $keywords = '';

    // 输出关键字
    if($metas->post) {
        $keywords_post = $metas->post ? $metas->post->keywords : '';
        if($keywords_post) $keywords = $keywords_post . $keywords;
    }
    echo '	<meta name="keywords" content="',htmlspecialchars($keywords),',',htmlspecialchars(get_opt('blog_name')),'" />'."\n";

    if($metas->post) echo $metas->post->header;
}

add_hook('tb_head', 'postmetas_head');

function postmetas_footer() {
    if(!isset($GLOBALS['postmetas'])) {
        // 如果head没被成功调用
        return;
    }
    $metas = $GLOBALS['postmetas'];

    if($metas->post) echo $metas->post->footer;

    unset($GLOBALS['postmetas']);
}

add_hook('tb_footer', 'postmetas_footer');
