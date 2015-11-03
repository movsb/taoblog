<?php

class TB_Comments {
	public $error = '';

	public function insert(&$arg){
		global $tbdb;
		global $tbpost;
		global $tbdate;
		global $tbopt;
		global $logged_in;

		$def = [
			'post_id'		=> false,
			'author'		=> '',
			'email'			=> '', 
			'url'			=> '',
			'ip'			=> $_SERVER['REMOTE_ADDR'],
			'date'			=> $tbdate->mysql_datetime_local(),
			'content'		=> '',
			'status'		=> 'public',
			'parent'		=> false,
			//'ancestor'	=> false,
		];

		$arg = tb_parse_args($def, $arg);

		// TRIM
		foreach(['author', 'email', 'url', 'content'] as &$f)
			$arg[$f] = trim($arg[$f]);

		if(!$arg['post_id']) {			$this->error = 'post_id 不能少！';	return false; } 

		if(!$arg['author'])	{			$this->error = 'author 不能为空';	return false; }
		if(preg_match("#\"|;|<|>|  |	|/|\\\\#", $arg['author'])){
			$this->error = '昵称不应包含特殊字符。';
			return false;
		}

		if(filter_var($arg['email'], FILTER_VALIDATE_EMAIL) === false) {
			$this->error = 'Email 不规范！';
			return false;
		}

		if($arg['url'] && filter_var($arg['url'], FILTER_VALIDATE_URL) === false) {
			$this->error = '网址不规范！';
			return false;
		}

		if(!$arg['content']) {
			$this->error = '评论内容不能为空!';
			return false; 
		}

		if(!$logged_in) {
			$not_allowed_authors = explode(',', $tbopt->get('not_allowed_authors').','.$tbopt->get('nickname'));
			foreach($not_allowed_authors as &$author) {
				if(strncmp($author,'/',1)==0 && preg_match($author.'i', $arg['author']) || strcasecmp($author, $arg['author'])==0) {
					$this->error = '您不应该使用此昵称。';
					return false;
				}
			}

			$not_allowed_emails = explode(',', $tbopt->get('not_allowed_emails').','.$tbopt->get('email'));
			foreach($not_allowed_emails as &$email) {
				if(strncmp($email,'/',1)==0 && preg_match($email.'i', $arg['email']) || strcasecmp($email, $arg['email'])==0) {
					$this->error = '您不应该使用此邮箱地址。';
					return false;
				}
			}
		}

		if(!$tbpost->have((int)$arg['post_id'])) {
			$this->error = '此评论对应的文章不存在！';
			return false;
		}

		// ancestor字段只能系统设置
		if(($arg['ancestor'] = $this->get_ancestor((int)$arg['parent'], true)) === false){
			return false;
		}

		$arg['date_gmt'] = $tbdate->mysql_local_to_gmt($arg['date']);

		$sql = "INSERT INTO comments (
			post_id,author,email,url,ip,date,content,status,parent,ancestor)
			VALUES (?,?,?,?,?,?,?,?,?,?)";
		if($stmt = $tbdb->prepare($sql)){
			if($stmt->bind_param('isssssssii',
				$arg['post_id'],	$arg['author'],		$arg['email'],
				$arg['url'],		$arg['ip'],			$arg['date_gmt'], 
				$arg['content'],	$arg['status'],
				$arg['parent'],		$arg['ancestor']))
			{
				$r = $stmt->execute();
				$stmt->close();

				if($r) {
					$id = $tbdb->insert_id;
					return $id;
				}
			}
		}

		$this->error = $tbdb->error;
		return false;
	}

	public function &get_children($p){
		global $tbdb;
		global $tbdate;

		$sql = "SELECT * FROM comments WHERE ancestor=".(int)$p;
		$result = $tbdb->query($sql);
		if(!$result){
			return false;
		}

		$children = [];
		while($obj = $result->fetch_object()){
			$obj->date = $tbdate->mysql_datetime_to_local($obj->date);
			$children[] = $obj;
		}

		$result->free();
		
		return $children;
	}

	public function &get(&$arg=[]) {
		global $tbdb;
		global $tbdate;

		$defs = [
			'id'		=> 0,
			'offset'	=> -1,
			'count'		=> -1,
			'post_id'	=> 0,
			'order'		=> 'asc',
			];

		$arg = tb_parse_args($defs, $arg);

		$id = (int)$arg['id'];
		if($id > 0) {
			$sql = "SELECT * FROM comments WHERE id=$id";
		} else {
			$sql = "SELECT * FROM comments WHERE parent=0 ";
			if((int)$arg['post_id'] > 0) {
				$sql .= " AND post_id=".(int)$arg['post_id'];
			}

			$sql .= " ORDER BY id ". (strtolower($arg['order'])==='asc' ? 'ASC' : 'DESC');

			$count = (int)$arg['count'];
			$offset = (int)$arg['offset'];
			if($count > 0) {
				if($offset >= 0) {
					$sql .= " LIMIT $offset,$count";
				} else {
					$sql .= " LIMIT $count";
				}
			}
		}

		$result = $tbdb->query($sql);
		if(!$result) {
			return false;
		}

		$cmts = [];
		while($obj = $result->fetch_object()){
			$obj->date = $tbdate->mysql_datetime_to_local($obj->date);
			$cmts[] = $obj;
		}

		if(!($id > 0)) {
			for($i=0; $i<count($cmts); $i++){
				$cmts[$i]->children = $this->get_children($cmts[$i]->id);
			}
		}

		$result->free();

		return $cmts;
	}

	public function get_ancestor($id, $return_this_id_if_zero=false) {
		global $tbdb;

		if((int)$id == 0) return 0;

		$sql = "SELECT ancestor FROM comments WHERE id=".(int)$id;
		$sql .= " LIMIT 1";
		$rows = $tbdb->query($sql);

		if(!$rows) {
			$this->error = $tbdb->error;
			return false;
		}

		if(!$rows->num_rows){
			$this->error = '待查询祖先的评论的ID不存在！';
			return false;
		}

		$ancestor = $rows->fetch_object()->ancestor;
		return $ancestor
			? $ancestor
			: ($return_this_id_if_zero 
				? $id
				: 0
				)
			;
	}

	public function &get_vars($fields, $where) {
		global $tbdb;

		$sql = "SELECT $fields FROM comments WHERE $where LIMIT 1";
		$rows = $tbdb->query($sql);
		if(!$rows) {
			$this->error = $tbdb->error;
			return false;
		}

		if(!$rows->num_rows) return null;

		$r = $rows->fetch_object();
		return $r;
	}

	public function &get_recent_comments() {
		global $tbdb;

		$cmts = [];

		$sql = "SELECT * FROM comments ORDER BY date DESC LIMIT 10";
		$rows = $tbdb->query($sql);
		if(!$rows) return $cmts;

		while($r = $rows->fetch_object())
			$cmts[] = $r;

		return $cmts;
	}
}

