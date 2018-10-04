<?php

class TB_Posts {
    public $error = '';

    private function after_posts_query(array $posts, bool $date=true) {
        global $tbquery;
        global $tbdate;


        for($i=0; $i<count($posts); $i++) {
            $p = &$posts[$i];

            if($date) {
                if(isset($p->date))
                    $p->date = $tbdate->mysql_datetime_to_local($p->date);

                if(isset($p->modified))
                    $p->modified = $tbdate->mysql_datetime_to_local($p->modified);
            }

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

    private function before_posts_query(array $sql) {
        global $logged_in;

        if (!$logged_in) {
            $sql['where'][] = "status='public'";
        }

        return $sql;
    }

    // 根据 id 查询单篇文章
    // 未查询到文章时返回 false 或 []
    // 查询到文章时返回数组（仅一篇文章）
    public function query_by_id(int $id, string $modified) {
        $posts = Invoke('/posts/'.$id.'?modified='.urlencode($modified), 'json', null, false);
        $posts = json_decode($posts);
        return $this->after_posts_query($posts,false);
    }

    // 查询别名对应的单篇文章
    public function query_by_slug(string $tax, string $slug, string $modified){
        global $tbdb;

        // 根据类似 /path/to/folder/post 的形式
        // 中 /path/to/folder 文件夹（分类层次）
        // 对应的分类中最后一个分类的 ID
        $taxid = Invoke('/categories!parse?tree='.$tax, 'json', null, false);
        $taxid = json_decode($taxid);
        
        $slug = $tbdb->real_escape_string($slug);

        $sql = array();
        $sql['select']  = '*';
        $sql['from']    = 'posts';
        $sql['where']   = [];
        $sql['where'][] = "taxonomy=$taxid";
        $sql['where'][] = "slug='".$slug."'";
        if($modified) {
            $sql['where'][] = "modified>'".$modified."'";
        }
        $sql['limit']   = 1;

        $sql = $this->before_posts_query($sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        $p = [];
        while($r = $rows->fetch_object()){
            $p[] = $r;
        }

        $p = $this->after_posts_query($p);

        return $p;
    }

    // 查询指定页面
    // 页面格式：/parent/page
    public function query_by_page(string $parents, string $page, string $modified){
        global $tbdb;

        $parents = strlen($parents) ? explode('/', substr($parents, 1)) : [];
        $pid = $this->get_the_last_parents_id($parents);

        if($pid === false) return false;

        $page = $tbdb->real_escape_string($page);

        $sql = array();
        $sql['select']  = '*';
        $sql['from']    = 'posts';
        $sql['where']   = [];
        $sql['where'][] = "type='page'";
        $sql['where'][] = "taxonomy=$pid";
        $sql['where'][] = "slug='".$page."'";
        $sql['limit']   = 1;
        if($modified) {
            $sql['where'][] = "modified>'".$modified."'";
        }

        $sql = $this->before_posts_query($sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        $p = [];
        while($r = $rows->fetch_object()){
            $p[] = $r;
        }

        $p = $this->after_posts_query($p);

        return $p;

    }

    // 虽然名字跟上下两个很像，并完全不是在同一个时间段写的，功能貌似也并不相同
    public function get_count_of_cats_all() {
        global $tbdb;

        $sql = array();
        $sql['select']  = 'count(id) count,taxonomy';
        $sql['from']    = 'posts';
        $sql['groupby'] = 'taxonomy';

        $sql = $this->before_posts_query($sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        $ca = [];
        while($r = $rows->fetch_object())
            $ca[$r->taxonomy] = $r->count;

        return $ca;
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

    public function &get_vars($fields, $where) {
        global $tbdb;

        $sql = "SELECT $fields FROM posts WHERE $where LIMIT 1";
        $rows = $tbdb->query($sql);
        if(!$rows) {
            $this->error = $tbdb->error;
            return false;
        }

        if(!$rows->num_rows) return null;

        $r = $rows->fetch_object();
        return $r;
    }

    private function the_tag_names($id) {
        $names = Invoke('/posts/'.$id.'/tags', 'json', null, false);
        return json_decode($names);
    }

    public function &get_related_posts($id) 
    {
        global $tbdb;

        $id = (int)$id;

        $posts = [];

        $tagids = json_decode(Invoke("/posts/$id/tags!ids",'json',null,false));

        if (!$tagids || !count($tagids)) {
            return $posts;
        }

        $in_tags = join(',', $tagids);

        $sql = array();
        $sql['select']  = 'p.id,p.title,count(p.id) as relevance';
        $sql['from']    = 'posts p, post_tags pt';

        $sql['where'] = [];
        $sql['where'][] = "pt.post_id!=$id";
        $sql['where'][] = "p.id=pt.post_id";
        $sql['where'][] = "pt.tag_id in ($in_tags)";

        $sql['groupby'] = 'p.id';
        $sql['orderby'] = 'relevance DESC';
        $sql['limit']   = 9;   // TODO make configable

        $sql = $this->before_posts_query($sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows || !$rows->num_rows)
            return $posts;

        while($r = $rows->fetch_object())
            $posts[] = $r;

        return $posts;
    }

    public function the_next_id() {
        global $tbdb;

        $sql = "SELECT AUTO_INCREMENT FROM information_schema.tables WHERE table_name='posts' AND table_schema = DATABASE()";

        return $tbdb->query($sql)->fetch_object()->AUTO_INCREMENT;
    }

    // 通过父页面树得到最后一个父页面的id（也就是当前待查询页面的id的父页面）
    // 比如：uri -> /aaa/bbb/ccc/ddd
    // 则 传入 ['aaa', 'bbb', 'ccc], 传出 ccc 的id
    public function get_the_last_parents_id($parents) {
        global $tbdb;

        if(count($parents) <= 0) return 0;

        $sql = "SELECT id FROM posts WHERE slug='".$tbdb->real_escape_string($parents[count($parents)-1])."'";
        if(count($parents) == 1) {
            $sql .= " AND taxonomy=0 LIMIT 1";
        } else {
            $sql .= " AND taxonomy IN (";
            for($i=count($parents)-2; $i>0; --$i)
                $sql .= "SELECT id FROM posts WHERE slug='".$tbdb->real_escape_string($parents[$i])."' AND taxonomy IN (";
            $sql .= "SELECT id FROM posts WHERE slug='".$tbdb->real_escape_string($parents[0])."'";
            for($i=count($parents)-1; $i > 0; --$i)
                $sql .= ")";
        }

        $rows = $tbdb->query($sql);
        if(!$rows || !$rows->num_rows) return false;

        return $rows->fetch_object()->id;
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

    public function get_date_archives() {
        global $tbdb;

        $sql = array();
        $sql['select']  = 'id,DATE_ADD(date, INTERVAL 8 HOUR) date';
        $sql['from']    = 'posts';
        $sql['where']   = [];
        $sql['where'][] = "type='post'";

        $sql = $this->before_posts_query($sql);
        $sql = make_query_string($sql);

        $sql = "SELECT year,month,count(id) count FROM ("
            . "SELECT id,date,year(date) year, month(date) month FROM ("
            . $sql
            .') x) x GROUP BY year,month;';

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        $dd = [];
        // $r = {year:2011, month: 2, count: 3}
        while($r = $rows->fetch_object())
            $dd[] = $r;

        $x = [];
        foreach($dd as $d) {
            if(!isset($x[$d->year])) {
                $x[$d->year] = [];
            }
            $x[$d->year][$d->month] = $d->count;
        }

        return $x;
    }
}
