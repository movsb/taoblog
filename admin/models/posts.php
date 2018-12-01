<?php

class TB_Posts {
    public $error = '';

    private function after_posts_query(array $posts) {
        for($i=0; $i<count($posts); $i++) {
            $p = &$posts[$i];

            $p->tag_names = $this->the_tag_names($p->id);

            if(isset($p->metas)) {
                $d = json_decode($p->metas);
                $p->metas_raw = $d ? $p->metas : '{}';
                $p->metas_obj = $d ? $d : new stdClass;
                unset($p->metas);
            }
        }

        return $posts;
    }

    // 根据 id 查询单篇文章
    // 未查询到文章时返回 false 或 []
    // 查询到文章时返回数组（仅一篇文章）
    public function query_by_id(int $id, string $modified) {
        $posts = Invoke('/posts/'.$id.'?modified='.urlencode($modified), 'json', null, false);
        $posts = json_decode($posts);
        return $this->after_posts_query($posts);
    }

    // 虽然名字跟上下两个很像，并完全不是在同一个时间段写的，功能貌似也并不相同
    public function get_count_of_cats_all() {
        $cats = Invoke('/categories!cat-count', 'json', null, false);
        $cats = json_decode($cats);
        // for compatible with php
        $c = [];
        foreach($cats as $cat) {
            $c[$cat->id] = $cat->count;
        }
        return $c;
    }

    // Go!
    public function have($id) {
        global $tbdb;

       $sql = "SELECT id FROM posts where id=".(int)$id;
        $sql .= " LIMIT 1";

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        return $rows->num_rows > 0; // 其实应该只能等于1的，如果有的话。
    }

    private function the_tag_names($id) {
        $names = Invoke('/posts/'.$id.'/tags', 'json', null, false);
        return json_decode($names);
    }

    public function the_next_id() {
        global $tbdb;

        $sql = "SELECT AUTO_INCREMENT FROM information_schema.tables WHERE table_name='posts' AND table_schema = DATABASE()";

        return $tbdb->query($sql)->fetch_object()->AUTO_INCREMENT;
    }

    // 得到父页面uri
    // 比如：page -> ddd，其父为 aaa -> bbb -> ccc
    // 则返回 /aaa/bbb/ccc，则最终的uri应为：/aaa/bbb/ccc/ddd
    public function get_the_parents_string($id) {
        global $tbdb;

        $id = (int)$id;

        $get_id = function ($id) use ($tbdb){
            $id = (int)$id;
            $sql = "SELECT type,taxonomy,slug FROM posts WHERE id=$id LIMIT 1";
            $rows = $tbdb->query($sql);
            if(!$rows || !$rows->num_rows) return false;

            $o = $rows->fetch_object();
            if($o->type != 'page') return false;

            return $o;
        };

        $uri = [];
        while($id) {
            $t = $get_id($id);
            if($t === false) return false;

            $uri[] = $t->slug;

            $id = $t->taxonomy;
        }

        // remove this
        unset($uri[0]);

        $uri = implode('/', array_reverse($uri));

        return $uri ? '/'.$uri : '';
    }
}
