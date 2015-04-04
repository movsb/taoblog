<?php

class TB_PostObject {
	public $id;
	public $date;
	public $modified;
	public $title;
	public $content;
	public $slug;
	public $type;
	public $taxonomy;
	public $status;
	public $comment_status;
	public $password;
}

class TB_Posts {
	private function row_to_post(&$r){
		$p = new TB_PostObject;
		$fs = ['id','date','modified','title','content', 'slug',
			'type','taxonomy','status','comment_status','password'];
		foreach($fs as $f){
			$p->{$f} = $r[$f];
		}

		return $p;
	}

	public function insert($arg){
		global $tbdb;
		global $tbdate;

		$def = [
			'date' => $tb->mysql_datetime_gmt(),
			'modified' => $tb->mysql_datetime_gmt(),
			'title' => '',
			'content' => '',
			'slug' => '',
			'type' => 'post',
			'taxonomy' => 1,
			'status' => 'public',
			'comment_status' => 1,
			'password' => ''
		];

		$arg = tb_parse_args($def, $arg);

		$sql = "INSERT INTO posts (
			date,modified,title,content,slug,type,taxonomy,status,comment_status,password)
			VALUES (?,?,?,?,?,?,?,?,?,?)";
		if($stmt = $tbdb->prepare($sql)){
			if($stmt->bind_param('ssssssisis',
				$arg['date'], $arg['modified'],
				$arg['title'], $arg['content'],$arg['slug'],
				$arg['type'], $arg['taxonomy'], $arg['status'],
				$arg['comment_status'], $arg['password']))
			{
				$r = $stmt->execute();
				$stmt->close();

				return $r ? $tbdb->insert_id : $r;
			} else {
				return false;
			}
		} else {
			return false;
		}
	}

	public function query(&$arg){
		global $tbquery;
		global $tbdate;

		if(!is_array($arg))
			return false;

		$defs = ['p' => '', 'tax' => '', 'slug' => '', 
			'yy' => '', 'mm' => '', 'dd' => '',
			'password' => '', 'status' => '',
			'noredirect'=>true,
			'pageno' => 1,
			'modified' => false,
			];

		$arg = tb_parse_args($defs, $arg);

		if($arg['modified'] && !$tbdate->is_valid_mysql_datetime($arg['modified']))
			return false;

		$tbquery->pageno = (int)$arg['pageno'];

		$queried_posts = [];

		if($arg['p']){
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
			$queried_posts =  $this->query_by_tax($arg);
		} else if($arg['yy']) {
			$tbquery->type = 'date';
			$queried_posts = $this->queryy_by_date($arg);
		} else {
			$tbquery->type = 'home';
			$queried_posts = $this->query_home($arg);
		}

		if(!is_array($queried_posts)) return false;

		for($i=0; $i<count($queried_posts); $i++) {
			$queried_posts[$i]->date = $tbdate->mysql_datetime_to_local($queried_posts[$i]->date);
			$queried_posts[$i]->modified = $tbdate->mysql_datetime_to_local($queried_posts[$i]->modified);
		}

		return $queried_posts;
	}

	private function query_by_id(&$arg) {
		global $tbdb;
		global $tbtax;
		global $tbopt;

		if($arg['noredirect']) {
			$sql = "SELECT * FROM posts WHERE id=".intval($arg['p']);
			if($arg['modified']) {
				$sql .= " AND modified>'".$arg['modified']."'";
			}
			$rows = $tbdb->query($sql);
			if(!$rows) return false;

			$p = [];
			if($r = $rows->fetch_assoc()){
				$r['content'] = apply_hooks('the_content', $r['content']);
				$p[] = $this->row_to_post($r);
			}

			return $p;
			
		} else {
			$sql = "SELECT taxonomy,slug FROM posts WHERE id=".intval($arg['p']);

			$rows = $tbdb->query($sql);
			if(!$rows) return false;

			if($r = $rows->fetch_assoc()){
				$tax = $r['taxonomy'];
				$slug = $r['slug'];
				$tree = $tbtax->tree_from_id($tax);
				if(!$tree) return false;

				$tree = $tree['slug'];
				header('HTTP/1.1 301 Moved Permanently');
				header("Location: ".$tbopt->get('home')."/$tree$slug.html");
				die(0);
			}
		}

		return [];
	}

	private function query_home($arg) {
		global $tbdb;
		global $tbquery;

		$ppp = $tbquery->posts_per_page;

		$sql = "SELECT * FROM posts LIMIT ".(((int)$arg['pageno']-1)*$ppp).','.(int)$ppp;
		if($arg['status']){
			$sql .= " AND status='".$tbdb->real_escape_string($arg['status'])."'";
		}

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$p = [];
		while($r = $rows->fetch_assoc()){
			$p[] = $this->row_to_post($r);
		}

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

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$p = [];
		while($r = $rows->fetch_assoc()){
			$p[] = $this->row_to_post($r);
		}

		return $p;

	}

	public function get_count() {
		global $tbdb;

		$sql = "SELECT count(*) FROM posts";

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		return (int) $rows->fetch_array()[0];
	}

	public function have($id) {
		global $tbdb;

		$sql = "SELECT id FROM posts where id=".(int)$id;

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		return $rows->num_rows > 0; // 其实应该只能等于1的，如果有的话。
	}

}

