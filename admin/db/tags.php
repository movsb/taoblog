<?php

class TB_Tags {
	public function insert_tag($name) {
		global $tbdb;

		$sql = "INSERT INTO tags (name) values (?);";
		if($stmt = $tbdb->prepare($sql)){
			$stmt->bind_param('s', $name);
			if($stmt->execute())
				return $tbdb->insert_id;
		}

		return false;
	}

	public function get_tag_id($name) {
		global $tbdb;

		$sql = "SELECT id FROM tags WHERE name='".$tbdb->real_escape_string($name)."' LIMIT 1";
		if(!($results = $tbdb->query($sql)) || !$results->num_rows) {
			return false;
		}

		return $results->fetch_object()->id;
	}

	public function has_tag_name($name) {
		return !!$this->get_tag_id($name);
	}

	public function &get_post_tag_names($id) {
		global $tbdb;

		$id = (int)$id;
		$sql = "SELECT tags.name FROM post_tags,tags WHERE post_tags.post_id=$id AND post_tags.tag_id=tags.id";

		$names = [];

		$results = $tbdb->query($sql);
		if(!$results) return $names;

		while($n = $results->fetch_object()) {
			$names[] = $n->name;
		}

		return $names;
	}

	public function &get_post_tag_ids($id) {
		global $tbdb;

		$id = (int)$id;
		$sql = "SELECT tag_id FROM post_tags WHERE post_id=$id";

		$ids = [];

		$results = $tbdb->query($sql);
		if(!$results) return $ids;

		while($n = $results->fetch_object()) {
			$ids[] = $n->tag_id;
		}

		return $ids;
	}

	/*
	  更新文章的标签列表

	  参数：
	      id - 文章ID
		  tags - 逗号分隔的标签列表

	  所需完成的操作：
	      得到原来的，与现在的作比较，计算出去掉的和增加的
		然后删除去掉的，插入增加的
	*/
	public function update_post_tags($id, $tags) {
		global $tbdb;

		$oldts = $this->get_post_tag_names($id);
		$newts = $tags ? explode(',', $tags) : [];

		$deleted = []; // 删除的
		$added = []; // 新增加的

		foreach($oldts as $o) {
			if(!in_array($o, $newts)) {
				$deleted[] = $o;
			}
		}

		foreach($newts as $n) {
            $n = trim($n);
			if($n && !in_array($n, $oldts)) {
				$added[] = $n;
			}
		}

		foreach($deleted as $d) {
			$tid = $this->get_tag_id($d);
			$this->delete_post_tag($id, $tid);
		}

		foreach($added as $a) {
			if(!$this->has_tag_name($a)) {
				$tid = $this->insert_tag($a);
			} else {
				$tid = $this->get_tag_id($a);
			}

			$this->insert_post_tag($id, $tid);
		}
	}

	public function insert_post_tag($post_id, $tag_id) {
		global $tbdb;

		$post_id = (int)$post_id;
		$tag_id = (int)$tag_id;

		$sql = "INSERT INTO post_tags (post_id,tag_id) values ($post_id,$tag_id)";
		$results = $tbdb->query($sql);
		if(!$results) return false;

		return $tbdb->insert_id;
	}

	public function delete_post_tag($post_id, $tag_id) {
		global $tbdb;

		$post_id = (int)$post_id;
		$tag_id = (int)$tag_id;

		$sql = "DELETE FROM post_tags WHERE post_id=$post_id AND tag_id=$tag_id LIMIT 1";
		$results = $tbdb->query($sql);
		if(!$results) return false;

		return true;
	}

    // 此函数用来获取所有的标签及其拥有的文章数
    // 正确返回：[{name: x},{name, x}]
    public function list_all_tags() {
        global $tbdb;

        $sql = "SELECT t.name,COUNT(pt.id) as size FROM post_tags pt,tags t WHERE pt.tag_id=t.id GROUP BY t.id" /* ORDER BY size DESC LIMIT ? */;
        $results = $tbdb->query($sql);
        if(!$results) return false;

        $tag_objs = [];
        while($to = $results->fetch_object())
            $tag_objs[] = $to;

        return $tag_objs;
    }
}

