class MyFileList extends HTMLElement {
	static FileItem = class extends HTMLElement {
		constructor() {
			super();
			this.innerHTML = `
<div>
	<div class="preview">
	</div>
	<div class="info">
		<div class="path"></div>
		<div class="details">
			<div>
				<span class="progress"></span>
				<span class="size"></span>
			</div>
			<div class="caption">
				<a class="button">添加注释</a>
			</div>
			<div class="message"></div>
			<div class="retry"><a class="button">重试</a></div>
		</div>
	</div>
</div>`;

			this._path = this.querySelector('.path');
			this._progress = this.querySelector('.progress');
			this._size = this.querySelector('.size');
			/** @type {HTMLDivElement} */
			this._message = this.querySelector('.message');
			/** @type {HTMLDivElement} */
			this._retry = this.querySelector('.retry');

			/** @type {File|null} */
			this._file = null;
		}

		_formatFileSize(bytes) {
			if (bytes === 0) return '0 B';
			const units = ['B', 'KB', 'MB', 'GB', 'TB'];
			const k = 1024;
			const i = Math.floor(Math.log(bytes) / Math.log(k));
			const size = bytes / Math.pow(k, i);
			return size.toFixed(2).replace(/\.00$/, '') + ' ' + units[i];
		}

		get options() { return this._options; }
		set options(value) { this._options = value; }
		get file() { return this._file; }
		set file(value) { this._file = value; }
		get spec() { return this._spec; }
		set spec(spec) {
			this._spec = spec;
			this._path.innerText = this._spec.path;
			this._path.title = this._spec.path;
			this._size.innerText = `大小：${this._formatFileSize(this._spec.size)}`;

			const fullPath = `/v3/posts/${TaoBlog.post_id}/files/${encodeURIComponent(this._spec.path)}`;

			const preview = this.querySelector('.preview');
			preview.innerHTML = '';
			let elem = null;

			const type = spec.type ?? '';
			const isImage = /^image\//.test(type);
			const isVideo = /^video\//.test(type);

			if(isImage) {
				const img = document.createElement('img');
				img.src = fullPath;
				elem = img;
			} else if(isVideo) {
				const video = document.createElement('video');0
				video.src = fullPath;
				elem = video;
			} else {
				const img = document.createElement('img');
				img.src = 'file.png';
				elem = img;
			}

			preview.appendChild(elem);
		}

		set finished(b) {
			this._progress.innerText = '';
			this.classList.add('finished');
		}

		set progress(v) {
			const s = v == 100 ? '处理中...' : `上传中：${v}%`;
			this._progress.innerText = s;
		}

		error(message) {
			if (message) {
				this._message.innerText = message;
				this._message.classList.add('error');
				this._message.style.display = 'block';
				this._progress.innerText = '';
				if (this._file) {
					this._retry.style.display = 'block';
				}
			} else {
				this._message.style.display = 'none';
				this._retry.style.display = 'none';
			}
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
			<ol></ol>
		`
		this._list = this.querySelector('ol');

		this._selected = [];

		this.addEventListener('click', (e) => {
			const target = e.target;
			if(target.tagName == 'A' && target.parentElement?.classList.contains('caption')) {
				const fi = target.closest('file-list-item');
				this._onEditCaption && this._onEditCaption(fi);
				return;
			}
			if(target.tagName == 'A' && target.parentElement?.classList.contains('retry')) {
				/** @type {MyFileList.FileItem} */
				const fi = target.closest('file-list-item');
				this._onRetry && this._onRetry(fi.file, fi.options);
				return;
			}

			const li = target.closest('li');
			if (!li) { return; }
			li.classList.toggle('selected');

			this._selected = this.querySelectorAll('li.selected file-list-item');
			console.log('选中的文件：', this._selected);
			if (this._onSelectionChange) {
				this._onSelectionChange(this._selected);
			}
		});

		const sortable = new Sortable(
			this._list,
			{
				handle: '.preview',
				animation: 150,
				onClone: function(event) {
					const old = event.item.firstElementChild;
					const cloned = event.clone.firstElementChild;
					cloned.spec = old.spec;
					console.log('拷贝元素', event);
				},
			},
		)
	}

	selectAll() {
		const items = this.querySelectorAll('li');
		items.forEach(item => item.classList.add('selected'));
		this._selected = this.querySelectorAll('li.selected file-list-item');
		console.log('选中的文件：', this._selected);
		if (this._onSelectionChange) {
			this._onSelectionChange(this._selected);
		}
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
	onEditCaption(callback) {
		this._onEditCaption = callback;
	}
	/**
	 * @param {(file: File, options: {keepPos: Boolean}) => {}} callback 
	 */
	onRetry(callback) {
		this._onRetry = callback;
	}

	// 返回旧的spec、新的fi。
	/**
	 * 
	 * @param {*} spec 
	 * @param {File} file 
	 * @param {{keepPos: Boolean}} options 
	 * @returns {{old: {}, now: MyFileList.FileItem}}
	 */
	addNew(file, spec, options) {
		let old = null;
		this._list.querySelectorAll('file-list-item').forEach(fi => {
			if (fi._spec.path == spec.path) {
				old = fi._spec;
				const li = fi.closest('li');
				li.remove();
			}
		});

		return {
			old: old,
			now: this._insert(this._list, file, spec, options),
		};
	}

	removeFile(fi) {
		const li = fi.closest('li');
		li.remove();
	}

	/**
	 * 
	 * @param {HTMLUListElement|HTMLOListElement} list 
	 * @param {File | null} file 
	 * @param {*} spec 
	 * @returns 
	 */
	_insert(list, file, spec, options) {
		const li = document.createElement('li');
		const fi = new MyFileList.FileItem();
		li.appendChild(fi);
		list.appendChild(li);
		fi.spec = spec;
		fi.file = file;
		fi.options = options;
		return fi;
	}

	// 如果同一张图片上传了两遍，则可能由前一张图片被转换了格式，导致出现两条记录。
	// 所以不能直接改 spec，需要去重。
	updateSpec(fi, spec) {
		const fis = this.querySelectorAll('file-list-item');
		fis.forEach(f => {
			if (f._spec.path == spec.path && f != fi) {
				this.removeFile(f);
			}
		});
		fi.spec = spec;
	}

	set files(list) {
		this._list.innerHTML = '';
		list.forEach(f => {
			const fi = this._insert(this._list, null, f, {});
			fi.finished = true;
		});
	}
}
customElements.define('file-list', MyFileList);
customElements.define('file-list-item', MyFileList.FileItem);

class FileManagerDialog {
	constructor(options) {
		this._dialog = document.querySelector('dialog[name="file-manager"]');
		this._fileList = this._dialog.querySelector('file-list');
		this._dialog.querySelector('.insert').addEventListener('click', e => {
			e.stopPropagation();
			e.preventDefault();
			const selected = this._fileList.selected;
			if (selected.length <= 0) { return; }
			this._onInsert && this._onInsert(selected);
		});
		this._onInsert = options?.onInsert;
		this._dialog.querySelector('.select-none').addEventListener('click', (e)=>{
			e.preventDefault();
			e.stopPropagation();
			this.clearSelection();
		});
		this._dialog.querySelector('.select-all').addEventListener('click', (e)=>{
			e.preventDefault();
			e.stopPropagation();
			this._fileList.selectAll();
		});
	}
	show() {
		this._dialog.inert = true;
		this._dialog.show();
		this._dialog.inert = false;
	}
	showModal() {
		this._dialog.showModal();
	}
	clearSelection() {
		this._fileList.clearSelection();
	}
}

class PostFormUI {
	constructor() {
		this._form = document.querySelector('#main');
		this._previewCallbackReturned = true;
		this._files = this._form.querySelector('#files');
		this._users = [];
		this._contentChanged = false;

		this._fileManagerDialog = document.querySelector('dialog[name="file-manager"]');
		/** @type {MyFileList} */
		this._fileList = this._fileManagerDialog.querySelector('file-list');
		this._fileList.onSelectionChange( selected => {
			this._fileManagerDialog.querySelector('.insert').disabled = selected.length <= 0;
			this._fileManagerDialog.querySelector('.delete').disabled = selected.length <= 0;
			this._fileManagerDialog.querySelector('.select-none').disabled = selected.length <= 0;
			// console.log('选中改变。');
		});
		this._fileList.onEditCaption(async fi => {
			const dialog = this._fileManagerDialog.querySelector('dialog[name="file-source-dialog"]');
			const captionEditor = dialog.querySelector('textarea');
			dialog.querySelector('.save').onclick = async ()=> {
				console.log('点击保存');
				const loading = dialog.querySelector('.status-icon');
				try {
					loading.classList.add('loading');
					const caption = captionEditor.value;
					await new FilesManager(TaoBlog.post_id).updateCaption(fi.spec.path, caption);
					fi.spec.meta.source.caption = caption;
					dialog.close();
				} catch(e) {
					alert(e);
				} finally {
					loading.classList.remove('loading');
				}
			};
			if(!fi.spec.meta.source) { fi.spec.meta.source = {}; }
			captionEditor.value = fi.spec.meta.source.caption ?? '';
			dialog.showModal();
		});
		this._fileList.onRetry(async (file, options) => {
			await this.uploadFile(file, options);
		});


		this._fileManagerDialog.querySelector('.delete').addEventListener('click', async (e) => {
			const selected = this._fileList.selected;
			if (selected.length <= 0) { return; }
			selected.forEach(async fi => {
				const path = fi.spec.path;
				console.log('删除文件：', path);
				try {
					await new FilesManager(TaoBlog.post_id).delete(path);
				} catch(e) {
					if (!(e instanceof Response && e.status == 404)) {
						alert('文件删除失败：' + e);
						return;
					}
				}
				this._fileList.removeFile(fi);
				this._fileList.clearSelection();
			});
		});
		this._fileManagerDialog.querySelector('.upload').addEventListener('click', (e) => {
			this._files.click();
		});
		this._form.querySelector('p.file-manager-button button').addEventListener('click', e => {
			this.showFileManager();
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
			p.classList.remove('status-public', 'status-private', 'status-partial', 'status-draft');
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

		const currentPost = TaoBlog.posts[TaoBlog.post_id];
		console.log(currentPost);
		if(currentPost.source_type == 'markdown') {
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
								this.showFileManager();
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
		} else {
			const editor = document.querySelector('#editor-container textarea[name=source]');
			editor.style.display = 'block';
		}

		const showPreview = localStorage.getItem('editor-config-show-preview') != '0';
		this.checkBoxTogglePreview.checked = showPreview;
		this.showPreview(showPreview);

		const setWrap = localStorage.getItem('editor-config-wrap') != '0';
		this.checkBoxWrap.checked = setWrap;
		this.setWrap(setWrap);

		this.sourceChanged(c => this.updatePreview(c));
		setTimeout(() => this.updatePreview(this.source), 0);
	}

	showFileManager(modal) {
		const dialog = new FileManagerDialog({
			onInsert: (selected) => {
				selected.forEach(fi => {
					if(this.editor.onChange) {
						console.log(fi.spec);
						const isImage = /^image\//.test(fi.spec.type);
						const isVideo = /^video\//.test(fi.spec.type);
						let block = undefined;
						if (isImage) {
							block = {
								type: 'image',
								props: {
									url: fi.spec.path,
								},
							};
						} else if (isVideo) {
							block = {
								type: 'video',
								props: {
									url: fi.spec.path,
								},
							};
						} else {
							block = {
								type: 'file',
								props: {
									url: fi.spec.path,
								},
							};
						}
						if (block) {
							const ref = this.editor.document[this.editor.document.length-1];
							this.editor.insertBlocks([block], ref, 'before');
						}
					} else {
						const text = fi.getInsertionText();
						if (this.editor) {
							this.editor.paste(text);
						} else {
							this.elemSource.value += text;
						}
					}
				});
				dialog.clearSelection();
			},
		});
		modal ? dialog.showModal() : dialog.show();
	}

	async updatePreview(content) {
		if (!this.checkBoxTogglePreview.checked) {
			return;
		}
		try {
			let rsp = await PostManagementAPI.previewPost(TaoBlog.post_id, TaoBlog.posts[TaoBlog.post_id].source_type, content);
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

	get source()    {
		return this.elemSource.value;
	}
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


	set source(v)   {
		this.elemSource.value = v;
	}
	setPreview(v, ok)  {
		if (!ok) {
			this.elemPreviewContainer.innerText = v;
			this.elemPreviewContainer.style.whiteSpace = 'pre';
		} else {
			this.elemPreviewContainer.innerHTML = v;
			this.elemPreviewContainer.style.whiteSpace = '';
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

	/**
	 * 
	 * @param {(file: File[]) => {}} callback 
	 */
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
		if (this.editor) { // tinymde
			this.editor.addEventListener('change', (e)=>{
				this._contentChanged = true;
				if (this._previewCallbackReturned == false) { return; }
				clearTimeout(debouncing);
				debouncing = setTimeout(() => {
					callback(e.content);
				}, 1500);
			});
		} else {
			// textarea
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

	/**
	 * 
	 * @param {File} file 
	 * @param {{keepPos: Boolean}} options 
	 * @returns 
	 */
	async uploadFile(file, options) {
		const f = file;
		if (f.size == 0) {
			alert(`看起来不像个文件？只支持文件上传哦。\n\n${f.name}`);
			return;
		}

		const { old, now } = this._fileList.addNew(f, {
			path: f.name,
			size: f.size,
			type: f.type,
		}, options);

		try {
			if (f.size > (10 << 20)) {
				throw new Error(`文件太大，不予上传。`);
			}

			let fm = new FilesManager(TaoBlog.post_id);
			const meta = {
				source: old?.meta?.source,
			};

			let rsp = await fm.create(f,
				{
					drop_gps_tags: !options.keepPos,
				},
				(p)=> {
					console.log(f.name, `${p}%`);
					now.progress = p;
				},
				meta,
			);
			// 可能会自动转换格式，所以用更新后的文件名。
			this._fileList.updateSpec(now, rsp.spec);
			now.finished = true;
			now.error('');
		} catch(e) {
			// alert(`文件 ${f.name} 上传失败：${e}`);
			now.error(`错误：${e.message ?? e}`);
			return;
		}
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
			throw rsp;
		}
		rsp = await rsp.json();
		return rsp.files;
	}

	async delete(path) {
		const url = `/v3/posts/${this._post_id}/files/${encodeURI(path)}`;
		let rsp = await fetch(url, {
			method: 'DELETE',
		});
		if (!rsp.ok) {
			throw rsp;
		}
	}

	async updateCaption(path, caption) {
		const url = `/v3/posts/${this._post_id}/files/${encodeURI(path)}:caption`;
		let rsp = await fetch(url, {
			method: 'PATCH',
			body: JSON.stringify({ caption }),
		});
		if (!rsp.ok) {
			throw rsp;
		}
	}

	// 创建一个文件。
	// f: <input type="file"> 中拿来的文件。
	// options: 用来指导对上传的文件作如何处理。
	// progress: 进度回调。
	// meta: 额外的元数据。但宽、高会自动设置。
	async create(f, options, progress, meta) {
		let dimension = await FilesManager.detectImageSize(f);
		dimension.width > 0 && console.log(`文件尺寸：`, f.name, dimension);
		console.log('creating:', f);

		meta = meta || {};
		meta.width = dimension.width;
		meta.height = dimension.height;

		let form = new FormData();
		form.set(`spec`, JSON.stringify({
			path: f.name,
			mode: 0o644,
			size: f.size,
			time: Math.floor(f.lastModified/1000),
			type: f.type, // 其实不应该上传，后端计算更靠谱。
			meta: meta,
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
				failure('上传失败。');
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
TaoBlog.formUI = formUI;
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
				source_type: p.source_type,
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
		alert('更新失败：' + e);
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
		keepPos = confirm('保留图片的位置信息（如果有）？');
	}

	Array.from(files).forEach(async f => {
		return await formUI.uploadFile(f, { keepPos });
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
