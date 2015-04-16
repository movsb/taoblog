<?php

class TB_Post_SnJS {
	public function insert($id) {
		global $tbdb;

		$post = new stdClass;
		$post->header = $_POST['snjs_header'];
		$post->footer = $_POST['snjs_footer'];

		$sql = "INSERT INTO post_snjs (type,tid,value) values ('post',?,?)";
		if(($stmt = $tbdb->prepare($sql))
			&& $stmt->bind_param('is', $id, json_encode($post))
			&& $stmt->execute()
			)
		{

		}

	}

	public function update($pid) {
		global $tbdb;

		$snjs = new stdClass;
		$snjs->header = $_POST['snjs_header'];
		$snjs->footer = $_POST['snjs_footer'];

		$sql = "UPDATE post_snjs SET value=? WHERE tid=".(int)$pid;
		if(($stmt = $tbdb->prepare($sql))
			&& $stmt->bind_param('s', json_encode($snjs))
			&& $stmt->execute()
			)
		{

		}
	}

	public function &get_snjs($pid=0, $tid=0) {
		global $tbdb;

		$snjs = new stdClass;

		if($pid > 0) {
			$snjs->post = new stdClass;
			$snjs->post->header = $snjs->post->footer = '';
		}

		if($tid > 0) {
			$snjs->tax = new stdClass;
			$snjs->tax->header = $snjs->tax->footer = '';
		}

		$pid = (int)$pid;
		$tid = (int)$tid;

		if($pid > 0) {
			$sql = "SELECT value FROM post_snjs WHERE type='post' AND tid=$pid LIMIT 1";
			if(($r = $tbdb->query($sql)) && $r->num_rows>0) {
				$value = json_decode($r->fetch_object()->value);
				if(isset($value->header)) $snjs->post->header = $value->header;
				if(isset($value->footer)) $snjs->post->footer = $value->footer;
			}
		}

		if($tid > 0) {
			$sql = "SELECT value FROM post_snjs WHERE type='tax' AND tid=$tid LIMIT 1";
			if(($r = $tbdb->query($sql)) && $r->num_rows>0) {
				$value = json_decode($r->fetch_object()->value);
				if(isset($value->header)) $snjs->tax->header = $value->header;
				if(isset($value->footer)) $snjs->tax->footer = $value->footer;
			}
		}

		return $snjs;
	}
}

