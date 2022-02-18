document.write(function(){/*
	<!--评论标题 -->
	<h3 id="comment-title">
		文章评论
		<span class="no-sel count-wrap"> <span class="loaded">0</span>/<span class="total">0</span></span>
		<span class="post-comment">发表评论</span>
	</h3>
	<!-- 评论列表  -->
	<ol id="comment-list"></ol>
	<!-- 评论功能区  -->
	<div class="comment-func">
		<div>还没有用户发表评论，<span class="post-comment">发表评论</span></div>
	</div>
	<!-- 评论框 -->
	<div id="comment-form-div">
		<div class="no-sel nc drag-header">
			<div class="ncbtns">
				<img src="/images/close.svg" width="20" height="20" title="隐藏" class="closebtn"/>
			</div>
			<div class="comment-title">
				<span>评论</span>
			</div>
		</div>
		<form id="comment-form">
			<textarea id="comment-content" name="source" wrap="off"></textarea>
			<div class="fields">
				<input type="text" name="author" placeholder="昵称" />
				<input type="text" name="email" placeholder="邮箱（接收评论，不公开）"/>
				<input type="text" name="url" placeholder="个人站点（可不填）" />
				<input type="submit" id="comment-submit" value="发表评论" />
				<div class=prompt style="margin-top: 1em;"><b>注：</b>评论内容支持 Markdown 语法。</div>
			</div>
		</form>
	</div>
*/}.toString().slice(14,-3));

function Comment() {
    this._count  =0;
    this._loaded = 0;       // 已加载评论数
	this._loaded_ch = 0;
	this.post_id = 0;
	this.parent = 0;
}

Comment.prototype.init = function() {
	var self = this;

	self.post_id = +_post_id;

	self.convert_commenter();

    $('.post-comment').click(function(){
        self.reply_to(0);
    });

    // Ajax评论提交
    $('#comment-submit').click(function() {
		async function work() {
			$(this).attr('disabled', 'disabled');
			$('#comment-submit').val('提交中...');
			try {
				var cmt = await self.postComment();
				cmt = self.normalize_comment(cmt);
				var parent = self.parent;
				if(parent == 0) {
					$('#comment-list').prepend(self.gen_comment_item(cmt));
					// 没有父评论，避免二次加载。
					self._loaded ++;
				} else {
					self.add_reply_div(parent);
					$('#comment-reply-'+parent + ' ol:first').append(self.gen_comment_item(cmt));
				}
				$('#comment-'+cmt.id).fadeIn();
				TaoBlog.events.dispatch('comment', 'post', $('#comment-'+cmt.id));
				$('#comment-content').val('');
				$('#comment-form-div').fadeOut();
				self.save_info();
				self._count++;
				self.toggle_post_comment_button();
			} catch(e) {
				alert(e);
			} finally {
				$('#comment-submit').val('发表评论');
				$('#comment-submit').removeAttr('disabled');
			}
		}
		work();
        return false;
    });

    $('#comment-form-div .closebtn').click(function(){
            $('#comment-form-div').fadeOut();
    });

    // hide comment box when ESC key pressed
    window.addEventListener('keyup', function(e) {
        if(e.keyCode == 27) {
            $('#comment-form-div').fadeOut();
        }
    });

    window.addEventListener('scroll', function() {
        self.load_essential_comments();
    });

    window.addEventListener('load', function() {
        self.get_count(function() {
            self.load_essential_comments();
            self.toggle_post_comment_button(self._count == 0);
        });
    });

	self.init_drag(document.getElementById('comment-form-div'));
};

Comment.prototype.toggle_post_comment_button = function(show) {
    if (typeof show == 'undefined') {
        if (this._count <= 0) {
            show = true
        } else if(this._count == 1) {
            show = false;
        } else {
            return;
        }
    }
    if(show) {
        $('#comment-title .post-comment').fadeOut();
        $('.comment-func').fadeIn();
    } else {
        $('#comment-title .post-comment').fadeIn();
        $('.comment-func').fadeOut();
    }
};

Comment.prototype.load_essential_comments = function() {
    if(this._loaded + this._loaded_ch < this._count
        && window.scrollY + window.innerHeight + 100 >= document.body.scrollHeight) 
    {
        this.load_comments();
    }
};

Comment.prototype.get_count = function(callback) {
    var self = this;
    $.get('/v3/posts/' + self.post_id + '/comments:count',
        function(data) {
            self._count = data.count;
            $('#comment-title .total').text(self._count);
            callback();
        },
        'json'
    );
};

Comment.prototype.gen_avatar = function(id) {
	return '/v3/comments/' + id + '/avatar';
};

Comment.prototype.normalize_content = function(c) {
    var s = this.h2t(c);
    s = s.replace(/```(\s*(\w+)\s*)?\r?\n([\s\S]+?)```/mg, '<pre class="code"><code class="language-$2">$3</code></pre>');
    return s;
};

Comment.prototype.normalize_comment = function(c) {
	c.author = c.author || '';
	c.email = c.email || '';
	c.url = c.url || '';
	c.ip = c.ip || '';
	c.children = c.children || [];
	c.is_admin = c.is_admin || false;
	c.source_type = c.source_type || 'plain';
	c.source = c.source || (c.source_type == 'plain' ? c.content : c.source);
	return c;
}

Comment.prototype.friendly_date = function(d) {
	let date = new Date(d*1000);
	return date.toLocaleString();
};

// https://stackoverflow.com/a/12034334/3628322
// escape html to text
Comment.prototype.h2t = function(h) {
    var map = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
    };

    return h.replace(/[&<>]/g, function (s) {
        return map[s];
    });
};

// escape html to attribute
Comment.prototype.h2a = function(h) {
    var map = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      "'": '&#39;',
      '"': '&quot;'
    };

    return h.replace(/[&<>'"]/g, function (s) {
        return map[s];
    });
}

Comment.prototype.gen_comment_item = function(cmt) {
	var s = '';

    var loggedin = cmt.ip != '';

    // 登录后可以显示评论者的详细信息
    var info = '';
    if(loggedin) {
        // author, email, url, ip, date
		info = '编号：' + cmt.id
			+ '\n作者：' + this.h2a(cmt.author)
            + '\n邮箱：' + this.h2a(cmt.email)
            + '\n网址：' + this.h2a(cmt.url)
            + '\n地址：' + cmt.ip
            + '\n日期：' + new Date(cmt.date*1000).toLocaleString()
            ;
    }

	s += '<li style="display: none;" class="comment-li" id="comment-' + cmt.id + '" itemprop="comment">\n';
	s += '<div class="comment-avatar">'
		+ '<img src="' + this.gen_avatar(cmt.id) + '" width="48px" height="48px" title="'+info+'"/>'
		+ '</div>\n';
	s += '<div class="comment-meta">\n';

	if(cmt.is_admin) {
		s += '<span class="author">【作者】' + this.h2t(cmt.author) + '</span>\n';
	} else {
		var nickname;
		if(typeof cmt.url == 'string' && cmt.url.length) {
			if(!cmt.url.match(/^https?:\/\//i))
				cmt.url = 'http://' + cmt.url;
			nickname = '<a rel="nofollow" target="_blank" href="' + this.h2a(cmt.url) + '">' + this.h2t(cmt.author) + '</a>';
		} else {
			nickname = this.h2t(cmt.author);
		}

		s += '<span class="nickname">' + nickname + '</span>\n';
	}

	s += '<time class="date" datetime="' + (new Date(cmt.date*1000)).toJSON() + '">' + this.friendly_date(cmt.date) + '</time>\n</div>\n';
	if(cmt.source_type === 'markdown') {
		s += '<div class="comment-content html-content">' + cmt.content + '</div>\n';
	} else {
		s += '<div class="comment-content">' + this.normalize_content(cmt.content) + '</div>\n';
	}
    s += '<div class="toolbar" style="margin-left: 54px;">';
	s += '<a class="no-sel" onclick="comment.reply_to('+cmt.id+');return false;">回复</a>';
	if(loggedin) {
		s += '<a class="no-sel" onclick="confirm(\'确定要删除？\') && comment.delete_me('+cmt.id+');return false;">删除</a>';
	}
    s += '</div>';
	s += '</li>';

	return s;
};

Comment.prototype.reply_to = function(p){
	this.parent = +p;

	// 设置已保存的作者/邮箱/网址,其实只需要在页面加载完成后设置一次即可，是嘛？
	if(window.localStorage) {
		var commenter = JSON.parse(localStorage.getItem('commenter') || '{}');
		$('#comment-form input[name=author]').val(commenter.name || '');
		$('#comment-form input[name=email]').val(commenter.email || '');
		$('#comment-form input[name=url]').val(commenter.url || '');
	}

	this.move_to_center();
	$('#comment-form-div').fadeIn();
    $('#comment-content').focus();
};

Comment.prototype.move_to_center = function() {
	var e = $('#comment-form-div');
	var ww = window.innerWidth, wh = window.innerHeight;
	var ew = e.outerWidth(), eh = e.outerHeight();
	var left = (ww-ew)/2, top = (wh-eh)/2;
	e.css('left', left+'px');
	e.css('top', top+'px');
	console.log(ww,wh,ew,eh,left,top)
};

// https://www.w3schools.com/howto/howto_js_draggable.asp
Comment.prototype.init_drag = function(elmnt) {
	console.log('init_drag');
  var pos1 = 0, pos2 = 0, pos3 = 0, pos4 = 0;
  var dragElem = elmnt.getElementsByClassName("drag-header");
  if(!dragElem) { dragElem = elmnt; }
  else {dragElem = dragElem[0];}
  dragElem.onmousedown = dragMouseDown;
  console.log(dragElem);

  function dragMouseDown(e) {
    e = e || window.event;
    e.preventDefault();
    // get the mouse cursor position at startup:
    pos3 = e.clientX;
    pos4 = e.clientY;
    document.onmouseup = closeDragElement;
    // call a function whenever the cursor moves:
    document.onmousemove = elementDrag;
  }

  function elementDrag(e) {
    e = e || window.event;
    e.preventDefault();
    // calculate the new cursor position:
    pos1 = pos3 - e.clientX;
    pos2 = pos4 - e.clientY;
    pos3 = e.clientX;
    pos4 = e.clientY;
    // set the element's new position:
    elmnt.style.top = (elmnt.offsetTop - pos2) + "px";
    elmnt.style.left = (elmnt.offsetLeft - pos1) + "px";
  }

  function closeDragElement() {
    // stop moving when mouse button is released:
    document.onmouseup = null;
    document.onmousemove = null;
  }
};

Comment.prototype.delete_me = function(p) {
    var self = this;
	$.ajax({
        url: '/v3/comments/' + p,
        type: 'DELETE',
        success: function() {
            $('#comment-'+p).remove();
            self._count--;
            self.toggle_post_comment_button();
		},
        error: function(){
            alert('删除失败。');
        }
    });
};

// 为上一级评论添加div
Comment.prototype.add_reply_div = function(id){
	if($('#comment-'+id+' .comment-replies').length === 0){
		$('#comment-'+id).append('<div class="comment-replies" id="comment-reply-'+id+'"><ol></ol></div>');
	}
};

Comment.prototype.append_children = function(ch, p) {
	for(var i=0; i<ch.length; i++){
		if(!ch[i]) continue;

		if(ch[i].parent == p) {
			ch[i] = this.normalize_comment(ch[i]);
			$('#comment-reply-'+p + ' ol:first').append(this.gen_comment_item(ch[i]));
			$('#comment-'+ch[i].id).fadeIn();
            TaoBlog.events.dispatch('comment', 'post', $('#comment-'+ch[i].id));
			this.add_reply_div(ch[i].id);
			this.append_children(ch, ch[i].id);
			delete ch[i];
		}
	}
};


Comment.prototype.save_info = function() {
	if(window.localStorage) {
		var commenter = {
			name: $('#comment-form input[name=author]').val(),
			email: $('#comment-form input[name=email]').val(),
			url: $('#comment-form input[name=url]').val(),
		};
		localStorage.setItem('commenter', JSON.stringify(commenter));
	} else {
		alert('抱歉，你的浏览器不支持 localStorage，评论者的相关信息将无法正确地保存以便后续使用。');
	}
};

Comment.prototype.load_comments = function() {
    if (this.loading) {
        return;
    }
    this.loading = true;

    var self = this;

    $.get('/v3/posts/' + self.post_id + '/comments',
        {
            limit: 10,
			offset: self._loaded,
			order_by: 'id desc'
        },
        function(resp) {
			let cmts = resp.comments || [];
            var ch_count = 0;
            for(var i=0; i<cmts.length; i++){
				cmts[i] = self.normalize_comment(cmts[i]);
                $('#comment-list').append(self.gen_comment_item(cmts[i]));
                $('#comment-'+cmts[i].id).fadeIn();
                TaoBlog.events.dispatch('comment', 'post', $('#comment-'+cmts[i].id));
                self.add_reply_div(cmts[i].id);
                if(cmts[i].children) {
                    self.append_children(cmts[i].children, cmts[i].id);
                    ch_count += cmts[i].children.length;
                }
            }

            if(typeof(jQuery)=='function' && typeof(jQuery.timeago)=='function')
                jQuery('.comment-meta .date').timeago();

            if(cmts.length != 0) {
                self._loaded += cmts.length;
            }
            self._loaded_ch += ch_count;
            $('#comment-title .loaded').text(self._loaded + self._loaded_ch);
        },
        'json'
    ).fail(function(x) {
        alert(x.responseText);
    }).always(function(){
        self.loading = false;
    });
};

Comment.prototype.formData = function() {
	var form = document.getElementById('comment-form');
	var obj = {
		post_id: this.post_id,
		source_type: 'markdown',
		parent: this.parent,
		author: form['author'].value,
		email: form['email'].value,
		url: form['url'].value,
		source: form['source'].value
	};
	return obj;
};

Comment.prototype.postComment = async function() {
	var body = this.formData();
	var resp = await fetch(
		'/v3/posts/' + this.post_id + '/comments',
		{
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(body)
		}
	);
	if(!resp.ok) {
		throw new Error('发表失败：' + (await resp.json()).message);
	}
	return resp.json();
};

Comment.prototype.convert_commenter = function() {
	if(window.localStorage && !localStorage.getItem('commenter') && localStorage.getItem('cmt_author')) {
		var commenter = {
			name: localStorage.getItem('cmt_author') || '',
			email: localStorage.getItem('cmt_email') || '',
			url: localStorage.getItem('cmt_url') || '',
		};
		localStorage.setItem('commenter', JSON.stringify(commenter));
		localStorage.removeItem('cmt_author');
		localStorage.removeItem('cmt_email');
		localStorage.removeItem('cmt_url');
	}
};

if(!(typeof fetch == 'function')) {
	alert('你的浏览器版本过低（不支持 fetch），页面不能正确显示。');
}

var comment = new Comment();
comment.init();
