<?php

class TB_Taxonomy_Object {
	public $id;
	public $name;
	public $slug;
	public $parent;
	public $ancestor;
}

class TB_Taxonomies {
	private function row_to_tax(&$r) {
		$t = new TB_Taxonomy_Object;
		$fs = ['id', 'name', 'slug', 'parent', 'ancestor'];
		foreach($fs as $f){
			$t->{$f} = $r[$f];
		}
		return $t;
	}

	public function get($id=0) {
		global $tbdb;
		
		$sql = "SELECT * FROM taxonomies";
		if($id) $sql .= " WHERE id=".intval($id);
		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$taxes = [];
		while($t = $rows->fetch_assoc())
			$taxes[] = $this->row_to_tax($t);

		return $taxes;
	}

	private function get_sons_by_id($id){
		global $tbdb;

		$sql = "SELECT id FROM taxonomies WHERE parent=".intval($id);
		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		$sons = [];
		while($r = $rows->fetch_assoc())
			$sons[] = $r['id'];

		return $sons;
	}		

	public function add($name, $slug, $parent, $ancestor){
		global $tbdb;
		
		$sql = "INSERT INTO taxonomies (name,slug,parent,ancestor) VALUES (?,?,?,?)";
		if(($stmt = $tbdb->prepare($sql)) 
			&& $stmt->bind_param('ssii', $name, $slug, $parent, $ancestor)
			&& $stmt->execute())
		{
			$stmt->close();
			$id = $tbdb->insert_id;
			return $id;
			//$tax = $this->get($id);
			//if(!$tax || !is_array($tax))
			//	return false;

			//return $tax[0];
		}

		return false;
	}

	public function modify($id, $name, $slug, $parent, $ancestor){
		global $tbdb;

		if(!$this->has($id)) return false;
		
		$sql = "UPDATE taxonomies SET name=?,slug=?,parent=?,ancestor=? WHERE id=?";
		if(($stmt = $tbdb->prepare($sql)) 
			&& $stmt->bind_param('ssiii', $name, $slug, $parent, $ancestor, $id)
			&& $stmt->execute())
		{
			$stmt->close();
			return true;
		}

		return false;
	}

	public function has($id){
		global $tbdb;
		$sql = 'SELECT name FROM taxonomies WHERE id=?';
		if(($stmt = $tbdb->prepare($sql))
			&& $stmt->bind_param('i', $id)
			&& $stmt->execute() )
		{
			$ret = $stmt->get_result();
			$stmt->close();
			return $ret!==false && $ret->num_rows>0;
		}

		return false;
	}

	/* 按ID删除某个分类
		需要删除的有：
			*. 递归删除以此ID为Parent的分类
			*. 删除以此ID为ID的分类
	*/
	public function del($id){
		global $tbdb;

		if(!$this->has($id)) return false;

		// 先取得孩子，全部递归删除
		$sons = $this->get_sons_by_id($id);
		foreach($sons as $son){
			$this->del($son);
		}

		// 最后删除自己
		$sql = "DELETE FROM taxonomies WHERE id=".intval($id);
		return $tbdb->query($sql);
	}

	public function tree_from_id($id, $ns=' => ', $ss='/') {
		global $tbdb;

		if(!$this->has($id)) return false;
		
		$slug = '';
		$name = '';
		while($id) {
			$t = $this->get($id)[0];
			$slug = $t->slug.$ss.$slug;
			$name = $t->name.$ns.$name;

			$id = $t->parent;
		}

		return compact('slug', 'name');
	}

	public function id_from_tree($tree) {
		global $tbdb;

		$ts = preg_split('~/~', $tree, -1, PREG_SPLIT_NO_EMPTY);
		if(count($ts)<1) return false;
		
		$sql = "SELECT id FROM taxonomies WHERE slug='".$tbdb->real_escape_string($ts[count($ts)-1])."'";
		if(count($ts) == 1) {
			$sql .= " AND parent=0";
		} else {
			$sql .= " AND parent IN (";
			for($i=count($ts)-2; $i>0; --$i) {
				$sql .= "SELECT id FROM `taxonomies` WHERE slug='".$tbdb->real_escape_string($ts[$i])."' AND parent IN (";
			}
			$sql .= "SELECT id FROM `taxonomies` WHERE slug='".$tbdb->real_escape_string($ts[0])."'";
			for($i=count($ts)-1; $i>0; --$i) {
				$sql .= ")";
			}
		}

		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		return $rows->fetch_assoc()['id'];
	}

	public function get_ancestor($id) {
		global $tbdb;
		
		$sql = "SELECT ancestor FROM taxonomies LIMIT 1 WHERE id=".intval($id);
		$rows = $tbdb->query($sql);
		if(!$rows) return false;

		return $rows->fetch_assoc()['ancestor'];
	}
}

