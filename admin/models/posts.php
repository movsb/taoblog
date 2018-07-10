<?php

class TB_Posts {
    public $error = '';

    public function update(&$arg){
        global $tbdb;
        global $tbdate;
        global $tbtax;
        global $tbtag;

        $def = [
            'id'            => 0,
            'date'          => '',
            'modified'      => '',
            'title'         => '',
            'content'       => '',
            'slug'          => '',
            'taxonomy'      => 1,
            'page_parents'  => '',
            'status'        => 'public',
            'metas'         => '',
            'source'        => '',
            'source_type'   => 'html',
        ];

        $arg = tb_parse_args($def, $arg);

        if(!$this->have($arg['id'])) {
            $this->error = '此文章不存在！';
            return false;
        }

        if(!$arg['title']) {
            $this->error = '标题不应为空！';
            return false;
        }

        if (preg_match('# |	|\'|"|;|/|\\\\|\\?|&|\.|<|>|:|@|\\$|%|\\^|\\*#', $arg['slug'])) {
            $this->error = '文章别名不规范！';
            return false;
        }

        if(!$tbtax->has((int)$arg['taxonomy'])) {
            $this->error = '文章所属分类不存在！';
            return false;
        }

        if(!in_array($arg['status'], ['public', 'draft'])) {
            $this->error = '文章发表状态不正确。';
            return false;
        }

        if(!in_array($arg['source_type'], ['html', 'markdown'])) {
            $this->error = '文章源类型不正确。';
            return false;
        }

        $type = $this->get_vars('type', 'id='.(int)$arg['id']);
        if(!$type) return false;
        if($type->type == 'page') {
            $parents = $arg['page_parents'];
            if($parents) {
                $parents = explode(',', $parents);
                $pid = $this->get_the_last_parents_id($parents);
                if($pid === false) {
                    $this->error = '父页面不存在';
                    return false;
                } else {
                    $arg['taxonomy'] = $pid;
                }
            } else {
                $arg['taxonomy'] = 0;
            }
        }


        $modified = &$arg['modified'];
        if(!$modified) {
            $modified = $tbdate->mysql_datetime_local();
        } else if($modified === '-') {
            $modified = '';
        }

        if($arg['date'] && !$tbdate->is_valid_mysql_datetime($arg['date'])
            || $modified && !$tbdate->is_valid_mysql_datetime($modified))
        {
            $this->error = '无效的时间格式!';
            return false;
        }

        // 转换成GMT时间
        if($arg['date']) $arg['date_gmt'] = $tbdate->mysql_local_to_gmt($arg['date']);
        if($arg['modified']) $arg['modified_gmt'] = $tbdate->mysql_local_to_gmt($arg['modified']);

        $succeed = false;

        if($arg['date_gmt']) {
            if($arg['modified']) {
                $sql = "UPDATE posts SET date=?,modified=?,title=?,content=?,slug=?,taxonomy=?,status=?,metas=?,source=?,source_type=? WHERE id=? LIMIT 1";
                if($stmt = $tbdb->prepare($sql)){
                    if($stmt->bind_param('sssssissssi',
                        $arg['date_gmt'],$arg['modified_gmt'],
                        $arg['title'], $arg['content'],$arg['slug'],
                        $arg['taxonomy'], $arg['status'], $arg['metas'],
                        $arg['source'], $arg['source_type'],
                        $arg['id']))
                    {
                        $r = $stmt->execute();
                        $stmt->close();

                        if($r) $succeed = true;;
                    }
                }
            } else {
                $sql = "UPDATE posts SET date=?,title=?,content=?,slug=?,taxonomy=?,status=?,metas=?,source=?,source_type=? WHERE id=? LIMIT 1";
                if($stmt = $tbdb->prepare($sql)){
                    if($stmt->bind_param('ssssissssi',
                        $arg['date_gmt'],
                        $arg['title'], $arg['content'],$arg['slug'],
                        $arg['taxonomy'], $arg['status'], $arg['metas'],
                        $arg['source'], $arg['source_type'],
                        $arg['id']))
                    {
                        $r = $stmt->execute();
                        $stmt->close();

                        if($r) $succeed = true;;
                    }
                }
            }
        } else {
            if($arg['modified']) {
                $sql = "UPDATE posts SET modified=?,title=?,content=?,slug=?,taxonomy=?,status=?,metas=?,source=?,source_type=? WHERE id=? LIMIT 1";
                if($stmt = $tbdb->prepare($sql)){
                    if($stmt->bind_param('ssssissi',
                        $arg['modified_gmt'], $arg['title'], $arg['content'],$arg['slug'],
                        $arg['taxonomy'], $arg['status'], $arg['metas'],
                        $arg['source'], $arg['source_type'],
                        $arg['id']))
                    {
                        $r = $stmt->execute();
                        $stmt->close();

                        if($r) $succeed = true;;
                    }
                }
            } else {
                $sql = "UPDATE posts SET title=?,content=?,slug=?,taxonomy=?,status=?,metas=?,source=?,source_type=? WHERE id=? LIMIT 1";
                if($stmt = $tbdb->prepare($sql)){
                    if($stmt->bind_param('sssissssi',
                        $arg['title'], $arg['content'],$arg['slug'],
                        $arg['taxonomy'], $arg['status'], $arg['metas'],
                        $arg['source'], $arg['source_type'],
                        $arg['id']))
                    {
                        $r = $stmt->execute();
                        $stmt->close();

                        if($r) $succeed = true;;
                    }
                }
            }
        }

        if($succeed) {
            $tbtag->update_post_tags((int)$arg['id'], $arg['tags']);
            return true;
        } else {
            $this->error = $stmt->error;
            return false;
        }
    }

    public function insert(&$arg){
        global $tbdb;
        global $tbdate;
        global $tbtax;
        global $tbtag;

        $def = [
            'date'              => '',
            'modified'          => '',
            'title'             => '',
            'content'           => '',
            'slug'              => '',
            'type'              => 'post',
            'taxonomy'          => 1,
            'status'            => 'public',
            'comment_status'    => 1,
            'tags'              => '',
            'page_parents'      => '',
            'metas'             => '',
            'source'            => '',
            'source_type'       => 'html',
        ];

        $arg = tb_parse_args($def, $arg);

        if(!$arg['title']) {
            $this->error = '标题不应为空！';
            return false;
        }

        if(!$arg['content']) {
            $this->error = '内容不应为空！';
            return false;
        }

        if(!$arg['slug']) $arg['slug'] = '-';
        if(!$arg['slug'] || preg_match('# |	|\'|"|;|/|\\\\|\\?|&|\\.|<|>|:|@|\\$|%|\\^|\\*#', $arg['slug'])) {
            $this->error = '文章别名不规范！';
            return false;
        }

        if(!$tbtax->has((int)$arg['taxonomy'])) {
            $this->error = '文章所属分类不存在！';
            return false;
        }

        if(!in_array($arg['status'], ['public', 'draft'])) {
            $this->error = '文章发表状态不正确。';
            return false;
        }

        if(!in_array($arg['source_type'], ['html', 'markdown'])) {
            $this->error = '文章源类型不正确。';
            return false;
        }

        $type = $arg['type'];
        if($type == 'page') {
            $parents = $arg['page_parents'];
            if($parents) {
                $parents = explode(',', $parents);
                $pid = $this->get_the_last_parents_id($parents);
                if($pid === false) {
                    $this->error = '父页面不存在';
                    return false;
                } else {
                    $arg['taxonomy'] = $pid;
                }
            } else {
                $arg['taxonomy'] = 0;
            }
        }

        if(!$arg['date']) {
            $arg['date'] = $tbdate->mysql_datetime_local();
        }

        if(!$arg['modified']) {
            $arg['modified'] = $arg['date']
            ? $arg['date']
            : $tbdate->mysql_datetime_local();
        }

        if(!$tbdate->is_valid_mysql_datetime($arg['date']) || !$tbdate->is_valid_mysql_datetime($arg['modified'])) {
            $this->error = '无效的时间格式!';
            return false;
        }

        // 转换成GMT时间
        $arg['date_gmt'] = $tbdate->mysql_local_to_gmt($arg['date']);
        $arg['modified_gmt'] = $tbdate->mysql_local_to_gmt($arg['modified']);

        $sql = "INSERT INTO posts (
            date,modified,title,content,slug,type,taxonomy,status,comment_status,metas,source,source_type)
            VALUES (?,?,?,?,?,?,?,?,?,?,?,?)";
        if($stmt = $tbdb->prepare($sql)){
            if($stmt->bind_param(
                'ssssssisisss',
                $arg['date_gmt'], $arg['modified_gmt'],
                $arg['title'], $arg['content'],$arg['slug'],
                $arg['type'], $arg['taxonomy'], $arg['status'],
                $arg['comment_status'], $arg['metas'],
                $arg['source'], $arg['source_type']
                )
            )
            {
                $r = $stmt->execute();
                $stmt->close();

                if($r) {
                    $iid = $tbdb->insert_id;

                    $tbtag->update_post_tags($iid, $arg['tags']);
                    return $iid;
                }
            }
        }

        $this->error = $stmt->error;

        return false;
    }

    private function after_posts_query(array $posts) {
        global $tbquery;
        global $tbdate;


        for($i=0; $i<count($posts); $i++) {
            $p = &$posts[$i];

            if(isset($p->date))
                $p->date = $tbdate->mysql_datetime_to_local($p->date);

            if(isset($p->modified))
                $p->modified = $tbdate->mysql_datetime_to_local($p->modified);

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
        global $tbdb;
        global $tbtax;
        global $tbopt;

        $sql = array();
        $sql['select']  = '*';
        $sql['from']    = 'posts';
        $sql['where']   = [];
        $sql['where'][] = 'id=' . $id;
        if($modified) {
            $sql['where'][] = "modified>'".$modified."'";
        }
        $sql['limit'] = 1;

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        $p = [];
        if($r = $rows->fetch_object()){
            $p[] = $r;
        }

        $p = $this->after_posts_query($p);

        return $p;
    }

    // 查询标签对应的文章集
    public function query_by_tags(string $tag, bool $all)
    {
        global $tbdb;
        global $tbquery;
        global $tbtag;

        $tbquery->tags = $tag;

        $ids = $tbtag->get_tag_id($tag);
        if ($ids === false) {
            return [];
        }
        $ids = $tbtag->getAliasTagsAll([$ids]);
        $ids = join(',', $ids);

        $sql = array();
        $sql['select']  = $all ? 'posts.*' : 'posts.id,posts.title';
        $sql['from']    = 'posts,post_tags';
        $sql['where']   = [];
        $sql['where'][] = "posts.id=post_tags.post_id";
        $sql['where'][] = "post_tags.tag_id in ($ids)";

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $results = $tbdb->query($sql);

        if (!$results) {
            return false;
        }

        $rows = $results;

        $p = [];
        while ($r = $rows->fetch_object()) {
            $p[] = $r;
        }

        $p = $this->after_posts_query($p);

        return $p;
    }

    // 获取日期对应的文章集
    public function query_by_date(int $yy, int $mm, int $count) {
        global $tbdb;
        global $tbquery;
        global $tbdate;

        $sql = array();
        $sql['select']  = '*';
        $sql['from']    = 'posts';
        $sql['where']   = [];

        if($yy >= 1970) {
            if($mm >= 1 && $mm <= 12) {
                $startend = $tbdate->the_month_startend_gmdate($yy, $mm);
            } else {
                $startend = $tbdate->the_year_startend_gmdate($yy);
            }

            $sql['where'][] = "date>='{$startend->start}' AND date<='{$startend->end}'";
        }

        $tbquery->date = (object)['yy'=>$yy,'mm'=>$mm];

        $sql['orderby'] = 'date DESC';

        if($count > 0) {
            $sql['limit'] = $count;
        }

        $sql = apply_hooks('before_query_posts', 0, $sql);
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

    // 获取最近（依据 date）
    public function query_by_latest(int $count)
    {
        return $this->query_by_date(0, 0, $count);
    }

    // 查询别名对应的单篇文章
    public function query_by_slug(string $tax, string $slug, string $modified){
        global $tbdb;
        global $tbtax;

        // 根据类似 /path/to/folder/post 的形式
        // 中 /path/to/folder 文件夹（分类层次）
        // 对应的分类中最后一个分类的 ID
        $taxid = $tbtax->id_from_tree($tax);
        if(!$taxid) return false;
        
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

        $sql = apply_hooks('before_query_posts', 0, $sql);
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

        $sql = apply_hooks('before_query_posts', 0, $sql);
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

    // 根据分类查询文章
    public function query_by_tax(string $tax){
        global $tbdb;
        global $tbtax;
        global $tbquery;

        $taxid = (int)$tbtax->id_from_tree($tax);
        if(!$taxid) return false;

        $tbquery->category = $tbtax->tree_from_id($taxid);

        $sql = array();

        $fields = "id,date,title,slug,type,taxonomy";

        $sql['select']  = $fields;
        $sql['from']    = 'posts';
        $sql['where']   = [];

        $s = "taxonomy=$taxid";
        $offsprings = $tbtax->get_offsprings($taxid);
        foreach($offsprings as $os)
            $s .= " OR taxonomy=$os";

        $sql['where'][] = $s;

        $sql['oderby']  = 'date DESC';

        $sql = apply_hooks('before_query_posts', 0, $sql);
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

    public function get_count_of_taxes($taxes=[]) {
        global $tbdb;

        $sql = array();
        $sql['select']  = 'count(id) as total';
        $sql['from']    = 'posts';
        $sql['where']   = [];

        foreach($taxes as $t) {
            $t = (int)$t;
            $sql['where'][] = "taxonomy=$t";
        }

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        return $rows->fetch_object()->total;
    }

    // 虽然名字跟上下两个很像，并完全不是在同一个时间段写的，功能貌似也并不相同
    public function get_count_of_cats_all() {
        global $tbdb;

        $sql = array();
        $sql['select']  = 'count(id) count,taxonomy';
        $sql['from']    = 'posts';
        $sql['groupby'] = 'taxonomy';

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        $ca = [];
        while($r = $rows->fetch_object())
            $ca[$r->taxonomy] = $r->count;

        return $ca;
    }

    public function get_count_of_date($yy=0, $mm=0) {
        global $tbdb;
        global $tbdate;

        $yy = (int)$yy;
        $mm = (int)$mm;

        $sql = array();
        $sql['select']  = 'count(id) as total';
        $sql['from']    = 'posts';
        $sql['where']   = [];

        if($yy >= 1970) {
            if($mm >= 1 && $mm <= 12) {
                $startend = $tbdate->the_month_startend_gmdate($yy, $mm);
            } else {
                $startend = $tbdate->the_year_startend_gmdate($yy);
            }

            $sql['where'][] = "date>='{$startend->start}' AND date<='{$startend->end}'";
        }

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        return $rows->fetch_object()->total;
    }

    public function get_title($id) {
        global $tbdb;

        $sql = "SELECT title FROM posts WHERE id=".(int)$id;
        $sql .= " LIMIT 1";
        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        return $rows->fetch_array()[0];
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

    private function &the_tag_names($id) {
        global $tbtag;

        return $tbtag->get_post_tag_names($id);
    }

    public function &get_all_posts_id() {
        global $tbdb;

        $sql = array();
        $sql['select']  = 'id';
        $sql['from']    = 'posts';
        $sql['where']   = [];
        $sql['where'][] = "type='post'";
        $sql['orderby'] = 'date DESC';

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $ids = [];
        $rows = $tbdb->query($sql);
        if(!$rows) return $ids;

        while($r = $rows->fetch_object())
            $ids[] = $r->id;

        return $ids;
    }

    public function &get_related_posts($id) 
    {
        global $tbdb;
        global $tbtag;

        $id = (int)$id;

        $posts = [];

        $tagids = $tbtag->get_post_tag_ids($id, true);

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

        $sql = apply_hooks('before_query_posts', 0, $sql);
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

    public function get_cat_posts($cid){
        global $tbdb;
        global $tbtax;
        global $tbquery;

        $cid = (int)$cid;
        if($cid <= 0) return false;

        $sql = array();
        $sql['select']  = 'id,date,title';
        $sql['from']    = 'posts';
        $sql['where']   = [];
        $sql['where'][] = "taxonomy=$cid";
        $sql['where'][] = "type='post'";
        $sql['orderby'] = 'date DESC';

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        $p = [];
        while($r = $rows->fetch_object()){
            $p[] = $r;
        }

        return $p;
    }

    public function get_date_archives() {
        global $tbdb;

        $sql = array();
        $sql['select']  = 'id,DATE_ADD(date, INTERVAL 8 HOUR) date';
        $sql['from']    = 'posts';
        $sql['where']   = [];
        $sql['where'][] = "type='post'";

        $sql = apply_hooks('before_query_posts', 0, $sql);
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

    // query_by_date 改的
    public function get_date_posts($yy, $mm) {
        global $tbdb;
        global $tbdate;

        $yy = (int)$yy;
        $mm = (int)$mm;

        $sql = array();
        $sql['select']  = 'id,date,title';
        $sql['from']    = 'posts';
        $sql['where']   = [];
        $sql['where'][] = "type='post'";

        if($yy >= 1970) {
            if($mm >= 1 && $mm <= 12) {
                $startend = $tbdate->the_month_startend_gmdate($yy, $mm);
            } else {
                $startend = $tbdate->the_year_startend_gmdate($yy);
            }

            $sql['where'][] = "date>='{$startend->start}' AND date<='{$startend->end}'";
        }

        $sql['orderby'] = 'date DESC';

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        $p = [];
        while($r = $rows->fetch_object()){
            $p[] = $r;
        }

        return $p;
    }

    public function get_count_of_type($type) {
        global $tbdb;

        $type = $tbdb->real_escape_string($type);

        $sql = array();
        $sql['select']  = 'count(*) as size';
        $sql['from']    = 'posts';
        $sql['where']   = [];
        $sql['where'][] = "type='$type'";

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return 0;
        return $rows->fetch_object()->size;
    }

    public function increase_page_view_count(int $pid) {
        global $tbdb;
        $sql = "UPDATE posts SET page_view=page_view+1 WHERE id=".$pid." LIMIT 1";
        $tbdb->query($sql);
    }

    // 获取所有文章，给文章管理用
    public function get_all_posts_for_manage() {
        global $tbdb;

        $sql = array();
        $sql['select']  = 'id,date,modified,title,page_view,source_type,comments';
        $sql['from']    = 'posts';
        $sql['where'][] = "type='post'";
        $sql['orderby'] = 'id DESC';

        $sql = apply_hooks('before_query_posts', 0, $sql);
        $sql = make_query_string($sql);

        $rows = $tbdb->query($sql);
        if(!$rows) return [];

        $posts = [];
        while($r = $rows->fetch_object()) {
            $posts[] = $r;
        }

        $posts = $this->after_posts_query($posts);

        return $posts;
    }
}

