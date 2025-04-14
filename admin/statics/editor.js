class PostFormUI {
	constructor() {
		this._form = document.querySelector('#main');
		this._previewCallbackReturned = true;
		this._files = this._form.querySelector('#files');
		this._users = [];
		this._contentChanged = false;
		
		document.querySelector('#geo_modify').addEventListener('click', (e)=> {
			e.preventDefault();
			navigator.geolocation.getCurrentPosition(
				(p)=> {
					const precision = 1e7;
					const longitude = Math.round(p.coords.longitude * precision) / precision;
					const latitude = Math.round(p.coords.latitude * precision) / precision;
					// 按 GeoJSON 来，经度在前，纬度在后。
					const s = `${longitude},${latitude}`;
					console.log('位置：', s);
					document.querySelector('#geo_location').value = s;
					this.updateGeoLocations(latitude, longitude);
				},
				()=> {
					alert('获取位置失败。');
				},
				{
					enableHighAccuracy: true,
				},
			);
		});

		this.elemStatus.addEventListener('change', ()=> {
			let value = this.elemStatus.value;
			let p = this._form.querySelector('p.status');
			p.classList.remove('status-public', 'status-draft', 'status-partial');
			p.classList.add('status-'+value);
		});

		this.elemSetACL.addEventListener('click', (e)=>{
			e.stopPropagation();
			e.preventDefault();

			const dialog = this.elemACLDialog;
			dialog.showModal();
		});
		this.elemACLDialog.querySelector('button').addEventListener('click', (e)=> {
			e.stopPropagation();
			e.preventDefault();
			this.elemACLDialog.close();
		})

		this.checkBoxTogglePreview.addEventListener('click', e=>{
			this.showPreview(e.target.checked);
		});
		this.checkBoxWrap.addEventListener('click', e=>{
			this.setWrap(e.target.checked);
		});
		this.checkBoxToggleDiff.addEventListener('click', e => {
			this.showDiff(e.target.checked);
		})

		const showPreview = localStorage.getItem('editor-config-show-preview') != '0';
		this.checkBoxTogglePreview.checked = showPreview;
		this.showPreview(showPreview);

		const setWrap = localStorage.getItem('editor-config-wrap') != '0';
		this.checkBoxWrap.checked = setWrap;
		this.setWrap(setWrap);

		window.addEventListener('beforeunload', (e)=>{ return this.beforeUnload(e); });

		if (typeof TinyMDE != 'undefined') {
			this.editor = new TinyMDE.Editor({
				element: document.querySelector('#editor-container'),
				textarea: document.querySelector('#editor-container textarea'),
			});
			this.editorCommands = new TinyMDE.CommandBar({
				element: document.getElementById('command-container'),
				editor: this.editor,
				commands: [
					{
						name: `undo`,
						title: `撤销`,
						innerHTML: `↩️ 撤销`,
						action: editor => {
							let e = editor._stack.undo();
							if (e !== undefined) {
								editor.setContent(e);
							}
						},
					},
					{
						name: `redo`,
						title: `重做`,
						innerHTML: `↪️ 重做`,
						action: editor => {
							let e = editor._stack.redo();
							if (e !== undefined) {
								editor.setContent(e);
							}
						},
					},
					{
						name: `insertImage`,
						title: `上传图片/视频/文件`,
						innerHTML: `⏫ 上传文件`,
						action: editor => {
							let files = document.getElementById('files');
							files.click();
						},
					},
					{
						name: `insertGallery`,
						title: `插入九宫格图`,
						innerHTML: `🧩 插入九宫格图`,
						action: editor => {
							const s = `\n<Gallery>\n\n\n\n</Gallery>\n`;
							editor.paste(s);
						},
					},
					{
						name: `insertTaskItem`,
						title: `插入任务`,
						innerHTML: `☑️ 插入任务`,
						action: editor => {
							editor.paste('- [ ] ');
						},
					},
					{
						name: `blockquote`,
						title: `切换选中文本为块引用`,
						innerHTML: `➡️ 插入块引用`,
					},
					{
						name: `divider`,
						title: `插入当时时间分割线`,
						innerHTML: `✂️ 插入分隔符`,
						action: editor => {
							const date = new Date();
							let formatted = date.toLocaleString().replaceAll('/', '-');
							formatted = `\n--- ${formatted} ---\n\n`;
							editor.paste(formatted);
						},
					},
				],
			});
			class UndoRedoStack {
				constructor(editor) {
					this._undo = [];
					this._redo = [];
					
					this._maxUndo = 100;
					this._debouncing;
					this._ignore = true;
					this._oldest = editor.getContent();

					editor.addEventListener('change', (e) => {
						if (this._ignore) {
							this._ignore = false;
							return;
						}
						clearInterval(this._debouncing);
						this._debouncing = setTimeout(() => {
							this.edit(e.content);
						}, 1000);
					});
				}
				edit(e) {
					this._undo.push(e);
					this._redo = [];
				}
				undo() {
					if (this._undo.length <= 0) { return; }
					let current = this._undo.pop();
					this._redo.push(current);
					let last = this._oldest;
					if (this._undo.length > 0) {
						last = this._undo[this._undo.length - 1];
					}
					this._ignore = true;
					return last;
				}
				redo() {
					if (this._redo.length < 1) { return; }
					let e = this._redo.pop();
					this._undo.push(e);
					this._ignore = true;
					return e;
				}
			}
			this.editor._stack = new UndoRedoStack(this.editor);
		} else {
			const editor = document.querySelector('#editor-container textarea[name=source]');
			editor.style.display = 'block';
		}
	}

	get elemSource()    { return this._form['source'];  }
	get elemTime()      { return this._form['time'];    }
	get elemPreviewContainer()  { return this._form.querySelector('#preview-container'); }
	get elemDiffContainer()     { return this._form.querySelector('#diff-container'); }
	get elemType()      { return this._form['type'];    }
	get elemStatus()    { return this._form['status'];  }
	get elemSetACL()    { return this._form['set-acl']; }
	get elemACLDialog() { return this._form.querySelector("[name='set-acl-dialog']"); }
	get checkBoxTogglePreview()     { return this._form.querySelector('#toggle-preview'); }
	get checkBoxWrap()              { return this._form.querySelector('#toggle-wrap'); }
	get checkBoxToggleDiff()        { return this._form.querySelector('#toggle-diff'); }
	get elemToc()       { return this._form['toc'];     }
	
	get geo() {
		const values = this._form['geo_location'].value.trim().split(',');
		if (values.length == 1 && values[0] == '') {
			return {
				name: '',
				longitude: 0,
				latitude: 0,
			};
		}
		if (values.length != 2) {
			throw new Error('坐标值格式错误。');
		}
		const longitude = parseFloat(values[0]);
		const latitude = parseFloat(values[1]);

		return {
			name: this._form['geo_name'].value,
			longitude: longitude,
			latitude: latitude,
		};
	}
	set geo(g) {
		if (!g) { return; }
		this._form['geo_name'].value = g.name ?? '';
		// 按 GeoJSON 来，经度在前，纬度在后。
		this._form['geo_location'].value = `${g.longitude},${g.latitude}`;
	}

	get usersForRequest() {
		const inputs = this.elemACLDialog.querySelectorAll('.list input[type=checkbox]');
		let users = [];
		inputs.forEach(i => {
			if(i.checked) {
				users.push(+i.value);
			}
		});
		return users;
	}
	set users(users) {
		this._users = users || [];

		const dialog = this.elemACLDialog;
		const list = dialog.querySelector('.list');
		list.innerHTML = '';
		this._users.forEach((u) => {
			const label = document.createElement('label');
			const input = document.createElement('input');
			input.type = 'checkbox';
			input.checked = u.can_read;
			input.value = u.user_id;
			label.appendChild(input);
			const name = document.createTextNode(u.user_name);
			label.appendChild(name);
			const li = document.createElement('li');
			li.appendChild(label)
			list.appendChild(li);
		});
	}

	get source()    { return this.elemSource.value;     }
	get time()      {
		let t = this.elemTime.value;
		let d = new Date(t).getTime() / 1000;
		if (!d) {
			d = new Date().getTime() / 1000;
		}
		d = Math.floor(d);
		return d;
	}
	set time(t)      {
		const convertToDateTimeLocalString = (date) => {
			const year = date.getFullYear();
			const month = (date.getMonth() + 1).toString().padStart(2, "0");
			const day = date.getDate().toString().padStart(2, "0");
			const hours = date.getHours().toString().padStart(2, "0");
			const minutes = date.getMinutes().toString().padStart(2, "0");
		  
			return `${year}-${month}-${day}T${hours}:${minutes}`;
		};
		this.elemTime.value = convertToDateTimeLocalString(new Date(t));
	}
	get type() { return this.elemType.value; }
	set type(t) { this.elemType.value = t; }
	get status() { return this.elemStatus.value; }
	set status(s) { this.elemStatus.value = s; }
	get toc() { return this.elemToc.value; }
	set toc(s) { return this.elemToc.value = s; }


	set source(v)   { this.elemSource.value = v;        }
	setPreview(v, ok)  {
		if (!ok) {
			this.elemPreviewContainer.innerText = v;
		} else {
			this.elemPreviewContainer.innerHTML = v;
		}
		this._previewCallbackReturned = true;
	}
	setDiff(v)  {
		this.elemDiffContainer.innerHTML = v;
	}

	// 会自动去重。
	_addFile(ol, f, append) {
		/**
		 * @param {string} s
		 */
		const encodePathAsURL = s => {
			// https://en.wikipedia.org/wiki/Percent-encoding
			// 只是尽量简单地编码必要的字符，不然会在 Markdown 里面很难看。
			// ! # $ & ' ( ) * + , / : ; = ? @ [ ]
			// 外加 % 空格
			const re = /!|#|\$|&|'|\(|\)|\*|\+|,|\/|:|;|=|\?|@|\[|\]|%| /g;
			return s.replace(re, c => '%' + c.codePointAt(0).toString(16).toUpperCase());
		};
		
		const h2a = (h) => {
			const map = {'&': '&amp;', "'": '&#39;', '"': '&quot;'};
			return h.replace(/[&'"]/g, c => map[c]);
		};

		let li = document.createElement('li');

		let name = document.createElement('span');
		name.innerText = f.path;
		li.appendChild(name);

		let insertButton = document.createElement('button');
		let text = '';
		let editor = this.editor;
		let insert = '';
		if (/^image\//.test(f.type)) {
			text = '🏞️';
			insert = `![](${encodePathAsURL(f.path)})\n`;
		} else if (/^video\//.test(f.type)) {
			text = '🎬';
			insert = `<video controls src="${h2a(encodePathAsURL(f.path))}"></video>\n`;
		} else if (/^audio\//.test(f.type)) {
			text = '🎵';
			insert = `<audio controls src="${h2a(encodePathAsURL(f.path))}"></audio>\n`;
		} else {
			text = '🔗';
			insert = `[${f.path}](${encodePathAsURL(f.path)})\n`;
		}
		insertButton.innerText = text;
		insertButton.addEventListener('click', (e) => {
			e.preventDefault();
			e.stopPropagation();
			editor.paste(insert);
		});
		li.appendChild(insertButton);

		if (append) {
			ol.appendChild(li);
		} else {
			ol.prepend(li);
		}

		return li;
	}

	/**
	 * @param {any[]} list
	 */
	set files(list) {
		console.log(list);
		let ol = this._form.querySelector('.files .list');
		ol.innerHTML = '';
		list.forEach(f => this._addFile(ol, f, true));
	}
	filesChanged(callback) {
		this._files.addEventListener('change', (e)=> {
			console.log('选中文件列表：', this._files.files);
			callback(this._files.files);
		});
	}

	static UploadingFile = class {
		constructor(formUI, spec) {
			const list = formUI._form.querySelector('.files .uploading');

			list.querySelectorAll(`li`).forEach(li=> {
				if(li._path == spec.path) {
					li.remove();
				}
			});

			this.li = formUI._addFile(list, spec, false);
			this.li._path = spec.path;

			this.name = spec.path;
			this.span = this.li.querySelector('span');
		}
		set progress(v) {
			this.span.innerText = `${this.name}(${v}%)`;
		}
	}

	tmpFile(spec) {
		return new PostFormUI.UploadingFile(this, spec);
	}

	submit(callback) {
		let submit = document.querySelector('input[type=submit]');
		submit.addEventListener('click', (e) => {
			e.preventDefault();
			e.stopPropagation();
			submit.disabled = true;
			callback(()=>{
				submit.disabled = false;
			});
		});
	}

	// debounced
	sourceChanged(callback) {
		let debouncing = undefined;
		if (this.editor) {
			this.editor.addEventListener('change', (e)=>{
				this._contentChanged = true;
				if (this._previewCallbackReturned == false) { return; }
				clearTimeout(debouncing);
				debouncing = setTimeout(() => {
					callback(e.content);
				}, 1500);
			});
		} else {
			this.elemSource.addEventListener('input', (e)=>{
				this._contentChanged = true;
				if (this._previewCallbackReturned == false) { return; }
				clearTimeout(debouncing);
				debouncing = setTimeout(() => {
					callback(this.elemSource.value);
				}, 500);
			});
		}
	}

	beforeUnload(e) {
		if(this._contentChanged) {
			e.preventDefault();
			return '';
		}
	}

	async updateGeoLocations(latitude, longitude) {
		const loading = this._form.querySelector('#geo_location_loading');
		loading.classList.add('icon-loading');
		try {
			const api = `/v3/utils/geo/resolve?latitude=${latitude}&longitude=${longitude}`;
			const rsp = await fetch(api);
			if (!rsp.ok) {
				let exception = await rsp.json();
				try {
					exception = JSON.parse(exception);
					exception = exception.message ?? exception;
				}
				catch {}
				throw exception;
			}
			const { names } = await rsp.json();
			const datalist = this._form.querySelector('.geo datalist');
			datalist.innerHTML = '';
			(names || []).forEach(name => {
				const option = document.createElement('option');
				option.value = name;
				datalist.appendChild(option);
			});
		} finally {
			loading.classList.remove('icon-loading');
		}
	}

	showPreview(show) {
		this.elemPreviewContainer.style.display = show ? 'block' : 'none';
		localStorage.setItem('editor-config-show-preview', show?'1':'0');
	}
	setWrap(wrap) {
		const editorContainer = this._form.querySelector('#editor-container');
		const diffContainer = this._form.querySelector('#diff-container');

		if(wrap) {
			editorContainer.classList.remove('no-wrap');
			diffContainer.classList.remove('no-wrap');
		} else {
			editorContainer.classList.add('no-wrap');
			diffContainer.classList.add('no-wrap');
		}
		localStorage.setItem('editor-config-wrap', wrap?'1':'0');
	}
	showDiff(show) {
		this.elemDiffContainer.style.display = show ? 'block' : 'none';
	}
}

class FilesManager {
	constructor(id) {
		if (!id) { throw new Error('无效文章编号。'); }
		this._post_id = +id;
	}
	// 列举所有的文件列表。
	// NOTE: 返回的文件用 path 代表 name。
	// 因为后端其实是支持目录的，只是前端上传的时候暂不允许。
	// 用 name 表示 path 容易误解。
	async list() {
		const url = `/v3/posts/${this._post_id}/files`;
		let rsp = await fetch(url);
		if (!rsp.ok) {
			throw new Error(`获取列表失败：`, rsp.statusText);
		}
		rsp = await rsp.json();
		return rsp.files;
	}

	// 创建一个文件。
	// f: <input type="file" /> 中拿来的文件。
	async create(f, progress) {
		return new Promise((success, failure) => {
			let form = new FormData();
			form.set(`spec`, JSON.stringify({
				path: f.name,
				mode: 0o644,
				size: f.size,
				time: Math.floor(f.lastModified/1000),
			}));
			form.set(`data`, f)

			let xhr = new XMLHttpRequest();
			xhr.open('POST', `/v3/posts/${this._post_id}/files`);
			xhr.addEventListener('abort', ()=>{
				failure('xhr: aborted');
			});
			xhr.addEventListener('error', ()=>{
				failure('xhr: error');
			});
			xhr.addEventListener('load', ()=> {
				if(xhr.readyState != XMLHttpRequest.DONE) {
					failure('readystate error');
					return;
				}
				if(xhr.status >= 200 && xhr.status < 300) {
					success();
					return;
				}
				console.log(xhr);
				failure('xhr: unknown');
			});
			xhr.upload.addEventListener('progress', (e)=> {
				progress((e.loaded / e.total * 100).toFixed(0));
			});
			xhr.send(form);
		});
	}
}

let postAPI = new PostManagementAPI();
let formUI = (() => {
	try {
		return new PostFormUI();
	} catch(e) {
		alert('创建表单失败：' + e);
	}
})();
formUI.submit(async (done) => {
	try {
		let p = TaoBlog.posts[TaoBlog.post_id];
		p.metas.geo = formUI.geo;
		p.metas.toc = !!+formUI.toc;
		let post = await postAPI.updatePost({
			id: TaoBlog.post_id,
			date: formUI.time,
			modified: p.modified,
			modified_timezone: TimeWithZone.getTimezone(),
			type: formUI.type,
			status: formUI.status,
			source: formUI.source,
			metas: p.metas,
		}, formUI.usersForRequest);
		formUI._contentChanged = false;
		window.location = `/${post.id}/`;
	} catch(e) {
		alert(e);
	} finally {
		done();
	}
});

(function(){
	let p = TaoBlog.posts[TaoBlog.post_id];
	formUI.time = p.date * 1000;
	formUI.status = p.status;
	formUI.type = p.type;
	if (p.metas && p.metas.geo) {
		formUI.geo = p.metas.geo;
	}
	formUI.toc = p.metas.toc ? "1" : "0";
	formUI.users = p.user_perms || [];
})();

formUI.filesChanged(async files => {
	if (files.length <= 0) { return; }
	Array.from(files).forEach(async f => {
		if (f.size > (10 << 20)) {
			alert(`文件 "${f.name}" 太大，不予上传。`);
			return;
		}
		if (f.size == 0) {
			alert(`看起来不像个文件？只支持文件上传哦。\n\n${f.name}`);
			return;
		}
		let up = formUI.tmpFile({
			path: f.name,
			type: f.type,
		});
		try {
			let fm = new FilesManager(TaoBlog.post_id);
			await fm.create(f, (p)=> {
				console.log(f.name, `${p}%`);
				up.progress = p;
			});
			// alert(`文件 ${f.name} 上传成功。`);
			// 奇怪，不是说 lambda 不会改变 this 吗？为什么变成 window 了……
			// 导致我的不得不用 formUI，而不是 this。
			formUI.files = await fm.list();
		} catch(e) {
			alert(`文件 ${f.name} 上传失败：${e}`);
			return;
		}
	});
});
let updatePreview = async (content) => {
	try {
		let rsp = await postAPI.previewPost(TaoBlog.post_id, content);
		formUI.setPreview(rsp.html, true);
		formUI.setDiff(rsp.diff);
	} catch (e) {
		formUI.setPreview(e, false);
	}
};
formUI.sourceChanged(async (content) => {
	await updatePreview(content);
});
updatePreview(formUI.source);
(async function() {
	try {
		let fm = new FilesManager(TaoBlog.post_id);
		formUI.files = await fm.list();
	} catch(e) {
		alert(e);
	}
})();
