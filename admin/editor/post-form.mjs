import { FilesManager } from './files.mjs';
import { FileManagerDialog } from './file-manager.mjs';
import { TabsManager } from './tabs.mjs';

/** @import { FileItem } from './file-manager.mjs' */

export class PostFormUI {
	constructor() {
		this._api = new PostManagementAPI();
		/** @type {HTMLFormElement} */
		this._form = document.querySelector('#main');
		/** @type {HTMLDivElement} */
		this._editorContainer = document.querySelector('#editor-container');
		this._previewRequestID = 0;
		this._sourceVersion = 0;
		// 自动保存串行执行，下一次请求必须使用上一次返回的 modified。
		this._autoSaveQueue = Promise.resolve();
		/** @type {HTMLInputElement} */
		this._files = this._form.querySelector('#files');
		/** @type {HTMLButtonElement} */
		this._time = this._form.elements['time'];
		this._users = [];
		this._contentChanged = false;

		/** @type {string[]} */
		this._pathsInUse = [];

		this._fileManager = new FileManagerDialog({
			onInsert: (selected) => {  this._handleInsertFiles(selected); },
			onRetryUploadFile: async (fi) => {
				fi.error('');
				await this.uploadFile(fi.file, fi.options);
			},
			onDelete: async (selected) => {
				selected.forEach(async fi => {
					const path = fi.spec.path;
					console.log('删除文件：', path);
					try {
						await new FilesManager(TaoBlog.post_id).delete(path);
					} catch(e) {
						if (!(e instanceof Error && e.cause instanceof Response && e.cause.status == 404)) {
							alert('文件删除失败：' + e);
							return;
						}
					}
					this._fileManager.removeFile(fi);
					this._tabManager.close(path);
				});
				this._fileManager.clearSelection();
			},
			onChooseFiles: () => { this._files.click(); },
			onCreateFile: async (type, path, data) => {
				const file = new File([data], path, {
					type: type,
					lastModified: Date.now(),
				});
				const spec = await this.uploadFile(file, {
					keepPos: true,
					showError: true,
				});
				if (typeof spec === 'boolean') {
					if (!spec) {
						this._fileManager.removeFileByPath(path);
						return false;
					}
				}
				this._tabManager.open(path);
				return true;
			},
			onEditFile: async (path) => {
				console.log('编辑：' + path);
				await this._tabManager.open(path);
				this._fileManager.close();
			},
			onRefreshList: async () => {
				this._refreshFileList();
			},
			onShowUnused: b => {
				this._fileManager.filterItems(fi => {
					const path = fi.spec.path;
					if(!b) return true;
					return !this._pathsInUse.includes(path);
				});
			},
		});

		this._tabManager = new TabsManager({
			onClose: (tab) => {
				if(tab.name == '文章') {
					console.log('“文章”不能被关闭。');
					return;
				}
				this._tabManager.close(tab.name);
				this._tabManager.open('文章');
			},
			getFile: async (path) => {
				const fm = new FilesManager(TaoBlog.post_id);
				const list = await fm.list();
				const spec = list.find(s => s.path == path);
				const data = await fm.get(path);
				return new File([data], spec.path, {
					lastModified: spec.time,
				});
			},
			/**
			 *
			 * @param {File[]} files
			 * @returns {Promise<boolean>}
			 */
			saveFiles: async (files) => {
				try {
					const fm = new FilesManager(TaoBlog.post_id);
					const promises = [];
					files.forEach(file => {
						const spec = fm.create(file, {}, ()=>{}, {});
						promises.push(spec);
					});
					const specs = await Promise.all(promises);
					console.log('文件更新成功：', specs);
				} catch(e) {
					alert('文件更新失败：' + e);
					return false;
				}
				try {
					await this.updatePreview(this.source, true);
					console.log('预览更新成功');
					return true;
				} catch(e) {
					alert('预览更新失败：' + e);
					return false;
				}
			},
		})

		this._tabManager.open('文章');

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
					/** @type {HTMLInputElement} */
					const loc = document.querySelector('#geo_location');
					const empty = loc.value.length == 0;
					loc.value = s;
					if(empty) { this._form['geo_private'].checked = true; }
					const pg = document.querySelector('p.geo');
					pg.classList.remove('no-data');
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
		this._form['geo_location'].addEventListener('input', ()=> {
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
					element: this._editorContainer,
					textarea: this._editorContainer.querySelector('textarea'),
				});

				const commandContainer = document.querySelector('#command-container');
				commandContainer.querySelector('.insert-task-item').addEventListener('click', (e)=>{
					e.stopPropagation();
					e.preventDefault();
					this.editor.paste('- [ ] ');
				});
				commandContainer.querySelector('.insert-separator').addEventListener('click', (e)=>{
					e.stopPropagation();
					e.preventDefault();

					const date = new Date();
					const year = date.getFullYear();
					const month = (date.getMonth() + 1).toString().padStart(2, "0");
					const day = date.getDate().toString().padStart(2, "0");
					const hours = date.getHours().toString().padStart(2, "0");
					const minutes = date.getMinutes().toString().padStart(2, "0");
					const seconds = date.getSeconds().toString().padStart(2, "0");

					const formatted = `--- ${year}-${month}-${day} ${hours}:${minutes}:${seconds} ---\n\n`;
					this.editor.paste(formatted);
				});
				commandContainer.querySelector('.open-file-manager').addEventListener('click',e=>{
					e.stopPropagation();
					e.preventDefault();
					this.showFileManager();
				});
				this.editor.addEventListener('selection', obj => {
					this.editor.__selection = obj;
				});
			} else {
				const editor = this._editorContainer.querySelector('textarea[name=source]');
				editor.style.display = 'block';
			}
		} else {
			const editor = this._editorContainer.querySelector('textarea[name=source]');
			editor.style.display = 'block';
		}

		const showPreview = localStorage.getItem('editor-config-show-preview') != '0';
		this.checkBoxTogglePreview.checked = showPreview;
		this.showPreview(showPreview);

		const setWrap = localStorage.getItem('editor-config-wrap') != '0';
		this.checkBoxWrap.checked = setWrap;
		this.setWrap(setWrap);

		this.sourceChanged(this._handleSourceChanged);
		setTimeout(() => this.updatePreview(this.source, false), 0);

		this._refreshFileList();
		this._handleWindowResize();

		this._files.addEventListener('change', async (e)=>{
			this._handleSelectFiles(this._files.files);
		});

		this._bindSubmit();

		/** @type {HTMLInputElement} */
		const fullscreenCheckbox = this._form.elements['fullscreen'];
		const fullscreenToolbar = this._editorContainer.querySelector('.fullscreen-toolbar');
		fullscreenCheckbox.addEventListener('change', e => {
			const checked = fullscreenCheckbox.checked;
			this._editorContainer.parentElement.classList.toggle('stretch', checked);
			fullscreenToolbar.style.display = checked ? 'block' : 'none';
		});
		fullscreenToolbar.querySelector('button.files').addEventListener('click', e => {
			this.showFileManager();
		});
		fullscreenToolbar.querySelector('button.exit').addEventListener('click', e => {
			this._editorContainer.parentElement.classList.remove('stretch');
			fullscreenToolbar.style.display = 'none';
			fullscreenCheckbox.checked = false;
		});

		// 这行代码用于允许接收 drop 事件。
		// [HTML Drag and Drop API - Web APIs | MDN](https://developer.mozilla.org/en-US/docs/Web/API/HTML_Drag_and_Drop_API#drop_target)
		document.addEventListener('dragover', e => e.preventDefault());
		document.addEventListener('drop', e => {
			e.preventDefault();
			this._handleSelectFiles(e.dataTransfer.files);
		});

		this._form.elements['time'].addEventListener('click', e => {
			e.preventDefault();
			e.stopPropagation();

			/** @type {HTMLButtonElement} */
			const btn = e.target;
			const tvz = this._getTime();
			new DateTimePicker(btn, tvz.time, tvz.zone, ({unix, timezone})=>{
				this._setTime(new TimeWithZone(unix, timezone));
			});
		});
	}

	_handleSourceChanged = (content) => {
		this.updatePreview(content, true);
	}

	// 小屏幕下使编辑区域占满屏幕。
	_handleWindowResize() {
		if(!('ontouchstart' in window)) { return; }
		const wv = window.visualViewport;
		/** @type {HTMLDivElement} */
		const ec = this._editorContainer;
		wv?.addEventListener('resize', ()=> {
			if(wv.width < 500 && wv.height < 500 && document.activeElement?.closest('#editor-container')) {
				ec.classList.add('stretch');
				ec.style.height = `${wv.height}px`;
			} else {
				ec.classList.remove('stretch');
				ec.style.removeProperty('height');
			}
		});
	}

	initFrom(p) {
		this._setTime(new TimeWithZone(p.date, p.date_timezone));
		this.status = p.status;
		this.type = p.type;
		if (p.metas && p.metas.geo) {
			this.geo = p.metas.geo;
		}
		this.autoIndent = !!p.metas.text_indent;
		this.top = p.top;
		this.users = p.user_perms || [];
	}

	_setContentChanged(b) {
		if (b) { this._sourceVersion++; }
		this._contentChanged = b;

		if(window.parent != window) {
			window.parent.postMessage({
				name: 'dirty',
				id: TaoBlog.post_id,
				dirty: b,
			}, window.location.origin);
		}
	}

	_bindSubmit() {
		const submit = this._form.querySelector('input[type=submit]');
		this._form.addEventListener('submit', async e => {
			e.preventDefault();
			e.stopPropagation();
			if (submit.disabled || !this._form.reportValidity()) { return; }
			await this._handleSubmit(submit);
		});
	}

	/**
	 * @param {HTMLInputElement} submitButton
	 */
	async _handleSubmit(submitButton) {
		try {
			submitButton.disabled = true;
			await this._updatePost();
			this._setContentChanged(false);
			window.location = `/${TaoBlog.post_id}/`;
		} catch(e) {
			submitButton.disabled = false;
			alert('更新失败：' + e);
		}
	}

	/**
	 * 快捷键触发的保存按钮。
	 */
	async triggerSubmit() {
		/** @type {HTMLInputElement} */
		const submit = this._form.querySelector('input[type=submit]');
		if(submit.disabled) { return; }

		this._form.requestSubmit(submit);
	}

	async _updatePost() {
		let p = TaoBlog.posts[TaoBlog.post_id];
		p.metas.geo = this.geo;
		p.metas.toc = !!+this.toc;
		p.metas.text_indent = this.autoIndent;

		const newPost = {
			id: TaoBlog.post_id,
			date: this._getTime().time,
			date_timezone: this._getTime().zone,
			modified: p.modified,
			modified_timezone: TimeWithZone.getTimezone(),
			type: this.type,
			status: this.status,
			source: this.source,
			source_type: p.source_type,
			metas: p.metas,
			top: this.top,
			category: +this.elemCategory.value,
		};

		return await this._api.updatePost(
			newPost,
			{
				users: this.usersForRequest,
				send_user_notify: this._form['send-user-notify'].checked,
			},
		);
	}

	async _refreshFileList() {
		try {
			const fm = new FilesManager(TaoBlog.post_id);
			const list = await fm.list();
			this._fileManager.files = list;
		} catch(e) {
			alert('获取文件列表失败：' + e);
		}
	}

	/**
	 *
	 * @param {File[]} files
	 * @returns
	 */
	async _handleSelectFiles(files) {
		if (files.length <= 0) { return; }

		// 提示是否需要保留图片位置信息。
		let haveImageFiles = Array.from(files).some(f => /^image\//.test(f.type));
		let keepPos = haveImageFiles && confirm('保留图片的位置信息（如果有）？');

		Array.from(files).forEach(async f => {
			return await this.uploadFile(f, { keepPos });
		});
	}

	/**
	 *
	 * @param {NodeListOf<FileItem>} selected
	 */
	_handleInsertFiles(selected) {
		// 一次性插入，避免多次插入时多次改变选区焦点导致插入位置/覆盖错误。
		let insertion = '';
		selected.forEach(fi => { insertion += fi.getInsertionText(); });
		if (this.editor) {
			this.editor.paste(insertion,
				this.editor.__selection?.anchor,
				this.editor.__selection?.focus,
			);
		} else {
			// TODO 插入到选中位置。
			this.elemSource.value += insertion;
		}
		this._fileManager.clearSelection();
	}

	showFileManager() {
		this._fileManager.show();
	}

	/**
	 *
	 * @param {string} content
	 * @param {boolean} autoSave 是否自动保存，即便是，只会在被嵌入的时候有效。
	 */
	async updatePreview(content, autoSave) {
		const id = TaoBlog.post_id;
		const post = TaoBlog.posts[id];
		const requestID = ++this._previewRequestID;
		const sourceVersion = this._sourceVersion;
		const isCurrentSource = () => sourceVersion === this._sourceVersion && content === this.source;
		const isCurrentPreview = () => requestID === this._previewRequestID && isCurrentSource();

		autoSave = autoSave && window.parent != window;

		const update = async () => {
			const rsp = await PostManagementAPI.previewPost(
				id, post.source_type, content, autoSave, post.modified,
			);

			if(autoSave) {
				// 即使用户已继续输入，也要记住这次保存的版本，供下一次保存使用。
				post.modified = rsp.updated_at;
				post.title = rsp.title;
				if (isCurrentSource()) {
					this._setContentChanged(false);
				}

				window.parent.postMessage({
					name: 'saved',
					id: id,
					title: post.title,
					updatedAt: post.modified,
				}, window.location.origin);
			}

			// 新请求或新输入都能使旧响应失效，包括尚处于防抖等待中的输入。
			if (!isCurrentPreview()) { return; }
			this.setPreview(rsp.html, true);
			this.setDiff(rsp.diff);
			this._updateReferencedFiles(rsp.paths || []);
			this._pathsInUse = rsp.paths || [];
			this._fileManager.updateUsedFilesList();
		};

		try {
			if (autoSave) {
				const pending = this._autoSaveQueue.then(update);
				// 单次失败不阻塞后续保存；错误仍由下面的 catch 展示。
				this._autoSaveQueue = pending.catch(() => {});
				await pending;
			} else {
				await update();
			}
		} catch (e) {
			if (isCurrentPreview()) { this.setPreview(e, false); }
		}
	};

	/**
	 *
	 * @param {string[]} paths
	 */
	_updateReferencedFiles(paths) {
		const buttonsParent = document.querySelector('#command-container');
		buttonsParent.querySelectorAll('.file').forEach(b => b.remove());

		// note: 路径是有重复的，需要手动去重
		[...new Set(paths)].forEach(path => {
			if(path.endsWith('.table')) {
				const btn = document.createElement('button');
				btn.type = 'button';
				btn.textContent = `编辑表格：${path}`;
				btn.addEventListener('click', ()=>{
					this._tabManager.open(path);
				});
				btn.classList.add('file');
				buttonsParent.appendChild(btn);
			}
			if(path.endsWith('.drawio')) {
				const btn = document.createElement('button');
				btn.type = 'button';
				btn.textContent = `编辑绘图：${path}`;
				btn.addEventListener('click', ()=>{
					this._tabManager.open(path);
				});
				btn.classList.add('file');
				buttonsParent.appendChild(btn);
			}
			if(path.endsWith('.tldraw')) {
				const btn = document.createElement('button');
				btn.type = 'button';
				btn.textContent = `编辑绘图：${path}`;
				btn.addEventListener('click', ()=>{
					this._tabManager.open(path);
				});
				btn.classList.add('file');
				buttonsParent.appendChild(btn);
			}
		});
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

		if(g.name || g.longitude || g.latitude || !!g['private']) {
			/** @type {HTMLParagraphElement} */
			const p = document.querySelector('p.geo');
			p.classList.remove('no-data');
		}
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

	/**
	 *
	 * @param {TimeWithZone} tvz
	 */
	_setTime(tvz)      {
		const convertToDateTimeLocalString = (date) => {
			const year = date.getFullYear();
			const month = (date.getMonth() + 1).toString().padStart(2, "0");
			const day = date.getDate().toString().padStart(2, "0");
			const hours = date.getHours().toString().padStart(2, "0");
			const minutes = date.getMinutes().toString().padStart(2, "0");
			const seconds = date.getSeconds().toString().padStart(2, "0");

			return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
		};

		const d = new Date(tvz.time*1000);
		let v = `${convertToDateTimeLocalString(d)}`;
		if(tvz.zone != TimeWithZone.getTimezone()) {
			v = `${v} ${tvz.zone}`;
		}
		this._time.textContent = v;
		this._time._tvz = tvz;
	}
	/**
	 *
	 * @returns {TimeWithZone}
	 */
	_getTime() {
		return this._time._tvz;
	}

	get type() { return this.elemType.value; }
	set type(t) { this.elemType.value = t; }
	get status() { return this.elemStatus.value; }
	set status(s) { this.elemStatus.value = s; }
	get toc() { return this.elemToc.value; }
	set toc(s) { this.elemToc.value = s ? "1": "0"; }
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
	}
	setDiff(v)  {
		this.elemDiffContainer.innerHTML = v;
	}

	// debounced
	/**
	 *
	 * @param {(content: string) => void} callback
	 */
	sourceChanged(callback) {
		let debouncing = undefined;
		if (this.editor) { // tinymde
			this.editor.addEventListener('change', (e)=>{
				this._setContentChanged(true);
				clearTimeout(debouncing);
				debouncing = setTimeout(() => {
					callback(e.content);
				}, 1500);
			});
		} else {
			// textarea
			this.elemSource.addEventListener('input', ()=>{
				this._setContentChanged(true);
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
				catch { /* empty */ }
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
		this.elemIndent.closest('label').style.display = show ? 'inline-block' : 'none';
		localStorage.setItem('editor-config-show-preview', show?'1':'0');
		if (show) { this.updatePreview(this.source, false); }
	}
	setWrap(wrap) {
		const diffContainer = this._form.querySelector('#diff-container');

		if(wrap) {
			this._editorContainer.classList.remove('no-wrap');
			diffContainer.classList.remove('no-wrap');
		} else {
			this._editorContainer.classList.add('no-wrap');
			diffContainer.classList.add('no-wrap');
		}
		localStorage.setItem('editor-config-wrap', wrap?'1':'0');
	}
	showDiff(show) {
		this.elemDiffContainer.style.display = show ? 'block' : 'none';
	}
	/**
	 * https://x.com/passluo/status/1967501847457128693
	 * @param {boolean} b
	 */
	setAutoIndent(b) {
		this.elemPreviewContainer.classList.toggle('auto-indent', b);
	}

	/**
	 *
	 * @param {File} file
	 * @param {{keepPos: Boolean, showError: boolean}} options
	 * @returns {Promise<boolean>}
	 */
	async uploadFile(file, options) {
		const f = file;

		if (f.name.trim() == '') {
			alert('无效文件名。');
			return false;
		}

		if (f.size == 0) {
			alert(`看起来不像个文件？只支持文件上传哦。\n\n${f.name}`);
			return false;
		}

		const now = this._fileManager.createFile(f, {
			path: f.name,
			size: f.size,
			type: f.type,
		}, options);

		try {
			if (f.size > (100 << 20)) {
				throw new Error(`文件太大，不予上传。`);
			}

			let fm = new FilesManager(TaoBlog.post_id);
			const meta = {};

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
			this._fileManager.updateFile(now, rsp.spec);
			now.finished = true;
			now.error('');
			return true;
		} catch(e) {
			now.error(`错误：${e.message ?? e}`, true);
			if(options.showError || !this._fileManager._dialog.open) {
				alert('错误：' + e);
			}
			return false;
		}
	}
}
