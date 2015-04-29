<?php

class TB_Comments {
	public $error = '';

	public function insert(&$arg){
		global $tbdb;
		global $tbpost;
		global $tbdate;

		$def = [
			'post_id'		=> false,
			'author'		=> false,
			'email'			=> false, 
			'url'			=> '',
			'ip'			=> $_SERVER['REMOTE_ADDR'],
			'date'			=> $tbdate->mysql_datetime_gmt(),
			'content'		=> '',
			'agent'			=> '',
			'status'		=> 'public',
			'parent'		=> false,
			//'ancestor'	=> false,
		];

		$arg = tb_parse_args($def, $arg);

		if(!$arg['post_id']) {			$this->error = 'post_id 不能少！';	return false; } 
		if(!$arg['author'])	{			$this->error = 'author 不能为空';	return false; }
		if(!is_email($arg['email'])) {	$this->error = 'Email 不规范！';	return false; }
		if(!$arg['content']) {			$this->error = '评论内容不能为空!';	return false; }

		// ancestor字段只能系统设置
		if(($arg['ancestor'] = $this->get_ancestor((int)$arg['parent'], true)) === false){
			return false;
		}

		$arg['url'] = '';

		if(preg_match("#\"|;|<|>|  |	|/|\\\\#", $arg['author'])){
			$this->error = '昵称不应包含特殊字符。';
			return false;
		}

		if(!$tbpost->have((int)$arg['post_id'])) {
			$this->error = '此评论对应的文章不存在！';
			return false;
		}

		$sql = "INSERT INTO comments (
			post_id,author,email,url,ip,date,content,agent,status,parent,ancestor)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)";
		if($stmt = $tbdb->prepare($sql)){
			if($stmt->bind_param('issssssssii',
				$arg['post_id'],	$arg['author'],		$arg['email'],
				$arg['url'],		$arg['ip'],			$arg['date'], 
				$arg['content'],	$arg['agent'],		$arg['status'],
				$arg['parent'],		$arg['ancestor']))
			{
				$r = $stmt->execute();
				$stmt->close();

				if($r) {
					$id = $tbdb->insert_id;
					apply_hooks('comment_posted', 0, $arg);
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

	public function get_count($post_id=0) {
		global $tbdb;

		$sql =  "SELECT count(*) FROM comments WHERE 1";
		if($post_id > 0) $sql .= " AND post_id=".(int)$post_id;

		$r = $tbdb->query($sql);
		if(!$r) return false;

		return (int)$r->fetch_assoc()['count(*)'];
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
}


