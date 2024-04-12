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

class CommentAPI
{
	constructor(postID) {
		this._postID = postID;
	}

	// 返回文章的评论数。
	async getCountForPost() {
		let path = `/v3/posts/${this._postID}/comments:count`;
		let rsp = await fetch(path);
		if (!rsp.ok) { throw rsp.statusText; }
		let json = await rsp.json();
		return +json.count;
	}
	
	// 创建一条评论。
	async createComment(bodyObj) {
		let path = `/v3/posts/${this._postID}/comments`;
		let rsp = await fetch(path, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(bodyObj)
		});
		if (!rsp.ok) {
			throw new Error('发表失败：' + (await rsp.json()).message);
		}
		return await rsp.json();
	}

	// 更新/“编辑”一条已有评论。
	// 返回更新后的评论项。
	// 参数：id        - 评论编号
	// 参数：source    - 评论 markdown 原文
	async updateComment(id, source) {
		let path = `/v3/comments/${id}`;
		let rsp = await fetch(path, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				comment: {
					source_type: 'markdown',
					source: source
				},
				update_mask: 'source,sourceType'
			})
		});
		if (!rsp.ok) { throw new Error('更新失败：' + (await rsp.json()).message); }
		return await rsp.json();
	}

	// 返回头像链接。
	avatarURLOf(id) {
		return `/v3/comments/${id}/avatar`;
	}
	

	// 删除一条评论。
	async deleteComment(id) {
		let path = `/v3/comments/${id}`;
		let rsp = await fetch(path, {
			method: 'DELETE'
		});
		if (!rsp.ok) { throw new Error(rsp.statusText); }
	}

	// 评论预览。
	async previewComment(postID, source) {
		let path = `/v3/comments:preview`;
		let rsp = await fetch(path, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				markdown: source,
				open_links_in_new_tab: true,
				post_id: postID
			})
		});
		if (!rsp.ok) {
			throw new Error((await rsp.json()).message);
		}
		return await rsp.json();
	}
	
	// 列举评论。
	async listComments(postID, args) {
		let path = `/v3/posts/${postID}/comments?${args}`;
		let rsp = await fetch(path);
		if (!rsp.ok) {
			throw new Error(rsp.statusText);
		}
		return await rsp.json();
	}
}

// 代表一个用来操作评论项的类。
class CommentNode {
	constructor(node) {
		this._node = node;
	}

	setContent(html) {
		let content = this._node.querySelector(':scope > .comment-content');
		content.innerHTML = html;
	}
}

function Comment() {
    this._count  =0;
    this._loaded = 0;       // 已加载评论数
	this._loaded_ch = 0;
	this.post_id = 0;
	
	this.being_replied = 0; // 正在回复的评论。
	this.being_edited = 0;  // 正在被编辑的 ID，仅编辑时有效，> 0 时有效

	// 页面评论数量。
	this._elemTotal = document.querySelector('#comment-title .total');
	// 总评论列表。
	this._elemList = document.querySelector('#comment-list');
	
	// 每篇文章都有 `_post_id`。
	this.api = new CommentAPI(+_post_id);

	// 当前页面的所有已加载的评论。
	// 缓存起来的目的是为了“编辑”。
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
	let submit = document.querySelector('#comment-submit');
	submit.addEventListener('click', async function(event){
		event.preventDefault();
		event.stopPropagation();
		
		try {
			self.setStates({submitting: true});
			if (self.being_edited >0) {
				await self.updateComment();
			 } else {
				await self.createComment();
			}
		} catch(e) {
			alert(e);
		} finally {
			self.setStates({submitted: true});
		}
	});

	document.querySelector('#comment-form-div .closebtn').addEventListener('click', function(){
		self.showCommentBox(false);
	});

    window.addEventListener('scroll', function() {
        self.load_essential_comments();
    });

	window.addEventListener('load', function() {
		self.getCount();
	});

	document.getElementById('comment-wrap-lines').addEventListener('click', self.wrapLines.bind(self));
	document.getElementById('comment-show-preview').addEventListener('click', self.showPreview.bind(self));

	self.init_drag(document.getElementById('comment-form-div'));
	
	TaoBlog.events.add('comment', 'post', function(item, cmt) {
		TaoBlog.comments[cmt.id] = cmt;
	});
};

// 返回指定编号的评论项根元素节点。
Comment.prototype.nodeOf = function(id) {
	return document.querySelector(`#comment-${id}`);
}
Comment.prototype.replyList = function(id) {
	return document.querySelector(`#comment-reply-${id} ol:first-child`);
}
Comment.prototype.clearContent = function() {
	let content = document.querySelector('#comment-content');
	content.value = '';
}

// 投递新增评论的通知消息。
Comment.prototype.dispatch = function(node, cmt) {
	TaoBlog.events.dispatch('comment', 'post', node, cmt);
}

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
		status.innerText = this.being_replied == 0
			? '发表评论'
			: this.being_replied > 0
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
	let btn = document.querySelectorAll('#comment-title .post-comment');
	let func = document.querySelectorAll('.comment-func');
	if(show) {
		btn.forEach((b)=>b.style.display = 'none');
		func.forEach((f)=>TaoBlog.fn.fadeIn(f));
	} else {
		btn.forEach((b)=> b.style.display = 'unset');
		func.forEach((f)=>TaoBlog.fn.fadeOut(f));
	}
};

Comment.prototype.load_essential_comments = function() {
	if(this._loaded + this._loaded_ch < this._count
		&& window.scrollY + window.innerHeight + 1000 >= document.body.scrollHeight) 
	{
		this.load_comments();
	}
};

// 获取文章的最新评论数。
// 获取完成后会自动按需加载评论。
Comment.prototype.getCount = async function(callback) {
	try {
		let count = await this.api.getCountForPost();

		this._elemTotal.innerText = count;
		this._count = count;

		this.load_essential_comments();
		this.toggle_post_comment_button(this._count == 0);
	} catch(e) {
		alert(e);
	}
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
		<img src="${this.api.avatarURLOf(cmt.id)}" width="48px" height="48px" title="${this.h2t(info)}"/>
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
	this.being_replied = +p;
	this.move_to_center();
	this.toggleShowPreview(false, false);
	this.showCommentBox(true, function() {
		document.querySelector('#comment-content').focus();
	});
};

Comment.prototype.edit = function(c) {
	this.being_edited = c;
	this.being_replied = -1;
	this.move_to_center();
	this.toggleShowPreview(false, false);
	this.showCommentBox(true, function() {
		document.querySelector('#comment-content').focus();
	}, {
		allowEditingInfo: false,
	});
};

Comment.prototype.move_to_center = function() {
	let div = document.querySelector('#comment-form-div');
	let ww = window.innerWidth, wh = window.innerHeight;
	let ew = parseInt(getComputedStyle(div)['width']);
	let eh = parseInt(getComputedStyle(div)['height']);
	let left = (ww-ew)/2, top = (wh-eh)/2;
	div.style.left = `${left}px`;
	div.style.top = `${top}px`;
	console.table({
		ww: ww,
		wh: wh,
		ew: ew,
		eh: eh,
		left: left,
		top: top,
	});
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

Comment.prototype.delete_me = async function(id) {
	try {
		this.api.deleteComment(id);
		let node = this.nodeOf(id);
		node.remove();
		this._count--;
		this.toggle_post_comment_button();
	} catch(e) {
		alert(e);
	}
};

// 为上一级评论添加div
Comment.prototype.add_reply_div = function(id){
	let replies = document.querySelectorAll(`#comment-${id} .comment-replies`);
	if (replies.length == 0) {
		let div = document.createElement('div');
		div.innerHTML = `<div class="comment-replies" id="comment-reply-${id}"><ol></ol></div>`;
		let parent = document.querySelector(`#comment-${id}`);
		parent.appendChild(div.firstElementChild);
	}
};

Comment.prototype.append_children = function(ch, p) {
	for(let i=0; i<ch.length; i++){
		if(!ch[i]) continue;

		if(ch[i].parent == p) {
			ch[i] = this.normalize_comment(ch[i]);
			let ol = document.querySelector(`#comment-reply-${p} ol:first-child`);
			let div = document.createElement('div');
			div.innerHTML = this.gen_comment_item(ch[i]);
			ol.appendChild(div.firstElementChild);
			let item = document.querySelector(`#comment-${ch[i].id}`);
			TaoBlog.fn.fadeIn(item);
			TaoBlog.events.dispatch('comment', 'post', item, ch[i]);
			this.add_reply_div(ch[i].id);
			this.append_children(ch, ch[i].id);
			delete ch[i];
		}
	}
};


Comment.prototype.save_info = function() {
	let commenter = {
		name: document.querySelector('#comment-form input[name=author]').value,
		email: document.querySelector('#comment-form input[name=email]').value,
		url: document.querySelector('#comment-form input[name=url]').value,
	};
	localStorage.setItem('commenter', JSON.stringify(commenter));
};

Comment.prototype.load_comments = async function() {
	if (this.loading) {
		return;
	}

	let cmts = [];

	try {
		this.loading = true;

		let args = new URLSearchParams;
		args.set('limit', '10');
		args.set('offset', `${this._loaded}`);
		args.set('order_by', 'id desc');

		let rsp = await this.api.listComments(this.post_id, args);
		cmts = rsp.comments;
	} catch(e) {
		alert(e);
		return;
	} finally {
		this.loading = false;
	}

	let ch_count = 0;
	for(let i=0; i<cmts.length; i++){
		cmts[i] = this.normalize_comment(cmts[i]);
		let node = this.createNodeFrom(cmts[i]);
		this._elemList.appendChild(node);
		TaoBlog.fn.fadeIn(node);
		TaoBlog.events.dispatch('comment', 'post', node, cmts[i]);
		this.add_reply_div(cmts[i].id);
		if(cmts[i].children) {
			this.append_children(cmts[i].children, cmts[i].id);
			ch_count += cmts[i].children.length;
		}
	}

	if(cmts.length != 0) {
		this._loaded += cmts.length;
	}
	this._loaded_ch += ch_count;
	document.querySelector('#comment-title .loaded').innerText = this._loaded + this._loaded_ch;
};

Comment.prototype.formData = function() {
	let form = document.getElementById('comment-form');
	let obj = {
		post_id: this.post_id,
		source_type: 'markdown',
		parent: this.being_replied,
		author: form['author'].value,
		email: form['email'].value,
		url: form['url'].value,
		source: form['source'].value
	};
	return obj;
};

Comment.prototype.updateComment = async function() {
	let { source } = this.formData();
	let id = this.being_edited;

	let cmt = await this.api.updateComment(id, source);
	cmt = this.normalize_comment(cmt);

	let node = this.nodeOf(id);
	let obj = new CommentNode(node);
	obj.setContent(cmt.content);

	this.dispatch(node, cmt);
	
	this.clearContent();
	this.showCommentBox(false);
	this.toggleShowPreview(false, false);

	return cmt;
};

Comment.prototype.createNodeFrom = function(cmt) {
	let div = document.createElement('div');
	div.innerHTML = this.gen_comment_item(cmt);
	return div.firstElementChild;
}

Comment.prototype.setStates = function(states) {
	let submitButton = document.querySelector('#comment-submit');
	
	if (states.submitting) {
		submitButton.setAttribute('disabled', 'disabled');
		submitButton.value = '提交中……';
	}
	if (states.submitted) {
		submitButton.value = '发表评论';
		submitButton.removeAttribute('disabled');
	}
}

Comment.prototype.createComment = async function() {
	let body = this.formData();

	let cmt = await this.api.createComment(body);
	cmt = this.normalize_comment(cmt);

	let node = this.createNodeFrom(cmt);

	// 没有回复谁，插入到最新评论。
	if (this.being_replied == 0) {
		this._elemList.prepend(node);
		this._loaded++;
	} else {
		this.add_reply_div(this.being_replied);
		let list = this.replyList(this.being_replied);
		list.appendChild(node);
	}

	this.dispatch(node, cmt);

	TaoBlog.fn.fadeIn(node);
	this._count++;
	this.toggle_post_comment_button();

	this.clearContent();
	this.showCommentBox(false);
	this.toggleShowPreview(false, false);
	this.save_info();

	return cmt;
};

Comment.prototype.convert_commenter = function() {
	if(!localStorage.getItem('commenter') && localStorage.getItem('cmt_author')) {
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

	try {
		let rsp = await this.api.previewComment(+this.post_id, source);
		previewBox.innerHTML = rsp.html;
	} catch(e) {
		previewBox.innerText = '预览失败：' + e;
		return;
	}
};

let comment = new Comment();
comment.init();
