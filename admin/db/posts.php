<?php

class TB_Posts {
	public $error = '';

	public function update(&$arg){
		global $tbdb;
		global $tbdate;
		global $tbtax;
		global $tbtag;

		$def = [
			'id'		=> 0,
			'date'		=> '',
			'modified'	=> '',
			'title'		=> '',
			'content'	=> '',
			'slug'		=> '',
			'taxonomy'	=> 1,
            'page_parents'  => '',
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

		if(!$arg['content']) {
			$this->error = '内容不应为空！';
			return false;
		}

		if(!$arg['slug'] || preg_match('# |	|\'|"|;|/|\\\\|\\?|&|\.|<|>|:|@|\\$|%|\\^|\\*#', $arg['slug'])) {
			$this->error = '文章别名不规范！';
			return false;
		}

		if(!$tbtax->has((int)$arg['taxonomy'])) {
			$this->error = '文章所属分类不存在！';
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
				$sql = "UPDATE posts SET date=?,modified=?,title=?,content=?,slug=?,taxonomy=? WHERE id=? LIMIT 1";
				if($stmt = $tbdb->prepare($sql)){
					if($stmt->bind_param('sssssii',
						$arg['date_gmt'],$arg['modified_gmt'],
						$arg['title'], $arg['content'],$arg['slug'],
						$arg['taxonomy'], $arg['id']))
					{
						$r = $stmt->execute();
						$stmt->close();

						if($r) $succeed = true;;
					} 
				}
			} else {
				$sql = "UPDATE posts SET date=?,title=?,content=?,slug=?,taxonomy=? WHERE id=? LIMIT 1";
				if($stmt = $tbdb->prepare($sql)){
					if($stmt->bind_param('ssssii',
						$arg['date_gmt'],
						$arg['title'], $arg['content'],$arg['slug'],
						$arg['taxonomy'], $arg['id']))
					{
						$r = $stmt->execute();
						$stmt->close();

						if($r) $succeed = true;;
					} 
				}
			}
		} else {
			if($arg['modified']) {
				$sql = "UPDATE posts SET modified=?,title=?,content=?,slug=?,taxonomy=? WHERE id=? LIMIT 1";
				if($stmt = $tbdb->prepare($sql)){
					if($stmt->bind_param('ssssii',
						$arg['modified_gmt'], $arg['title'], $arg['content'],$arg['slug'],
						$arg['taxonomy'], $arg['id']))
					{
						$r = $stmt->execute();
						$stmt->close();

						if($r) $succeed = true;;
					} 
				}
			} else {
				$sql = "UPDATE posts SET title=?,content=?,slug=?,taxonomy=? WHERE id=? LIMIT 1";
				if($stmt = $tbdb->prepare($sql)){
					if($stmt->bind_param('sssii',
						$arg['title'], $arg['content'],$arg['slug'],
						$arg['taxonomy'], $arg['id']))
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
			'date' => '',
			'modified' => '',
			'title' => '',
			'content' => '',
			'slug' => '',
			'type' => 'post',
			'taxonomy' => 1,
			'status' => 'public',
			'comment_status' => 1,
			'password' => '',
			'tags' => '',
            'page_parents' => '',
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
			date,modified,title,content,slug,type,taxonomy,status,comment_status,password)
			VALUES (?,?,?,?,?,?,?,?,?,?)";
		if($stmt = $tbdb->prepare($sql)){
			if($stmt->bind_param('ssssssisis',
				$arg['date_gmt'], $arg['modified_gmt'],
				$arg['title'], $arg['content'],$arg['slug'],
				$arg['type'], $arg['taxonomy'], $arg['status'],
				$arg['comment_status'], $arg['password']))
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

	public function query(&$arg){
		global $tbquery;
		global $tbdate;

		if(!is_array($arg))
			return false;

		$defs = ['id' => '', 'tax' => '', 'slug' => '',
            'parents' => '', 'page' => '',
			'yy' => '', 'mm' => '',
			'pageno' => '',
			'password' => '', 'status' => '',
			'modified' => false,
			'feed' => '',
			'no_content' => false,
			'tags' => '',
            'comments' => '',
			];

		$arg = tb_parse_args($defs, $arg);

		if($arg['modified'] && !$tbdate->is_valid_mysql_datetime($arg['modified']))
			return false;

		if($arg['id'] && intval($arg['id']) <=0
			|| $arg['yy'] && intval($arg['yy']) < 1970
			|| $arg['mm'] && (intval($arg['mm'])<1 || intval($arg['mm'])>12)
			|| $arg['pageno'] && intval($arg['pageno']) < 1)
		{
			return false;
		}

		$arg['id'] = (int)$arg['id'];
		$arg['yy'] = (int)$arg['yy'];
		$arg['mm'] = (int)$arg['mm'];
		$arg['pageno'] = (int)$arg['pageno'];

		$tbquery->pageno = max(1,$arg['pageno']);

		$queried_posts = [];

		if($arg['id']){
			$tbquery->type = 'post';
			$queried_posts = $this->query_by_id($arg);
			// 查询相关文章
			if(is_array($queried_posts) && count($queried_posts)) {
				$tbquery->related_posts = $this->get_related_posts($queried_posts[0]->id);
			}
		} else if($arg['slug']) {
            $tbquery->type = 'post';
            $queried_posts = $this->query_by_slug($arg);
			// 查询相关文章
			if(is_array($queried_posts) && count($queried_posts)) {
				$tbquery->related_posts = $this->get_related_posts($queried_posts[0]->id);
			}
        } else if($arg['page']) {
            $tbquery->type = 'page';
            $queried_posts = $this->query_by_page($arg);
		} else if($arg['tax']) {
			$tbquery->type = 'tax';
			$arg['no_content'] = true;
			$queried_posts =  $this->query_by_tax($arg);
		} else if($arg['yy']) {
			$tbquery->type = 'date';
			$arg['no_content'] = true;
			$queried_posts = $this->query_by_date($arg);
		} else if($arg['pageno']) {
			$tbquery->type = 'date';
			$arg['no_content'] = true;
			$queried_posts = $this->query_by_date($arg);
		} else if($arg['feed']) {
			$tbquery->type = 'feed';
			unset($arg);
			$arg = ['pageno' => '1', 'yy'=>'', 'mm'=>'', 'no_content'=>false];
			$queried_posts = $this->query_by_date($arg);
		} else if($arg['tags']) {
			$tbquery->type = 'tag';
			$queried_posts = $this->query_by_tags($arg);
		} else {
			$tbquery->type = 'home';
			$queried_posts = [];
		}

		if(!is_array($queried_posts)) return false;

		for($i=0; $i<count($queried_posts); $i++) {
			$p = &$queried_posts[$i];

			if(isset($p->date))
				$p->date = $tbdate->mysql_datetime_to_local($p->date);
			if(isset($p->modified))
				$p->modified = $tbdate->mysql_datetime_to_local($p->modified);

			$p->tag_names = $this->the_tag_names($p->id);
		}

		return $queried_posts;
	}

	private function query_by_id(&$arg) {
		global $tbdb;
		global $tbtax;
		global $tbopt;

		$sql = "SELECT * FROM posts WHERE id=".intval($arg['id']);
		if($arg['modified']) {
			$sql .= " AND modified>'".$arg['modified']."'";
		}
		$sql .= " LIMIT 1";
		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$p = [];
		if($r = $rows->fetch_object()){
			$p[] = $r;
		}

		return $p;
	}

	private function query_by_tags(&$arg) {
		global $tbdb;
		global $tbquery;
		
		$tags = $tbdb->real_escape_string($arg['tags']);
		$tbquery->tags = $tags;
		
		$sql = "SELECT posts.* FROM posts,post_tags,tags ";
		$sql .= " WHERE posts.id=post_tags.post_id AND post_tags.tag_id=tags.id AND tags.name='$tags'";

		$results = $tbdb->query($sql);
		if(!$results) return false;

		$rows = $results;

		$p = [];
		while($r = $rows->fetch_object()) {
			$p[] = $r;
		}

		return $p;
	}

	private function query_by_date($arg) {
		global $tbdb;
		global $tbquery;
		global $tbdate;

		$yy = (int)$arg['yy'];
		$mm = (int)$arg['mm'];

		$content_filed = $arg['no_content'] ? '' : ',content';
		$fields = "id,date,title$content_filed,slug,type,taxonomy";
		$sql = "SELECT $fields FROM posts WHERE 1";
		if($yy >= 1970) {
			if($mm >= 1 && $mm <= 12) {
				$startend = $tbdate->the_month_startend_gmdate($yy, $mm);
			} else {
				$startend = $tbdate->the_year_startend_gmdate($yy);
			}

			$sql .= " AND date>='{$startend->start}' AND date<='{$startend->end}'";
		}

		$tbquery->date = (object)['yy'=>$yy,'mm'=>$mm];

		$ppp = (int)$tbquery->posts_per_page;
		$pageno = intval($arg['pageno']);
		$offset = ($pageno >= 1 ? $pageno-1 : 0) * $ppp;

		$sql .= " ORDER BY date DESC";
		$sql .= " LIMIT $offset,$ppp";

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$p = [];
		while($r = $rows->fetch_object()){
			$p[] = $r;
		}

		$tbquery->total = $this->get_count_of_date($yy, $mm);

		return $p;
	}
	
	private function query_by_slug($arg){
		global $tbdb;
		global $tbtax;

		$tax = $arg['tax'];
		$slug = $arg['slug'];

		$taxid = $tbtax->id_from_tree($tax);
		if(!$taxid) return false;

		$sql = "SELECT * FROM posts WHERE taxonomy=$taxid AND slug='".$tbdb->real_escape_string($slug)."'";
		if($arg['modified']) {
			$sql .= " AND modified>'".$arg['modified']."'";
		}
		$sql .= " LIMIT 1";

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$p = [];
		while($r = $rows->fetch_object()){
			$p[] = $r;
		}

		return $p;

	}

	private function query_by_page($arg){
		global $tbdb;

        $parents = $arg['parents'];
        $parents = strlen($parents) ? explode('/', substr($parents, 1)) : [];
        $pid = $this->get_the_last_parents_id($parents);

        if($pid === false) return false;
        
		$slug = $arg['page'];

		$sql = "SELECT * FROM posts WHERE type='page' AND taxonomy=$pid AND slug='".$tbdb->real_escape_string($slug)."'";
		if($arg['modified']) {
			$sql .= " AND modified>'".$arg['modified']."'";
		}
		$sql .= " LIMIT 1";

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$p = [];
		while($r = $rows->fetch_object()){
			$p[] = $r;
		}

		return $p;

	}

	private function query_by_tax($arg){
		global $tbdb;
		global $tbtax;
		global $tbquery;

		$tax = $arg['tax'];

		$taxid = (int)$tbtax->id_from_tree($tax);
		if(!$taxid) return false;

		$tbquery->category = $tbtax->tree_from_id($taxid);

		$ppp = (int)$tbquery->posts_per_page;
		$pageno = intval($arg['pageno']);
		$offset = ($pageno >= 1 ? $pageno-1 : 0) * $ppp;

		$content_filed = $arg['no_content'] ? '' : ',content';
		$fields = "id,date,title$content_filed,slug,type,taxonomy";
		$sql = "SELECT $fields FROM posts WHERE taxonomy=$taxid";

		$offsprings = $tbtax->get_offsprings($taxid);
		foreach($offsprings as $os)
			$sql .= " OR taxonomy=$os";

		$sql .= " ORDER BY date DESC";
		$sql .= " LIMIT $offset,$ppp";

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$p = [];
		while($r = $rows->fetch_object()){
			$p[] = $r;
		}

		$tbquery->total = $this->get_count_of_taxes(array_merge([$taxid],$offsprings));

		return $p;

	}

	public function get_count_of_taxes($taxes=[]) {
		global $tbdb;

		$sql = "SELECT count(id) as total FROM posts WHERE 0";
		foreach($taxes as $t) {
			$t = (int)$t;
			$sql .= " OR taxonomy=$t";
		}

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		return $rows->fetch_object()->total;
	}

    // 虽然名字跟上下两个很像，并完全不是在同一个时间段写的，功能貌似也并不相同
    public function get_count_of_cats_all() {
        global $tbdb;
        $sql = "SELECT count(id) count,taxonomy FROM `posts` WHERE 1 GROUP BY taxonomy";
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

		$sql = "SELECT count(id) as total FROM posts WHERE 1";

		if($yy >= 1970) {
			if($mm >= 1 && $mm <= 12) {
				$startend = $tbdate->the_month_startend_gmdate($yy, $mm);
			} else {
				$startend = $tbdate->the_year_startend_gmdate($yy);
			}

			$sql .= " AND date>='{$startend->start}' AND date<='{$startend->end}'";
		}

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

		$sql = "SELECT id FROM posts WHERE type='post' AND status='public' ORDER BY date DESC";

		$ids = [];
		$rows = $tbdb->query($sql);
		if(!$rows) return $ids;
		
		while($r = $rows->fetch_object())
			$ids[] = $r->id;

		return $ids;
	}

	public function &get_related_posts($id) {
		global $tbdb;
		global $tbtag;

		$id = (int)$id;

		$posts = [];

		$tagids = $tbtag->get_post_tag_ids($id);
		if(!$tagids || !count($tagids))
			return $posts;

		$in_tags = join(',', $tagids);
		
		$select = "SELECT p.id,p.title,count(p.id) as relevance FROM posts p,post_tags pt ";
		$where = " WHERE pt.post_id!=$id AND p.id=pt.post_id AND pt.tag_id in ($in_tags) ";
		$sql = $select . $where . 'GROUP BY p.id ORDER BY relevance DESC LIMIT 5';

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

		$fields = "id,date,title";
		$sql = "SELECT $fields FROM posts WHERE taxonomy=$cid AND type='post'";

		/*$offsprings = $tbtax->get_offsprings($cid);
		foreach($offsprings as $os)
			$sql .= " OR taxonomy=$os";
         */

		$sql .= " ORDER BY date DESC";
		//$sql .= " LIMIT $offset,$ppp";

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
        $sql = "SELECT year,month,count(id) count FROM (SELECT id,date,year(date) year, month(date) month FROM (SELECT id,DATE_ADD(date, INTERVAL 8 HOUR) date FROM posts WHERE type='post') x) x GROUP BY year,month;";
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

		$fields = "id,date,title";
		$sql = "SELECT $fields FROM posts WHERE 1 AND type='post'";     // TODO where
		if($yy >= 1970) {
			if($mm >= 1 && $mm <= 12) {
				$startend = $tbdate->the_month_startend_gmdate($yy, $mm);
			} else {
				$startend = $tbdate->the_year_startend_gmdate($yy);
			}

			$sql .= " AND date>='{$startend->start}' AND date<='{$startend->end}'";
		}

		$sql .= " ORDER BY date DESC";

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$p = [];
		while($r = $rows->fetch_object()){
			$p[] = $r;
		}

		return $p;
	}

    // query_by_tags 改的
	public function get_tag_posts($tag) {
		global $tbdb;
		
		$tag = $tbdb->real_escape_string($tag);
		$sql = "SELECT posts.id,posts.date,posts.title FROM posts,post_tags,tags ";
		$sql .= " WHERE type='post' AND posts.id=post_tags.post_id AND post_tags.tag_id=tags.id AND tags.name='$tag'"; // TODO WHERE

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$p = [];
		while($r = $rows->fetch_object()) {
			$p[] = $r;
		}

		return $p;
	}
}

