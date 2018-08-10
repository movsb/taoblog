<?php

class TB_Post_Metas {
    public $error = '';
    public $allowed_types = ['post', 'page', 'tax'];

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
    global $tbtax;
    global $tbopt;

    if(!$tbquery->is_singular() || !$tbquery->count) {
        return false;
    }

    $tbpm = new TB_Post_Metas;

    global $the; // the post/page object

    $id = $the->id;
    $type = $the->type;
    $tax = $the->taxonomy;

    $ispost = $type === 'post';

    if($ispost) {
        $parents_ids = $tbtax->get_parents_ids($tax);
        $tax_metas = [];

        foreach($parents_ids as $pid) {
            $t = $tbpm->get($pid, 'tax');
            if($t) $tax_metas[] = $t;
        }
        $this_tax_meta = $tbpm->get($tax, 'tax');
        if($this_tax_meta) $tax_metas[] = $this_tax_meta;
    }

    $post_metas = $tbpm->get($id, $type);

    $metas = new stdClass;
    if($ispost)
        $metas->tax = &$tax_metas;
    $metas->post = $post_metas;

    $GLOBALS['postmetas'] = $metas;

    $keywords = '';

    if($ispost) {
        // TODO: 优化到post一起
        $names = array_reverse($tbtax->tree_from_id($tax)['name']);
        if($this_tax_meta && $this_tax_meta->keywords) unset($names[0]);

        $keywords = implode(',', $names);
        if($this_tax_meta && $this_tax_meta->keywords) $keywords = $this_tax_meta->keywords . $keywords;
    }

    // 输出关键字
    if($metas->post) {
        $keywords_post = $metas->post ? $metas->post->keywords : '';
        if($keywords_post) $keywords = $keywords_post . $keywords;
    }
    echo '	<meta name="keywords" content="',htmlspecialchars($keywords),',',htmlspecialchars($tbopt->get('blog_name')),'" />'."\n";

    if($ispost) {
        // 依次输出header
        foreach($tax_metas as $tm) {
            echo $tm->header;
        }
    }
    if($metas->post) echo $metas->post->header;
}

add_hook('tb_head', 'postmetas_head');

function postmetas_footer() {
    if(!isset($GLOBALS['postmetas'])) {
        // 如果head没被成功调用
        return;
    }
    $metas = $GLOBALS['postmetas'];

    if(isset($metas->tax)) {
        foreach($metas->tax as $tm)
            echo $tm->footer;
    }
    
    if($metas->post) echo $metas->post->footer;

    unset($GLOBALS['postmetas']);
}

add_hook('tb_footer', 'postmetas_footer');
