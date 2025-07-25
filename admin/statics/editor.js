class FileList extends HTMLElement {
	static FileItem = class extends HTMLElement {
		connectedCallback() {
			this.innerHTML = `
<div>
	<div class="preview">
		<img>
	</div>
	<div class="info">
		<div class="path"></div>
		<div class="details">
			<span class="progress"></span>
			<span class="size"></span>
		</div>
	</div>
</div>`;

			this._path = this.querySelector('.path');
			this._progress = this.querySelector('.progress');
			this._img = this.querySelector('img');
			this._size = this.querySelector('.size');
		}

		_formatFileSize(bytes) {
			if (bytes === 0) return '0 B';
			const units = ['B', 'KB', 'MB', 'GB', 'TB'];
			const k = 1024;
			const i = Math.floor(Math.log(bytes) / Math.log(k));
			const size = bytes / Math.pow(k, i);
			return size.toFixed(2).replace(/\.00$/, '') + ' ' + units[i];
		}

		get spec() { return this._spec; }
		set spec(spec) {
			this._spec = spec;
			this._path.innerText = this._spec.path;
			this._path.title = this._spec.path;
			this._size.innerText = `大小：${this._formatFileSize(this._spec.size)}`;
			// TODO: 这里是硬编码的。
			this._img.src = `/${TaoBlog.post_id}/${encodeURIComponent(this._spec.path)}`;
		}

		set finished(b) {
			this._progress.innerText = '';
		}

		set progress(v) {
			const s = v == 100 ? '处理中……' : `上传中：${v}%`;
			this._progress.innerText = s;
		}

		getInsertionText() {
			const _encodePathAsURL = (s) => {
				// https://en.wikipedia.org/wiki/Percent-encoding
				// 只是尽量简单地编码必要的字符，不然会在 Markdown 里面很难看。
				// ! # $ & ' ( ) * + , / : ; = ? @ [ ]
				// 外加 % 空格
				const re = /!|#|\$|&|'|\(|\)|\*|\+|,|\/|:|;|=|\?|@|\[|\]|%| /g;
				return s.replace(re, c => '%' + c.codePointAt(0).toString(16).toUpperCase());
			}
				
			const _escapeAttribute = (h) => {
				const map = {'&': '&amp;', "'": '&#39;', '"': '&quot;'};
				return h.replace(/[&'"]/g, c => map[c]);
			}

			const getInsertionText = () => {
				const f = this._spec;
				const pathEncoded = _encodePathAsURL(f.path);

				if (/^image\//.test(f.type)) {
					return `![](${pathEncoded})\n`;
				} else if (/^video\//.test(f.type)) {
					return `<video controls src="${_escapeAttribute(pathEncoded)}"></video>\n`;
				} else if (/^audio\//.test(f.type)) {
					return `<audio controls src="${_escapeAttribute(pathEncoded)}"></audio>\n`;
				} else {
					return `[${f.path}](${pathEncoded})\n`;
				}
			};

			return getInsertionText();
		}
	}

	connectedCallback() {
		this.innerHTML = `
			<ol class="new"></ol>
			<ol class="old"></ol>
		`
		this._old = this.querySelector('ol.old');
		this._new = this.querySelector('ol.new');

		this._selected = [];

		this.addEventListener('click', (e) => {
			const target = e.target;
			const li = target.closest('li');
			if (!li) { return; }
			li.classList.toggle('selected');

			this._selected = this.querySelectorAll('li.selected file-list-item');
			console.log('选中的文件：', this._selected);
			if (this._onSelectionChange) {
				this._onSelectionChange(this._selected);
			}
		});
	}

	clearSelection() {
		this._selected = [];
		this.querySelectorAll('li.selected').forEach(li => {
			li.classList.remove('selected');
		});
		this._onSelectionChange && this._onSelectionChange(this._selected);
	}

	get selected() {
		return this._selected;
	}

	onSelectionChange(callback) {
		this._onSelectionChange = callback;
	}

	addNew(spec) {
		this._new.querySelectorAll('file-list-item').forEach(fi => {
			if (fi._spec.path == spec.path) {
				fi.remove();
			}
		});

		return this._insert(this._new, spec);
	}

	removeFile(fi) {
		const li = fi.closest('li');
		li.remove();
	}

	_insert(list, spec) {
		const li = document.createElement('li');
		const fi = new FileList.FileItem();
		li.appendChild(fi);
		list.appendChild(li);
		fi.spec = spec;
		return fi;
	}

	// 如果同一张图片上传了两遍，则可能由前一张图片被转换了格式，导致出现两条记录。
	// 所以不能直接改 spec，需要去重。
	updateSpec(fi, spec) {
		const fis = this.querySelectorAll('file-list-item');
		fis.forEach(f => {
			if (f._spec.path == spec.path && f != fi) {
				f.remove();
			}
		});
		fi.spec = spec;
	}

	set files(list) {
		this._old.innerHTML = '';
		list.forEach(f => {
			this._insert(this._old, f);
		});
	}
}
customElements.define('file-list', FileList);
customElements.define('file-list-item', FileList.FileItem);

class PostFormUI {
	constructor() {
		this._form = document.querySelector('#main');
		this._previewCallbackReturned = true;
		this._files = this._form.querySelector('#files');
		this._users = [];
		this._contentChanged = false;

		this._fileManagerDialog = document.querySelector('dialog[name="file-manager"]');
		this._fileList = this._fileManagerDialog.querySelector('file-list');
		this._fileList.onSelectionChange( selected => {
			this._fileManagerDialog.querySelector('.insert').disabled = selected.length <= 0;
			this._fileManagerDialog.querySelector('.delete').disabled = selected.length <= 0;
			console.log('选中改变。');
		});
		this._fileManagerDialog.querySelector('.insert').addEventListener('click', (e) => {
			const selected = this._fileList.selected;
			if (selected.length <= 0) { return; }
			selected.forEach(fi => {
				const text = fi.getInsertionText();
				if (this.editor) {
					this.editor.paste(text);
				} else {
					this.elemSource.value += text;
				}
				this._fileList.clearSelection();
			});
		});
		this._fileManagerDialog.querySelector('.delete').addEventListener('click', async (e) => {
			const selected = this._fileList.selected;
			if (selected.length <= 0) { return; }
			selected.forEach(async fi => {
				const path = fi.spec.path;
				console.log('删除文件：', path);
				try {
					await new FilesManager(TaoBlog.post_id).delete(path);
					this._fileList.removeFile(fi);
				} catch(e) {
					alert('删除文件失败：' + e);
					return;
				}
				this._fileList.clearSelection();
			});
		});
		this._fileManagerDialog.querySelector('.upload').addEventListener('click', (e) => {
			this._files.click();
		});
		
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

		let geoInputDebounceTimer = undefined;
		this._form['geo_location'].addEventListener('input', (e)=> {
			clearTimeout(geoInputDebounceTimer);
			geoInputDebounceTimer = setTimeout(async ()=> {
				try {
					const geo = this.geo;
					await this.updateGeoLocations(geo.latitude, geo.longitude);
				} catch(e) {
					console.log('获取地理位置失败：' + e);
					return;
				}
			}, 500);
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
		this.elemIndent.addEventListener('click', ()=>{
			this.setAutoIndent(this.autoIndent);
		})

		const showPreview = localStorage.getItem('editor-config-show-preview') != '0';
		this.checkBoxTogglePreview.checked = showPreview;
		this.showPreview(showPreview);

		const setWrap = localStorage.getItem('editor-config-wrap') != '0';
		this.checkBoxWrap.checked = setWrap;
		this.setWrap(setWrap);

		window.addEventListener('beforeunload', (e)=>{ return this.beforeUnload(e); });

		let lastCategoryIndex = this.elemCategory.selectedIndex;
		this.elemCategory.addEventListener('change', (e) => {
			const value = e.target.value;
			if (value == '-1') {
				this.elemNewCatDialog.showModal();
			} else {
				lastCategoryIndex = e.target.selectedIndex;
				// console.log('选择分类：', lastCategoryIndex);
			}
		});
		this.elemNewCatDialog.addEventListener('close',()=>{
			this.elemCategory.selectedIndex = lastCategoryIndex;
			console.log('新建分类对话框关闭：', lastCategoryIndex);
		});
		this.elemNewCatDialog.querySelector('.cancel').addEventListener('click', (e)=> {
			e.preventDefault();
			console.log('新建分类对话框取消');
			this.elemNewCatDialog.close();
		});
		this.elemNewCatDialog.querySelector('form').addEventListener('submit', async (e)=> {
			console.log('新建分类对话框提交');
			e.stopPropagation();
			e.preventDefault();

			try {
				let cat = await PostManagementAPI.createCategory({
					name: this.elemNewCatDialog.querySelector('input[name=name]').value.trim(),
				});

				const option = document.createElement('option');
				option.value = cat.id;
				option.innerText = cat.name;

				this.elemCategory.insertBefore(
					option,
					this.elemCategory.querySelector('.insert-before'),
				);
				lastCategoryIndex = option.index

				this.elemNewCatDialog.close();
				return;
			} catch(e) {
				alert('创建分类失败：' + e);
				return;
			}
		})

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
						name: `fileManager`,
						title: `上传图片/视频/文件`,
						innerHTML: `📄 文件管理`,
						action: editor => {
							const d = this._fileManagerDialog;
							d.inert = true;
							d.show();
							d.inert = false;
							// 如果需要打开的时候自动弹出选择文件对话框，则取消注释下面的代码。
							// if(this._fileList.selected.length <= 0) {
							// 	this._files.click();
							// }
						},
					},
					{
						name: `insertTaskItem`,
						title: `插入任务`,
						innerHTML: `☑️ 任务`,
						action: editor => {
							editor.paste('- [ ] ');
						},
					},
					{
						name: `blockquote`,
						title: `切换选中文本为块引用`,
						innerHTML: `➡️ 块引用`,
					},
					{
						name: `divider`,
						title: `插入当时时间分割线`,
						innerHTML: `✂️ 分隔符`,
						action: editor => {
							const date = new Date();
							let formatted = date.toLocaleString().replaceAll('/', '-');
							formatted = `\n--- ${formatted} ---\n\n`;
							editor.paste(formatted);
						},
					},
				],
			});
		} else {
			const editor = document.querySelector('#editor-container textarea[name=source]');
			editor.style.display = 'block';
		}

		this.sourceChanged(c => this.updatePreview(c));
		setTimeout(() => this.updatePreview(this.source), 0);
	}

	async updatePreview(content) {
		if (!this.checkBoxTogglePreview.checked) {
			return;
		}
		try {
			let rsp = await PostManagementAPI.previewPost(TaoBlog.post_id, content);
			this.setPreview(rsp.html, true);
			this.setDiff(rsp.diff);
		} catch (e) {
			this.setPreview(e, false);
		}
	};

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
	get elemTop()       { return this._form['top'];     }
	get elemIndent()    { return this._form['auto-indent']; }
	get elemCategory()  { return this._form['category']; }
	get elemNewCatDialog() { return document.body.querySelector('[name=new-category-dialog]'); }
	
	get geo() {
		const private_ = this._form['geo_private'].checked;
		const values = this._form['geo_location'].value.trim().split(',');
		if (values.length == 1 && values[0] == '') {
			return {
				name: '',
				longitude: 0,
				latitude: 0,
				'private': private_,
			};
		}
		if (values.length != 2) {
			throw new Error('坐标值格式错误。');
		}
		if (values[0]=='' || values[1]=='') {
			throw new Error('坐标值格式错误。');
		}
		const longitude = parseFloat(values[0]);
		const latitude = parseFloat(values[1]);

		return {
			name: this._form['geo_name'].value,
			longitude: longitude,
			latitude: latitude,
			'private': private_,
		};
	}
	set geo(g) {
		if (!g) { return; }
		this._form['geo_name'].value = g.name ?? '';
		// 按 GeoJSON 来，经度在前，纬度在后。
		if (g.longitude > 0 && g.latitude > 0) {
			this._form['geo_location'].value = `${g.longitude},${g.latitude}`;
		}
		this._form['geo_private'].checked = !!g['private'];
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
			const seconds = date.getSeconds().toString().padStart(2, "0");
		  
			return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}`;
		};
		this.elemTime.value = convertToDateTimeLocalString(new Date(t));
	}
	get type() { return this.elemType.value; }
	set type(t) { this.elemType.value = t; }
	get status() { return this.elemStatus.value; }
	set status(s) { this.elemStatus.value = s; }
	get toc() { return this.elemToc.value; }
	set toc(s) { return this.elemToc.value = s ? "1": "0"; }
	get top()   { return this.elemTop.value == 1; }
	set top(v)  { this.elemTop.value = v ? "1": "0"; }
	get autoIndent() { return this.elemIndent.checked; }
	set autoIndent(v) {
		this.elemIndent.checked = v;
		this.setAutoIndent(v);
	}


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

	/**
	 * @param {any[]} list
	 */
	set files(list) {
		console.log(list);
		this._fileList.files = list;
	}

	filesChanged(callback) {
		this._files.addEventListener('change', (e)=> {
			console.log('选中文件列表：', this._files.files);
			callback(this._files.files);
		});
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
		if (show) { this.updatePreview(this.source); }
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
	setAutoIndent(b) {
		this.elemPreviewContainer.classList.toggle('auto-indent', b);
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

	async delete(path) {
		const url = `/v3/posts/${this._post_id}/files/${encodeURIComponent(path)}`;
		let rsp = await fetch(url, {
			method: 'DELETE',
		});
		if (!rsp.ok) {
			throw new Error(`文件删除失败：`, rsp.statusText);
		}
	}

	// 创建一个文件。
	// f: <input type="file"> 中拿来的文件。
	async create(f, options, progress) {
		let dimension = await FilesManager.detectImageSize(f);
		dimension.width > 0 && console.log(`文件尺寸：`, f.name, dimension);

		let form = new FormData();
		form.set(`spec`, JSON.stringify({
			path: f.name,
			mode: 0o644,
			size: f.size,
			time: Math.floor(f.lastModified/1000),
			type: f.type, // 其实不应该上传，后端计算更靠谱。

			meta: {
				width: dimension.width,
				height: dimension.height,
			},
		}));

		form.set(`data`, f)

		form.set(`options`, options ? JSON.stringify(options) : '{}');

		return new Promise((success, failure) => {
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
					success(JSON.parse(xhr.responseText));
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

	// 通过浏览器快速判断图片文件的尺寸。
	// f: <input type="file"> 中拿来的文件。
	// 不会抛异常。
	static detectImageSize(f) {
		return new Promise((success, failure) => {
			const isImage = /^image\//.test(f.type);
			const isVideo = /^video\//.test(f.type);

			if (!isImage && !isVideo) {
				return success({ width: 0, height: 0 });
			}

			const url = URL.createObjectURL(f);

			let revoke = (data)=> {
				URL.revokeObjectURL(url);
				return success(data);
			}

			if (isImage) {
				const img = new Image();
				img.addEventListener('load', () => {
					return revoke({ width: img.naturalWidth, height: img.naturalHeight });
				});
				img.addEventListener('error', () => {
					return revoke({ width: 0, height: 0 });
				});
				img.src = url;
			} else if (isVideo) {
				const video = document.createElement('video');
				video.preload = 'metadata';
				video.onloadedmetadata = ()=> {
					return revoke({ width: video.videoWidth, height: video.videoHeight });
				};
				video.onerror = ()=> {
					return revoke({ width: 0, height: 0});
				};
				video.src = url;
			} else {
				revoke({ width: 0, height: 0});
			}
		});
	}
}

document.addEventListener('DOMContentLoaded', () => {

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
		p.metas.text_indent = formUI.autoIndent;
		let post = await postAPI.updatePost(
			{
				id: TaoBlog.post_id,
				date: formUI.time,
				modified: p.modified,
				modified_timezone: TimeWithZone.getTimezone(),
				type: formUI.type,
				status: formUI.status,
				source: formUI.source,
				metas: p.metas,
				top: formUI.top,
				category: +formUI.elemCategory.value,
			},
			{
				users: formUI.usersForRequest,
			},
		);
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
	formUI.autoIndent = !!p.metas.text_indent;
	formUI.top = p.top;
	formUI.users = p.user_perms || [];
})();

formUI.filesChanged(async files => {
	if (files.length <= 0) { return; }

	// 提示是否需要保留图片位置信息。
	let haveImageFiles = Array.from(files).some(f => /^image\//.test(f.type));
	let keepPos = true;
	if (haveImageFiles) {
		keepPos = confirm('选中的文件包含图片，是否需要保留图片的位置信息（如果有）？');
	}

	Array.from(files).forEach(async f => {
		if (f.size > (10 << 20)) {
			alert(`文件 "${f.name}" 太大，不予上传。`);
			return;
		}
		if (f.size == 0) {
			alert(`看起来不像个文件？只支持文件上传哦。\n\n${f.name}`);
			return;
		}

		try {
			let fm = new FilesManager(TaoBlog.post_id);
			const fi = formUI._fileList.addNew({
				path: f.name,
				size: f.size,
				type: f.type,
			});
			let rsp = await fm.create(f, {
				drop_gps_tags: !keepPos,
			}, (p)=> {
				console.log(f.name, `${p}%`);
				fi.progress = p;
			});
			// 可能会自动转换格式，所以用更新后的文件名。
			formUI._fileList.updateSpec(fi, rsp.spec);
			// fi.spec = rsp.spec;
			fi.finished = true;
			// alert(`文件 ${f.name} 上传成功。`);
			// 奇怪，不是说 lambda 不会改变 this 吗？为什么变成 window 了……
			// 导致我的不得不用 formUI，而不是 this。
			// formUI.files = await fm.list();
		} catch(e) {
			alert(`文件 ${f.name} 上传失败：${e}`);
			return;
		}
	});
});
(async function() {
	try {
		let fm = new FilesManager(TaoBlog.post_id);
		formUI.files = await fm.list();
	} catch(e) {
		alert(e);
	}
})();

});
