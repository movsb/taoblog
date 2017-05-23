<?php

class TB_Query {
	public $category;

	public $type;
	public $objs;

	public $uri;
	public $query;
	private $internal_query;

	public $total;
	public $count;
	private $index;

	public $is_query_modification = false;

	// 查询类别
	public function is_home()       { return $this->type === 'home'; }
	public function is_post()       { return $this->type === 'post'; }
	public function is_page()       { return $this->type === 'page'; }
	public function is_singular()   { return $this->is_post() || $this->is_page(); }
	public function is_category()   { return $this->type === 'category' || $this->type === 'tax'; }
	public function is_date()       { return $this->type === 'date'; }
	public function is_tag()        { return $this->type === 'tag'; }
	public function is_404()        { return $this->type === '404'; }
	public function is_feed()       { return $this->type === 'feed'; }
	public function is_sitemap()    { return $this->type === 'sitemap'; } 
    public function is_archive()    { return $this->type === 'archive'; }
    public function is_memory()     { return $this->type === 'memory'; }

    // 临时使用，待寻找更好的解决办法
    public function push_404() {
        $this->type_origin = $this->type;
        $this->type = '404';
    }

    // 临时使用，待寻找更好的解决办法
    public function pop_404() {
        $this->type = $this->type_origin;
    }

	public function __construct() {
		global $tbopt;
		$this->internal_query = [];
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
        global $tbquery;
		global $tbopt;
		global $tbpost;
		global $tbtax;
		global $tbdate;

		global $logged_in;

		$this->parse_query_args();

		if(preg_match('#(\'|"|;|\./|\\\\|&|=|>|<)#', $this->uri))
			return false;

		$rules = [
            '^/(\d+)(/)?$'                                      => 'short=1&id=$1&slash=$2',
            '^(/\d+/[^/]+)$'                                    => 'file=$1',
            '^/archives/(\d+)\.html$'                           => 'id=$1',
            '^/date/((\d{4})/((\d{2})/)?)$'                     => 'yy=$2&mm=$4',
            '^/(.+)/([^/]+)\.html$'                             => 'long=1&tax=$1&slug=$2',
            '^/tags/(.+)$'                                      => 'tags=$1',
            '^/(feed|rss)(\.xml)?$'                             => 'feed=1',
            '^/sitemap\.xml$'                                   => 'sitemap=1',
            '^/memory$'                                         => 'memory=1',
            '^/archives$'                                       => 'archives=1',
            '^/(.+)/$'                                          => 'tax=$1',
            '^((/[0-9a-zA-Z\-_]+)*)/([0-9a-zA-Z\-_]+)$'         => 'parents=$1&page=$3',
            '^/index\.php$'                                     => '',
            '^/$'                                               => '',
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
				$this->internal_query['modified'] = $modified;
				$this->is_query_modification = true;

                // 检查实体标签
                $etag = $_SERVER['HTTP_IF_NONE_MATCH'] ?? '';
                if($etag) {
                    $version = TB_VERSION;
                    if(!preg_match("/$version-/", $etag)) {
                        unset($this->internal_query['modified']);
                        $this->is_query_modification = false;
                    }
                }
			}
		}

		$need_redirect = false;
		if(isset($this->query['p'])) {
			$this->internal_query['id'] = (int)$this->query['p'];
			$need_redirect = true;
		}

		$this->internal_query = array_merge($this->internal_query, parse_query_string($u, false, false));

		// 把类似 "/1234" 重定向到 "/12324/"
		if(isset($this->internal_query['short']) && !$this->internal_query['slash']) {
            $query = isset($_SERVER['QUERY_STRING']) && strlen($_SERVER['QUERY_STRING']) ? '?'.$_SERVER['QUERY_STRING'] : '';
			header('HTTP/1.1 301 Moved Permanently');
			header('Location: /'.$this->internal_query['id'].'/'.$query);
			die(0);
		}

        // 处理文件
        if(isset($this->internal_query['file'])) {
            $relative = '/'.FILE_DIR.$this->internal_query['file'];
            $absolute = TBPATH.$relative;
            if($logged_in && file_exists($absolute)) {
                header('HTTP/1.1 302 Use local file');
                header('Location: '.$relative);
            }
            else {
                header('HTTP/1.1 302 Use backup file');
                header('Location: '.FILE_HOST.$relative);
            }
            die(0);
        }

		// 处理RSS
		if($this->is_query_modification && isset($this->internal_query['feed'])) {
			if($tbdate->mysql_local_to_http_gmt($tbopt->get('last_post_time')) === $_SERVER['HTTP_IF_MODIFIED_SINCE']) {
				header('HTTP/1.1 304 Not Modified');
				die(0);
			}
		}

		// 处理 sitemap
		if(isset($this->internal_query['sitemap'])) {
			$this->type = 'sitemap';
			return true;
		}

        // 处理归档
        if(isset($this->internal_query['archives'])) {
            $this->type = 'archive'; // 没有 s
            return true;
        }

        // 处理说说
        if(isset($this->internal_query['memory'])) {
            $this->type = 'memory';
            return true;
        }

        // 查询文章
        $r = [];
        $q = &$this->internal_query;

        if($q['id'] ?? '') {
            $tbquery->type = 'post';
            $r = $tbpost->query_by_id((int)$q['id'], $q['modified']??'');
            if(is_array($r) && count($r)) {
                $tbquery->related_posts = $tbpost->get_related_posts($r[0]->id);
                $r[0]->page_view++;
                $tbpost->increase_page_view_count($r[0]->id);
            }
        }
        else if($q['slug'] ?? '') {
            $tbquery->type = 'post';
            $r = $tbpost->query_by_slug($q['tax'], $q['slug'], $q['modified']??'');
            if(is_array($r) && count($r)) {
                $tbquery->related_posts = $tbpost->get_related_posts($r[0]->id);
                $r[0]->page_view++;
                $tbpost->increase_page_view_count($r[0]->id);
            }
        }
        else if($q['page'] ?? '') {
            $tbquery->type = 'page';
            $r = $tbpost->query_by_page($q['parents'], $q['page'], $q['modified']??'');
            if(is_array($r) && count($r)) {
                $r[0]->page_view++;
                $tbpost->increase_page_view_count($r[0]->id);
            }
        }
        else if($q['tax'] ?? '') {
            $tbquery->type = 'tax';
            $r = $tbpost->query_by_tax($q['tax']);
        }
        else if($q['tags'] ?? '') {
            $tbquery->type = 'tag';
            $r = $tbpost->query_by_tags($q['tags']);
        }
        else if($q['feed'] ?? '') {
            $tbquery->type = 'feed';
            $r = $tbpost->query_by_latest(20);
        }
        else {
            $tbquery->type = 'home';
            $r = $tbpost->query_by_latest(20);
        }

		if($r === false || !is_array($r)) return $r;

		// 页面不能通过id访问，重定向到slug
		if(isset($this->internal_query['short']) && count($r) && $r[0]->type == 'page'){
			$need_redirect = true;
		}

		// 输出内容之前过滤
		for($i=0; $i<count($r); $i++) {
			$p = &$r[$i];

			if(isset($p->content))
				$p->content = apply_hooks('the_content', $p->content, $p->id);
		}

		$this->objs = &$r;

		$this->count = count($this->objs);
		$this->index = 0;

		if($need_redirect && $this->count) {
			$link = the_link($this->objs[0]);
			header('HTTP/1.1 301 Moved Permanently');

            // 干掉可能导致无限重定向的p参数
            unset($_GET['p']);
            $query = [];
            foreach($_GET as $k=>$v)
                $query[] = $k.'='.$v;
            $query = implode('&', $query);

			header('Location: '.$link.($query ? '?'.$query : ''));
			die(0);
		}

		return true;
	}

	private function parse_query_args(){
		$full_uri = filter_var($_SERVER['REQUEST_URI'], FILTER_SANITIZE_URL);
		$pos = strpos($full_uri, '?');
		if($pos !== false){
			$uri = substr($full_uri, 0, $pos);
			$query = substr($full_uri, $pos+1);
			if($query===false) $query = '';
		} else {
			$uri = $full_uri;
			$query = '';
		}

		// 都是解码后的值
		$this->uri = urldecode($uri);
		$this->query = parse_query_string($query);
	}
}

