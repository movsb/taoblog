<?php

class TB_Query {
	public $category;

	public $type;
	public $objs;

	public $uri;
	public $query;

	public $count;
	private $index;

	public $pageno;
	public $pages_per_page;

	public $is_query_modification = false;

	// 查询类别
	public function is_home() {
		return $this->type === 'home';
	}

	public function is_post() {
		return $this->type === 'post';
	}

	public function is_page() {
		return $this->type === 'page';
	}

	public function is_singular() {
		return $this->is_post()
			|| $this->is_page();
	}

	public function is_category() {
		return $this->type === 'category'
			|| $this->type === 'tax';
	}

	public function is_archive() {
		return $this->type === 'archive';
	}

	public function is_404() {
		return $this->type === '404';
	}

	public function is_feed() {
		return $this->type === 'feed';
	}

	public function __construct() {
		global $tbopt;
		$this->posts_per_page = (int)$tbopt->get('posts_per_page', 10);
	}

	public function have() {
		return $this->count && $this->index<$this->count;
	}

	public function has() {
		return $this->have();
	}

	public function &the() {
		if($this->index >= $this->count) return false;
		return $this->objs[$this->index++];
	}

	public function query() {
		global $tbopt;
		global $tbpost;
		global $tbtax;
		global $tbdate;

		global $logged_in;

		$this->parse_query_args();

		if(preg_match('#(\'|"|;|\./|\\\\)#', $this->uri))
			return false;

		$rules = [
			'^/archives/(\d+)\.html$'		=> 'id=$1',
			'^/page/(\d+)$'					=> 'pageno=$1',
			'^/(.+)/([^/]+)\.html$'			=> 'tax=$1&slug=$2',
			'^/(feed|rss)(\.xml)?$'				=> 'feed=1',
			'^/([0-9a-z]+)$'				=> 'slug=$1',
			'^/(.+)/$'						=> 'tax=$1',
			'^/index\.php$'					=> '',
			'^/$'							=> '',
			];
		
		foreach($rules as $rule => $rewrite){
			$pattern = '#'.$rule.'#';
			if(preg_match($pattern, $this->uri)){
				$u = preg_replace($pattern, $rewrite, $this->uri);
				break;
			}
		}

		if(!isset($u)) {
			$this->type = 'unknown';
			$this->objs = null;

			return false;
		}


		$this->is_query_modification = false;
		if(!$logged_in && isset($_SERVER['HTTP_IF_MODIFIED_SINCE'])) {
			$modified = $tbdate->http_gmt_to_mysql_datetime_gmt($_SERVER['HTTP_IF_MODIFIED_SINCE']);
			if($tbdate->is_valid_mysql_datetime($modified)) {
				$this->query['modified'] = $modified;
				$this->is_query_modification = true;
			}
		}

		$this->query = array_merge($this->query, parse_query_string($u));

		foreach($this->query as $n => $v){
			$this->query[$n] = urldecode($v);
		}

		$need_redirect = false;
		if(isset($this->query['p'])) {
			$this->query['id'] = (int)$this->query['p'];
			$need_redirect = true;
		}

		$r = $tbpost->query($this->query);
		if($r === false) return $r;

		$this->objs = &$r;

		$this->count = count($this->objs);
		$this->index = 0;

		if($need_redirect && $this->count) {
			$link = the_post_link($this->objs[0]);
			header('HTTP/1.1 301 Moved Permanently');
			header('Location: '.$link);
			die(0);
		}

		return true;
	}

	private function parse_query_args(){
		$full_uri = $_SERVER['REQUEST_URI'];
		$pos = strpos($full_uri, '?');
		if($pos !== false){
			$uri = substr($full_uri, 0, $pos);
			$query = substr($full_uri, $pos+1);
			if($query===false) $query = '';
		} else {
			$uri = $full_uri;
			$query = '';
		}

		$this->uri = sanitize_uri($uri);
		$this->query = parse_query_string($query);
	}
}

