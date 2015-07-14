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

		// 页面无分类
		if($arg['type'] == 'page')
			$arg['taxonomy'] = 0;

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
			'yy' => '', 'mm' => '',
			'pageno' => '',
			'password' => '', 'status' => '',
			'modified' => false,
			'feed' => '',
			'no_content' => false,
			'tags' => '',
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
		} else if($arg['slug']) {
			if($arg['tax']) {
				$tbquery->type = 'post';
				$queried_posts = $this->query_by_slug($arg);
			} else {
				$tbquery->type = 'page';
				$queried_posts =  $this->query_by_page($arg);
			}
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

		$slug = $arg['slug'];

		$sql = "SELECT * FROM posts WHERE type='page' AND slug='".$tbdb->real_escape_string($slug)."'";
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

}

