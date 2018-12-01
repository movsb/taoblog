<?php

class TB_Query {
    public function query() {
        $this->is_query_modification = false;
        if(!$logged_in && isset($_SERVER['HTTP_IF_MODIFIED_SINCE'])) {
            $modified = $tbdate->http_gmt_to_mysql_datetime_gmt($_SERVER['HTTP_IF_MODIFIED_SINCE']);
            if($tbdate->is_valid_mysql_datetime($modified)) {
                $this->internal_query['modified'] = $modified;
                $this->is_query_modification = true;

                // 检查实体标签
                $etag = $_SERVER['HTTP_IF_NONE_MATCH'] ?? '';
                if($etag) {
                    $version = TB_VERSION;
                    if(!preg_match("/$version-/", $etag)) {
                        unset($this->internal_query['modified']);
                        $this->is_query_modification = false;
                    }
                }
            }
        }

        $need_redirect = false;
        if(isset($this->query['p'])) {
            $this->internal_query['id'] = (int)$this->query['p'];
            $need_redirect = true;
        }

        $this->internal_query = array_merge($this->internal_query, parse_query_string($u, false, false));

        // 查询文章
        $r = [];
        $q = &$this->internal_query;

        if($q['id'] ?? '') {
            $tbquery->type = 'post';
            $r = $tbpost->query_by_id((int)$q['id'], $q['modified']??'');
            if(is_array($r) && count($r)) {
                $tbquery->related_posts = $tbpost->get_related_posts($r[0]->id);
                $r[0]->page_view++;
                inc_page_view($r[0]->id);
            }
        }
        else if($q['slug'] ?? '') {
            $tbquery->type = 'post';
            $r = $tbpost->query_by_slug($q['tax'], $q['slug'], $q['modified']??'');
            if(is_array($r) && count($r)) {
                $tbquery->related_posts = $tbpost->get_related_posts($r[0]->id);
                $r[0]->page_view++;
                inc_page_view($r[0]->id);
            }
        }
        else if($q['page'] ?? '') {
            $tbquery->type = 'page';
            $r = $tbpost->query_by_page($q['parents'], $q['page'], $q['modified']??'');
            if(is_array($r) && count($r)) {
                $r[0]->page_view++;
                inc_page_view($r[0]->id);
            }
        }
        else {
            $tbquery->type = 'home';
            $posts = Invoke('/posts!latest?limit=20', 'json', null, false);
            $r = json_decode($posts);
        }
    }
}
