const commentHTML = function(){/*
<div id="comments">
<!--评论标题 -->
<div id="comment-title">
	文章评论
	<span class="item"><span class="total">0</span></span>
	<a class="post-comment item pointer" onclick="comment.reply_to(0)">发表评论</a>
	<span class="right item login-panel">
		<a class="sign-in pointer" onclick="comment.login()">登录</a>
		<span class="sign-out"><span class="login-name"></span>(<a class="pointer" onclick="comment.logout()">登出</a>)</span>
	</span>
</div>
<!-- 评论列表  -->
<ol id="comment-list"></ol>
<!-- 评论功能区  -->
<div class="comment-func">
	<div>还没有用户发表过评论，我要<a class="post-comment pointer" onclick="comment.reply_to(0)">发表评论</a>。</div>
</div>
<!-- 评论框 -->
<div id="comment-form-div">
	<div class="no-sel nc drag-header">
		<div title="隐藏" class="close" onclick="comment.hide();">×</div>
		<div class="comment-title">
			<span id="comment-title-status">编辑评论</span>
		</div>
	</div>
	<form id="comment-form">
		<div class="content-area">
			<textarea class="overlay" id="comment-content" name="source" wrap="on" required></textarea>
			<div class="overlay" id="comment-preview" style="display: none;"></div>
		</div>
		<div class="fields">
			<input type="text" name="author" placeholder="昵称" required>
			<input type="email" name="email" placeholder="邮箱(不公开)" required>
			<input type="url" name="url" placeholder="网站(可不填)">
			<input type="submit" id="comment-submit" value="发表评论">
			<div class="field">
				<label><input type="checkbox" id="comment-wrap-lines" checked>自动折行</label>
			</div>
			<div class="field">
				<label><input type="checkbox" id="comment-show-preview">显示预览</label>
			</div>
		</div>
	</form>
</div>
*/}.toString().slice(14,-3);

class CommentAPI
{
	constructor(postID) {
		this._postID = postID;
	}

	_normalize(c) {
		if (c.id == undefined) {
			throw "评论编号无效。";
		}

		c.id = +c.id;
		c.post_id = +c.post_id;

		c.parent = +(c.parent ?? 0);
		c.root = +(c.root ?? 0);

		c.author = c.author ?? '';
		c.email = c.email ?? '';
		c.url = c.url ?? '';
		c.ip = c.ip ?? '';
		c.source_type = c.source_type ?? 'plain';
		c.source = c.source || (c.source_type == 'plain' ? c.content : c.source);

		c.date = +(c.date ?? 0);
		c.modified = +(c.modified ?? 0);
		c.date_fuzzy = c.date_fuzzy ?? '';
		c.date_timezone = c.date_timezone ?? '';
		c.modified_timezone = c.modified_timezone ?? '';

		c.user_id = c.user_id ?? 0;
		c.geo_location = c.geo_location ?? '';
		c.can_edit = c.can_edit ?? false;
		c.avatar = +(c.avatar ?? 0);

		return c;
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
		let c = await rsp.json();
		return this._normalize(c);
	}

	// 更新/“编辑”一条已有评论。
	// 返回更新后的评论项。
	// 参数：id        - 评论编号
	// 参数：source    - 评论 markdown 原文
	async updateComment(id, modified, source) {
		let path = `/v3/comments/${id}`;
		let rsp = await fetch(path, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				comment: {
					source_type: 'markdown',
					source: source,
					modified: modified.time,
					modified_timezone: modified.zone,
				},
				update_mask: 'source,sourceType,modified'
			})
		});
		if (!rsp.ok) { throw new Error('更新失败：' + (await rsp.json()).message); }
		let c = await rsp.json();
		return this._normalize(c);
	}

	// 返回头像链接。
	avatarURLOf(ephemeral) {
		return `/v3/avatar/${ephemeral}`;
	}

	// 删除一条评论。
	async deleteComment(id) {
		let path = `/v3/comments/${id}`;
		let rsp = await fetch(path, { method: 'DELETE' });
		if (!rsp.ok) {throw new Error(await rsp.text()); }
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
				post_id: postID,
			})
		});
		if (!rsp.ok) {
			throw new Error((await rsp.json()).message);
		}
		return await rsp.json();
	}
}

// 代表一个用来操作评论项的类（即 #comment-N）。
class CommentNodeUI {
	// 参数 node 可以是 id 或者 html 节点。
	constructor(node_or_id) {
		if (node_or_id instanceof HTMLElement) {
			this._node = node_or_id;
		} else if (typeof node_or_id == 'number') {
			let node = document.querySelector(`#comment-${node_or_id}`);
			if (!node) throw new Error('未找到此评论：' + node.toString());
			this._node = node;
		} else {
			throw new Error('初始参数类型不正确。');
		}
	}

	static createElem(c, gen) {
		let div = document.createElement('div');
		div.innerHTML = gen(c);
		return div.firstElementChild;
	}

	get elemContent()   { return this._node.querySelector(':scope > .comment-content'); }
	get htmlID()        { return this._node.id; }
	get htmlNode()      { return this._node; }
	get replyList()     { return this._node.querySelector(`:scope .comment-replies`); }

	setContent(html) { this.elemContent.innerHTML = html; }
	locate() { this._node.scrollIntoView({behavior: 'smooth'}); }
	remove() { this.htmlNode.remove(); }

	setEdited(edited) {
		const time = this._node.querySelector('time');
		time.classList.toggle('edited', edited);
	}
}

// 预览管理对象。
class CommentPreviewUI {
	constructor(toggleContent) {
		this._generated = false;
		this._toggleContent = toggleContent;
	}

	get checkBox()      { return document.getElementById('comment-show-preview'); }
	get container()     { return document.getElementById('comment-preview'); }

	get checked()       { return this.checkBox.checked; }
	
	on(callback) {
		this.checkBox.addEventListener('click', function() {
			if (this.checked) {
				this.clear();
				this.show(true);
				return callback();
			} else {
				this.show(false);
			}
		}.bind(this));
	}

	setHTML(html)   {
		this.container.innerHTML = html;
		this._generated = true;
	}
	setError(text)  {
		let p = document.createElement('div');
		p.style.color = 'red';
		p.innerText = text;
		this.setHTML(p.outerHTML);
		this._generated = true;
	}
	clear() {
		this.container.innerText = '';
		this._generated = false;
		setTimeout(function() {
			if (!this._generated) {
				this.container.innerText = '请稍等...';
			}
		}.bind(this), 500);
	}
	show(yes) {
		this._toggleContent(!yes);
		this.container.style.display = yes ? 'block' : 'none';
		this.checkBox.checked = yes;
	}
}

class CommentFormUI {
	constructor() {
		this._form = document.getElementById('comment-form');
		this._stashedContent = "";
	}

	get elemAuthor()    { return this._form['author'];  }
	get elemEmail()     { return this._form['email'];   }
	get elemURL()       { return this._form['url'];     }
	get elemSource()    { return this._form['source'];  }

	get author()    { return this.elemAuthor.value;     }
	get email()     { return this.elemEmail.value;      }
	get url()       { return this.elemURL.value;        }
	get source()    { return this.elemSource.value;     }

	set author(v)   { this.elemAuthor.value = v;        }
	set email(v)    { this.elemEmail.value = v;         }
	set url(v)      { this.elemURL.value = v;           }
	set source(v)   { this.elemSource.value = v;        }

	save() {
		let commenter = {
			author: this.author,
			email:  this.email,
			url:    this.url,
		};
		let json = JSON.stringify(commenter);
		window.localStorage.setItem('commenter', json);
	}

	restore() {
		let c = JSON.parse(window.localStorage.getItem('commenter') || '{}');
		this.author = c.author ?? c.name ?? '';
		this.email = c.email ?? '';
		this.url = c.url ?? '';
	}

	// 点击“发表评论”时要做的事儿。
	// NOTE：如果表单内容不合法，不会触发 callback。
	submit(callback) {
		let submit = document.querySelector('#comment-submit');
		submit.addEventListener('click', (e) => {
			e.preventDefault();
			e.stopPropagation();
			if (this._form.reportValidity && !this._form.reportValidity()) {
				console.log('表单内容不合法。');
				return;
			}
			submit.disabled = true;
			callback(()=>{
				submit.disabled = false;
			});
		});
	}

	stashContent() {
		this._stashedContent = this.source;
	}
	popContent() {
		this.source = this._stashedContent;
	}
}

class CommentListUI {
	constructor() {
		// 从 API 获取的总评论数。
		this._count = 0;

		// 所有的原始评论对象。
		// 缓存起来是为了再编辑。
		this._comments = {};
	}

	get comments()  { return this._comments; }

	get root()      { return document.querySelector('#comment-list'); }
	get count()     { return this._count; }
	set count(n)    {
		this._count = n;
		this._updateStats();
	}

	// to: 父评论 ID（0 代表顶级评论）
	_append(to, rawComment, before) {
		let parent = this.root;
		if (to != 0) {
			parent = new CommentNodeUI(to).replyList;
		}
		let elem = CommentNodeUI.createElem(rawComment,
			comment.gen_comment_item.bind(comment) // TODO 这是个全局变量
		);

		if (before) {
			parent.prepend(elem);
		} else {
			parent.appendChild(elem);
		}

		elem.style.display = 'block';

		this._comments[rawComment.id] = rawComment;
	}

	_updateStats() {
		let total = document.querySelector('#comment-title .total');
		total.innerText = this._count;
	}

	// 插入评论列表。
	// 注意：评论必须是严格排序的，否则插入顺序可能乱？？？
	// 更新：插入顺序，新的评论始终在上面，不管是父、子评论。
	insert(comments_or_comment) {
		if (Array.isArray(comments_or_comment)) {
			let comments = comments_or_comment;
			let recurse = (id) => {
				let children = comments.filter((c) => c.parent == id);

				// 新的插前面。
				children.sort((a,b) => (id == 0 ? -1 : +1)*(a.date - b.date));

				children.forEach((c) => {
					this._append(id, c, false);
					recurse(c.id);
				});
			};

			recurse(0);
		} else {
			let comment = comments_or_comment;
			this._append(comment.parent, comment, comment.parent == 0);
			this._count         += 1;
		}

		this._updateStats();
	}

	update(comment) {
		let ui = new CommentNodeUI(comment.id);
		ui.setContent(comment.content);
		ui.setEdited(comment.date != comment.modified);
		this._comments[comment.id] = comment;
	}

	remove(id) {
		let ui = new CommentNodeUI(id);
		this._count--;
		// TODO 不确定是删除了子/顶级评论
		ui.remove();
		delete(this._comments[id]);
		this._updateStats();
	}
}

class CommentManager {
	constructor(postID) {
		this.post_id = postID;

		this.being_replied = 0; // 正在回复的评论。
		this.being_edited = 0; // 正在被编辑的 ID，仅编辑时有效，> 0 时有效

		this.api = new CommentAPI(this.post_id);

		// 预览操作对象。
		this.preview = new CommentPreviewUI((show) => {
			this.showContent(show);
		});

		// 表单管理对象。
		this.form = new CommentFormUI();

		// 列表管理对象
		this.list = new CommentListUI();
	}
	init() {
		let self = this;

		this.form.submit(async (done) => {
			try {
				self.setStates({ submitting: true });
				if (self.being_edited > 0) {
					await self.updateComment();
				} else {
					await self.createComment();
				}
			} catch (e) {
				alert(e);
			} finally {
				self.setStates({ submitted: true });
				done();
			}
		});

		document.getElementById('comment-wrap-lines').addEventListener('click', self.wrapLines.bind(self));
		this.preview.on(this.showPreview.bind(this));

		self.init_drag(document.getElementById('comment-form-div'));

		let debouncing = undefined;
		window.addEventListener('resize', () => {
			clearTimeout(debouncing);
			debouncing = setTimeout(this.keepInside, 100);
		});

		this.preload();

		new TextareaWithTab(this.form.elemSource)

		this.updateLoginName();
	}

	updateLoginName() {
		const nickName = TaoBlog.getNickname();
		if(nickName != '') {
			const elem = document.querySelector('#comments .login-name');
			elem.textContent = nickName;
		}
	}

	preload() {
		let comments = TaoBlog.comments;
		this.list.count = comments.length;
		for (let i=0; i<comments.length; i++) {
			comments[i] = this.api._normalize(comments[i]);
		}
		this.list.insert(comments);
		this.toggle_post_comment_button();
	}

	setContent(value) {
		let content = document.querySelector('#comment-content');
		content.value = value;
	}
	clearContent() {
		this.setContent("");
	}
	showContent(yes) {
		let elem = document.querySelector('#comment-content');
		elem.style.display = yes ? 'block' : 'none';
		if (yes) elem.focus();
	}

	hide() {
		this.showCommentBox(false);
	}
	// show         是否显示评论框
	// callback     显示/隐藏完成后的回调函数
	// options
	//      allowEditingInfo    是否允许编辑评论者的信息
	showCommentBox(show, callback, options) {
		let self = this;

		let box = document.getElementById('comment-form-div');
		if (!show && (box.style.display == '' || box.style.display == 'none')) {
			return;
		}

		if(show) {
			box.style.display = 'block';
			this.focus();
		} else {
			box.style.display = 'none';
		}

		if (show) {
			if (typeof options != 'object') {
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
			const inputs = (selectors => selectors.map(selector => document.querySelector(`#comment-form .fields ${selector}`)))([
				'input[name=author]',
				'input[name=email]',
				'input[name=url]',
			]);
			let allowEditingInfo = options.allowEditingInfo ?? true;
			inputs.forEach(input => input.style.display = allowEditingInfo ? 'block' : 'none');

			// 编辑框初始值
			// 设置已保存的作者/邮箱/网址,其实只需要在页面加载完成后设置一次即可，是嘛？
			{
				this.form.restore();

				inputs[0].disabled = false;

				// 新发表，使用登录用户信息。
				if (this.being_edited <= 0) {
					const userID = TaoBlog.getUserID();
					const nickname = TaoBlog.getNickname();
					if (userID > 0 && nickname != ``) {
						this.form.author = nickname;
						this.form.email = 'unused@example.com';
						this.form.url = '';
						inputs[1].style.display = 'none';
						inputs[2].style.display = 'none';
						inputs[0].disabled= true;
					}
				}

				// 其它时候（未提交之前）不应该修改编辑的内容
				if (this.being_edited > 0) {
					this.setContent(this.list.comments[this.being_edited].source);
				}
			}

			let onEsc = function (e) {
				if (e.key === 'Escape') {
					self.hide();
					window.removeEventListener('keyup', onEsc);
				}
			};

			// 按 ESC 关闭（隐藏）窗口。
			window.addEventListener('keyup', onEsc);
		}
	}
	toggle_post_comment_button(show) {
		let root = document.querySelector('#comments');
		if (this.list.count > 0) {
			root.classList.remove('no-comments');
		} else {
			root.classList.add('no-comments');
		}
	}

	locate(id) {
		let ui = new CommentNodeUI(id);
		ui.locate();
		history.replaceState(null, '', `#${ui.htmlID}`);
	}
	gen_comment_item(cmt) {
		// 把可能的 HTML 特殊字符转义以作为纯文本嵌入到页面中。
		// 单、双引号均没必要转换，任何时候都不会引起歧义。
		const h2t = (h) => {
			const map = {'&': '&amp;', '<': '&lt;', '>': '&gt;'};
			return h.replace(/[&<>]/g, c => map[c]);
		};
		// 转义成属性值。
		// 两种情况：手写和非手写。
		// 手写的时候知道什么时候需要把值用单、双引号包起来，跟本函数无关。
		// 如果是构造 HTML，则（我）总是放在单、双引号中，所以 < > 其实没必要转义，
		// 而如果可能不放在引号中，则需要转义。' " 则总是需要转义。
		// 试了一下在火狐中执行 temp0.setAttribute('title', 'a > b')，不管是查看或者编辑，都没被转义。
		// https://mina86.com/2021/no-you-dont-need-to-escape-that/
		const h2a = (h) => {
			const map = {'&': '&amp;', "'": '&#39;', '"': '&quot;'};
			return h.replace(/[&'"]/g, c => map[c]);
		};

		let loggedin = cmt.ip != '';
		let date = new TimeWithZone(cmt.date, cmt.date_timezone);

		// 登录后可以显示评论者的详细信息
		let info = '';
		if (loggedin) {
			info = `编号：${cmt.id}
用户：${cmt.user_id}
邮箱：${cmt.email}
地址：${cmt.ip}
位置：${cmt.geo_location}
`;
		}

		let urlContent = '';
		if (typeof cmt.url == 'string' && cmt.url.length) {
			let url = cmt.url;
			if (!url.match(/^https?:\/\//i)) {
				url = `http://${url}`;
			}
			try {
				let parsed = new URL(url);
				urlContent = `<span class="home"><a rel="nofollow" target="_blank" href="${h2a(url)}">${h2t(parsed.origin)}</a></span>`;
			} catch (e) {
				console.log(e);
			}
		}

		const isAuthor = cmt.user_id == TaoBlog.posts[+TaoBlog.post_id].user_id;

		let html = `
<li style="display: none;" class="comment-li" id="comment-${cmt.id}">
	<div class="comment-avatar">
		<a href="#comment-${cmt.id}" onclick="comment.locate(${cmt.id});return false;">
			<img src="${this.api.avatarURLOf(cmt.avatar)}" width="48px" height="48px" title="${h2a(info)}" loading=lazy>
		</a>
	</div>
	<div class="comment-meta">
		<span class="nickname${isAuthor ? " author" : ""}">${h2t(cmt.author)}</span>
		${urlContent}
		<time class="date${cmt.date!=cmt.modified?' edited':''}" datetime="${date.toJSON()}" data-timezone="${date.zone}" data-unix="${date.time}">${cmt.date_fuzzy}</time>
	</div>
	${cmt.source_type === 'markdown'
				? `<div class="comment-content html-content reset-list-style-type">${cmt.content}</div>`
				: `<div class="comment-content reset-list-style-type">${h2t(cmt.content)}</div>`}
	<div class="toolbar no-sel">
		<a class="" onclick="comment.reply_to(${cmt.id});return false;">回复</a>
		<a class="edit-comment ${cmt.can_edit ? 'can-edit' : ''}" onclick="comment.edit(${cmt.id});return false;">编辑</a>
		<a class="delete-comment" onclick="confirm('确定要删除？') && comment.delete_me(${cmt.id});">删除</a>
	</div>
	<ol class="comment-replies" id="comment-reply-${cmt.id}"></ol>
</li>
`;

		return html;
	}
	reply_to(p) {
		if (this.being_edited > 0) {
			this.being_edited = -1;
			this.form.popContent();
		}
		this.being_replied = +p;
		this.move_to_center();
		this.preview.show(false);
		this.showCommentBox(true, () => this.focus(), {
			allowEditingInfo: true,
		});
	}
	edit(c) {
		if (this.being_replied > -1) {
			this.being_replied = -1;
			this.form.stashContent();
		}
		this.being_edited = c;
		this.move_to_center();
		this.preview.show(false);
		this.showCommentBox(true, ()=>this.focus(), {
			allowEditingInfo: false,
		});
	}
	focus() {
		document.querySelector('#comment-content').focus();
	}
	move_to_center() {
		let div = document.querySelector('#comment-form-div');
		let ww = window.innerWidth;
		let wh = window.innerHeight;
		let cw = getComputedStyle(div)['width'];
		let ch = getComputedStyle(div)['height'];
		let ew = /\d+%/.test(cw) ? parseInt(cw) / 100 * ww : parseInt(cw);
		let eh = /\d+%/.test(ch) ? parseInt(ch) / 100 * wh : parseInt(ch);
		let left = (ww - ew) / 2, top = (wh - eh) / 2;
		div.style.left = `${left}px`;
		div.style.top = `${top}px`;
		console.log({ ww, wh, cw, ch, ew, eh, left, top });
	}
	keepInside() {
		let div = document.querySelector('#comment-form-div');
		let ww = window.innerWidth;
		let wh = window.innerHeight;
		let cw = getComputedStyle(div)['width'];
		let ch = getComputedStyle(div)['height'];

		// 移动设备谁没事儿拖来拖去？🤔
		if (/\d+%/.test(cw) || /\d+%/.test(ch)) { return; }
		let ew = parseInt(cw), eh = parseInt(ch);
		let left = parseInt(div.style.left), top = parseInt(div.style.top);

		if (!(left<0 || top<0 || left+ew>ww || top+eh>wh)) {
			return;
		}

		// NOTE：left & top 两次被调整，仍然可能超出。
		const padding = 10;
		left = Math.max(left,   padding          );
		top  = Math.max(top,    padding          );
		left = Math.min(left,   ww - ew - padding);
		top  = Math.min(top,    wh - eh - padding);
		
		div.style.left = `${left}px`;
		div.style.top = `${top}px`;
	}
	// https://www.w3schools.com/howto/howto_js_draggable.asp
	init_drag(elmnt) {
		// console.log('init_drag');
		let pos1 = 0, pos2 = 0, pos3 = 0, pos4 = 0;
		let dragElem = elmnt.getElementsByClassName("drag-header");
		if (!dragElem) { dragElem = elmnt; }
		else { dragElem = dragElem[0]; }
		dragElem.onmousedown = dragMouseDown.bind(this);
		// console.log(dragElem);

		function dragMouseDown(e) {
			e = e || window.event;
			e.preventDefault();
			// get the mouse cursor position at startup:
			pos3 = e.clientX;
			pos4 = e.clientY;
			document.onmouseup = closeDragElement.bind(this);
			// call a function whenever the cursor moves:
			document.onmousemove = elementDrag.bind(this);
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

			this.keepInside();
		}
	}
	async delete_me(id) {
		try {
			await this.api.deleteComment(id);
			this.list.remove(id);
			this.toggle_post_comment_button();
		} catch (e) {
			alert(e);
		}
	}

	formData() {
		return {
			post_id: this.post_id,
			source_type: 'markdown',
			parent: this.being_replied, // 更新时没用这个字段
			author: this.form.author,
			email: this.form.email,
			url: this.form.url,
			source: this.form.source,
		};
	}
	async updateComment() {
		let { source } = this.formData();
		let id = this.being_edited;
		let raw = this.list.comments[id];

		let updated = await this.api.updateComment(id, new TimeWithZone(raw.modified), source);
		this.list.update(updated);

		this.clearContent();
		this.hide();
		this.preview.show(false);

		return updated;
	}
	setStates(states) {
		let submitButton = document.querySelector('#comment-submit');

		if (states.submitting) {
			submitButton.setAttribute('disabled', 'disabled');
			submitButton.value = '提交中...';
		}
		if (states.submitted) {
			submitButton.value = '发表评论';
			submitButton.removeAttribute('disabled');
		}
	}
	async createComment() {
		let body = this.formData();
		
		body.date_timezone = TimeWithZone.getTimezone();

		let cmt = await this.api.createComment(body);
		this.list.insert(cmt);
		this.toggle_post_comment_button();

		this.hide();
		this.clearContent();
		this.preview.show(false);
		this.form.save();

		return cmt;
	}
	wrapLines() {
		let checkBox = document.getElementById('comment-wrap-lines');
		let textarea = document.getElementById('comment-content');
		textarea.wrap = checkBox.checked ? "on" : "off";
	}
	async showPreview() {
		let source = document.getElementById('comment-form')['source'].value;
		try {
			let rsp = await this.api.previewComment(+this.post_id, source);
			this.preview.setHTML(rsp.html);
		} catch (e) {
			this.preview.setError('预览失败：' + e);
		}
	}
	// TODO 登录功能从评论中移除。
	// 因为文章也是可以在登录后展示编辑按钮的。
	// 登录操作不再仅限于评论区。
	async login() {
		try {
			let wa = new WebAuthn();
			await wa.login();
			document.body.classList.add('signed-in');
			this.updateLoginName();
		} catch(e) {
			if (e instanceof DOMException && ["NotAllowedError", "AbortError"].includes(e.name)) {
				console.log('已取消操作。');
				return;
			}
			alert(e);
		}
	}
	async logout() {
		try {
			let path = `/admin/logout`;
			let rsp = await fetch(path, { method: 'POST'});
			if (!rsp.ok) { throw new Error(await rsp.text()); }
			document.body.classList.remove('signed-in');
		} catch (e) {
			alert('登出失败：' + e);
			return;
		}
	}
}

// 有些扩展脚本会同时处理评论内容中的数据，需要尽早（在 DOMContentLoaded 之前）完成。
// 在主题模板渲染函数里面被调用。
window.__initComments = () => {
	const articles = document.getElementsByTagName('article');
	// 只在单篇文章页显示的时候启用评论功能。
	if(articles.length != 1) {
		return;
	}

	const article = articles[0];
	article.insertAdjacentHTML('beforeend', commentHTML);

	let comment = new CommentManager(TaoBlog.post_id);
	window.comment = comment; // 全局变量，供 HTML 事件处理器使用
	comment.init();
};
