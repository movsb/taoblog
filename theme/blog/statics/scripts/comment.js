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
				<span id="comment-title-status">编辑评论</span>
			</div>
		</div>
		<form id="comment-form">
			<div class="content-area">
				<textarea class="overlay" id="comment-content" name="source" wrap="on"></textarea>
				<div class="overlay" id="comment-preview" style="display: none;"></div>
			</div>
			<div class="fields">
				<input type="text" name="author" placeholder="昵称" />
				<input type="text" name="email" placeholder="邮箱(不公开)"/>
				<input type="text" name="url" placeholder="个人站点(可不填)" />
				<input type="submit" id="comment-submit" value="发表评论" />
				<div class="field">
					<input type="checkbox" id="comment-wrap-lines" checked />
					<label for="comment-wrap-lines">自动折行</label>
				</div>
				<div class="field">
					<input type="checkbox" id="comment-show-preview" />
					<label for="comment-show-preview">显示预览</label>
				</div>
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
	this.being_edited = 0; // 正在被编辑的 ID，仅编辑时有效，> 0 时有效

	TaoBlog.comments = {};
}

Comment.prototype.init = function() {
	let self = this;

	self.post_id = +_post_id;

	self.convert_commenter();
	
	let postComments = document.querySelectorAll('.post-comment');
	postComments.forEach(function (elem) {
		elem.addEventListener('click', function(elem) {
			self.reply_to(0);
		});
	});

    // Ajax评论提交
    $('#comment-submit').click(function() {
		async function work() {
			$(this).attr('disabled', 'disabled');
			$('#comment-submit').val('提交中...');
			try {
				let cmt = {};
				if (self.being_edited >0) {
					cmt = await self.updateComment();
					cmt = self.normalize_comment(cmt);
					let wrap = document.querySelector(`#comment-${cmt.id}`);
					let content = wrap.querySelector(':scope > .comment-content');
					content.innerHTML = cmt.content;
				 } else {
					cmt = await self.postComment();
					cmt = self.normalize_comment(cmt);
					let parent = self.parent;
					if(parent == 0) {
						$('#comment-list').prepend(self.gen_comment_item(cmt));
						// 没有父评论，避免二次加载。
						self._loaded ++;
					} else {
						self.add_reply_div(parent);
						$('#comment-reply-'+parent + ' ol:first').append(self.gen_comment_item(cmt));
					}
					$('#comment-'+cmt.id).fadeIn();
					self._count++;
					self.toggle_post_comment_button();
				}
				TaoBlog.events.dispatch('comment', 'post', $('#comment-'+cmt.id), cmt);
				$('#comment-content').val('');
				self.showCommentBox(false);
				self.toggleShowPreview(false, false);
				self.save_info();
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
		self.showCommentBox(false);
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

	document.getElementById('comment-wrap-lines').addEventListener('click', self.wrapLines.bind(self));
	document.getElementById('comment-show-preview').addEventListener('click', self.showPreview.bind(self));

	self.init_drag(document.getElementById('comment-form-div'));
	
	TaoBlog.events.add('comment', 'post', function(jItem, cmt) {
		TaoBlog.comments[cmt.id] = cmt;
	});
};

// show         是否显示评论框
// callback     显示/隐藏完成后的回调函数
// options
//      allowEditingInfo    是否允许编辑评论者的信息
//      commenter           评论者的信息
Comment.prototype.showCommentBox = function(show, callback, options) {
	let self = this;

	let box = document.getElementById('comment-form-div');
	if (!show && (box.style.display == '' || box.style.display == 'none')) {
		return;
	}
	(show ? TaoBlog.fn.fadeIn : TaoBlog.fn.fadeOut)(box, callback);
	
	if (show) {
		if (typeof options != 'object' ) {
			options = {};
		}
		
		// 标题框
		let status = document.getElementById('comment-title-status');
		status.innerText = this.parent == 0
			? '发表评论'
			: this.parent > 0
				? '回复评论'
				: this.being_edited > 0
					? '编辑评论'
					: '（？？？）';
		
		// 编辑框是否可编辑？
		let inputs = document.querySelectorAll('#comment-form .fields input[type=text]');
		let allowEditingInfo =  options.allowEditingInfo ?? true;
		inputs.forEach(function(input) {
			// input.disabled = allowEditingInfo ? '' : 'disabled';
			// console.log(input);
			input.style.display = allowEditingInfo ? 'block' : 'none';
		});
		
		// 编辑框初始值
		// 设置已保存的作者/邮箱/网址,其实只需要在页面加载完成后设置一次即可，是嘛？
		{
			let savedCommenter = JSON.parse(localStorage.getItem('commenter') || '{}');
			let commenter = options.commenter ?? {
				author: savedCommenter.name ?? '',
				email: savedCommenter.email ?? '',
				url: savedCommenter.url ?? ''
			};
			let inputName = document.querySelector('#comment-form input[name=author]');
			let inputEmail = document.querySelector('#comment-form input[name=email]');
			let inputURL = document.querySelector('#comment-form input[name=url]');
			inputName.value = commenter.author;
			inputEmail.value = commenter.email;
			inputURL.value = commenter.url;

			let inputContent = document.querySelector('#comment-content');
			// 其它时候（未提交之前）不应该修改编辑的内容
			if (this.being_edited > 0) {
				inputContent.value = TaoBlog.comments[this.being_edited].source;
			}
		}
		
		let onEsc = function(e) {
			if (e.key === 'Escape') {
				self.showCommentBox(false);
				window.removeEventListener('keyup', onEsc);
			}
		};

		// 按 ESC 关闭（隐藏）窗口。
		window.addEventListener('keyup', onEsc);
	}
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
		&& window.scrollY + window.innerHeight + 1000 >= document.body.scrollHeight) 
	{
		this.load_comments();
	}
};

Comment.prototype.get_count = function(callback) {
    let self = this;
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
    let s = this.h2t(c);
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
	c.date_fuzzy = c.date_fuzzy ?? '';
	c.geo_location = c.geo_location ?? '';
	c.can_edit = c.can_edit ?? false;
	return c;
}

// https://stackoverflow.com/a/12034334/3628322
// escape html to text
Comment.prototype.h2t = function(h) {
	let map = {
		'&': '&amp;',
		'<': '&lt;',
		'>': '&gt;',
		'"': '&quot;',
		'\'': '&apos;',
	};

    return h.replace(/[&<>'"]/g, function (s) {
        return map[s];
    });
};

// escape html to attribute
Comment.prototype.h2a = function(h) {
    let map = {
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
	let loggedin = cmt.ip != '';
	let date = new Date(cmt.date * 1000);

	// 登录后可以显示评论者的详细信息
	let info = '';
    if(loggedin) {
		info = `编号：${cmt.id}
作者：${this.h2a(cmt.author)}
邮箱：${this.h2a(cmt.email)}
网址：${this.h2a(cmt.url)}
地址：${cmt.ip}
位置：${cmt.geo_location}
日期：${date.toLocaleString()}
`;
	}

	let urlContent = '';
	if(typeof cmt.url == 'string' && cmt.url.length) {
		let url = cmt.url;
		if(!url.match(/^https?:\/\//i)) {
			url = `http://${url}`;
		}
		try {
			let parsed = new URL(url);
			urlContent = `<span class="home"><a rel="nofollow" target="_blank" href="${this.h2a(url)}">${this.h2t(parsed.origin)}</a></span>`;
		} catch(e) {
			console.log(e);
		}
	}

	let html = `
<li style="display: none;" class="comment-li" id="comment-${cmt.id}">
	<div class="comment-avatar">
		<img src="${this.gen_avatar(cmt.id)}" width="48px" height="48px" title="${this.h2t(info)}"/>
	</div>
	<div class="comment-meta">
		<span class="${cmt.is_admin ? "author" : "nickname"}">${this.h2t(cmt.author)}</span>
		${urlContent}
		<time class="date" datetime="${date.toJSON()}" title="${date.toLocaleString()}">${cmt.date_fuzzy}</time>
	</div>
	${cmt.source_type === 'markdown'
		? `<div class="comment-content html-content">${cmt.content}</div>`
		: `<div class="comment-content">${this.normalize_content(cmt.content)}</div>`
	}
	<div class="toolbar" style="margin-left: 54px;">
		<a class="no-sel" onclick="comment.reply_to(${cmt.id});return false;">回复</a>
		${cmt.can_edit ? `<a class="no-sel" onclick="comment.edit(${cmt.id});return false;">编辑</a>` : ''}
		${!loggedin ? '' : `<a class="no-sel" onclick="confirm('确定要删除？') && comment.delete_me(${cmt.id});return false;">删除</a>`}
	</div>
</li>
`;

	return html;
};

Comment.prototype.reply_to = function(p){
	this.being_edited = -1;
	this.parent = +p;
	this.move_to_center();
	this.toggleShowPreview(false, false);
	this.showCommentBox(true, function() {
		$('#comment-content').focus();
	});
};

Comment.prototype.edit = function(c) {
	this.being_edited = c;
	this.parent = -1;
	this.move_to_center();
	this.toggleShowPreview(false, false);
	this.showCommentBox(true, function() {
		$('#comment-content').focus();
	}, {
		allowEditingInfo: false,
	});
};

Comment.prototype.move_to_center = function() {
	let e = $('#comment-form-div');
	let ww = window.innerWidth, wh = window.innerHeight;
	let ew = e.outerWidth(), eh = e.outerHeight();
	let left = (ww-ew)/2, top = (wh-eh)/2;
	e.css('left', left+'px');
	e.css('top', top+'px');
};

// https://www.w3schools.com/howto/howto_js_draggable.asp
Comment.prototype.init_drag = function(elmnt) {
	console.log('init_drag');
	let pos1 = 0, pos2 = 0, pos3 = 0, pos4 = 0;
	let dragElem = elmnt.getElementsByClassName("drag-header");
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
    let self = this;
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
	if($(`#comment-${id} .comment-replies`).length === 0){
		$(`#comment-${id}`).append(`<div class="comment-replies" id="comment-reply-${id}"><ol></ol></div>`);
	}
};

Comment.prototype.append_children = function(ch, p) {
	for(let i=0; i<ch.length; i++){
		if(!ch[i]) continue;

		if(ch[i].parent == p) {
			ch[i] = this.normalize_comment(ch[i]);
			$(`#comment-reply-${p} ol:first`).append(this.gen_comment_item(ch[i]));
			$(`#comment-${ch[i].id}`).fadeIn();
			TaoBlog.events.dispatch('comment', 'post', $(`#comment-${ch[i].id}`), ch[i]);
			this.add_reply_div(ch[i].id);
			this.append_children(ch, ch[i].id);
			delete ch[i];
		}
	}
};


Comment.prototype.save_info = function() {
	if(window.localStorage) {
		let commenter = {
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

    let self = this;

	$.get(
		`/v3/posts/${self.post_id}/comments`,
        {
            limit: 10,
			offset: self._loaded,
			order_by: 'id desc'
        },
        function(resp) {
			let cmts = resp.comments || [];
            let ch_count = 0;
            for(let i=0; i<cmts.length; i++){
				cmts[i] = self.normalize_comment(cmts[i]);
                $('#comment-list').append(self.gen_comment_item(cmts[i]));
                $('#comment-'+cmts[i].id).fadeIn();
                TaoBlog.events.dispatch('comment', 'post', $('#comment-'+cmts[i].id), cmts[i]);
                self.add_reply_div(cmts[i].id);
                if(cmts[i].children) {
                    self.append_children(cmts[i].children, cmts[i].id);
                    ch_count += cmts[i].children.length;
                }
            }

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
	let form = document.getElementById('comment-form');
	let obj = {
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

Comment.prototype.updateComment = async function() {
	let { source } = this.formData();
	let resp = await fetch(
		`/v3/comments/${this.being_edited}`,
		{
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				comment: {
					source_type: 'markdown',
					source: source,
				},
				update_mask: "source,sourceType"
			})
		}
	);
	if(!resp.ok) {
		throw new Error('编辑失败：' + (await resp.json()).message);
	}
	return resp.json();
};
Comment.prototype.postComment = async function() {
	let body = this.formData();
	let resp = await fetch(
		`/v3/posts/${this.post_id}/comments`,
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
		let commenter = {
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

Comment.prototype.wrapLines = function() {
	let checkBox = document.getElementById('comment-wrap-lines');
	let textarea = document.getElementById('comment-content');
	textarea.wrap = checkBox.checked ? "on" : "off";
};

Comment.prototype.toggleShowPreview = function(showPreview, checked) {
	let textarea   = document.getElementById('comment-content');
	let previewBox = document.getElementById('comment-preview');
	let checkBox   = document.getElementById('comment-show-preview');

	textarea.style.display = showPreview ? 'none' : 'block';
	previewBox.style.display = showPreview ? 'block' : 'none';
	
	if (typeof checked == 'boolean') {
		checkBox.checked = checked;
	}
};

Comment.prototype.showPreview = async function() {
	let checkBox   = document.getElementById('comment-show-preview');
	let previewBox = document.getElementById('comment-preview');

	if (!checkBox.checked) {
		this.toggleShowPreview(false);
		return;
	}
	
	this.toggleShowPreview(true);

	let source = document.getElementById('comment-form')['source'].value;
	let resp = await fetch('/v3/comments:preview', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({
			markdown: source,
			open_links_in_new_tab: true,
			post_id: +this.post_id
		})
	});

	if (!resp.ok) {
		previewBox.innerText = '预览失败：' + (await resp.json()).message;
		return;
	}

	let html = (await resp.json()).html;
	previewBox.innerHTML = html;
};

if(!(typeof fetch == 'function')) {
	alert('你的浏览器版本过低（不支持 fetch），评论不能正确显示。');
}

let comment = new Comment();
comment.init();
