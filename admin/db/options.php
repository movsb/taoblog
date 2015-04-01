<?php

class TB_Options {
	public function get($name){
		global $tbdb;
		$sql = 'SELECT value FROM options where name=?';
		if($stmt = $tbdb->prepare($sql)){
			$stmt->bind_param('s', $name);
			$stmt->execute();
			$stmt->bind_result($value);
			$stmt->fetch();
			$stmt->close();
			return $value;
		}

		return false;
	}

	public function set($name, $value){
		global $tbdb;
		if($this->has($name)){
			$sql = 'UPDATE options SET value=? WHERE name=?';
			if($stmt = $tbdb->prepare($sql)){
				$stmt->bind_param('ss', $value, $name);
				return $stmt->execute();
			}
		} else {
			$sql = 'INSERT INTO options (name,value) VALUES (?,?)';
			if($stmt = $tbdb->prepare($sql)){
				$stmt->bind_param('ss', $name, $value);
				return $stmt->execute();
			}
		}

		return false;
	}		

	public function has($name){
		global $tbdb;
		$sql = 'SELECT name FROM options WHERE name=?';
		if($stmt = $tbdb->prepare($sql)){
			$stmt->bind_param('s', $name);
			$stmt->execute();
			$ret = $stmt->get_result();
			return $ret!==false && $ret->num_rows>0;
		}

		return false;
	}

	public function del($name) {
		global $tbdb;
		$sql = 'DELETE FROM options WHERE name=?';
		if($stmt = $tbdb->prepare($sql)){
			$stmt->bind_param('s', $name);
			return $stmt->execute();
		}

		return false;
	}
}

