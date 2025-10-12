/**
 * @typedef {import('../dynamic/script.js')} BUNDLE
 * @typedef {import { Sortable } from "./sortable.js"}
 */

class FilesManager {
	constructor(id) {
		if (!id) { throw new Error('无效文章编号。'); }
		this._post_id = +id;
	}
	// 列举所有的文件列表。
	// NOTE: 返回的文件用 path 代表 name。
	// 因为后端其实是支持目录的，只是前端上传的时候暂不允许。
	// 用 name 表示 path 容易误解。
	/**
	 * 
	 * @returns {Promise<FileSpec[]>}
	 */
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

	/**
	 * 
	 * @param {string} path 
	 */
	async get(path) {
		const url = `/v3/posts/${this._post_id}/files/${encodeURI(path)}`;
		let rsp = await fetch(url);
		if (!rsp.ok) {
			throw rsp;
		}
		return rsp.bytes();
	}

	// 创建一个文件。
	// options: 用来指导对上传的文件作如何处理。
	// meta: 额外的元数据。但宽、高会自动设置。
	/**
	 * 
	 * @param {File} f 
	 * @param {*} options 
	 * @param {(progress: number) => void} progress 
	 * @param {{}} meta 
	 * @returns 
	 */
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
			size: f.size,
			time: Math.floor(f.lastModified/1000),
			type: f.type, // 其实不应该上传，后端计算更靠谱。
			meta: meta,
			parent_path: f.parentPath ?? '',
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
				failure(`HTTP: ${xhr.status}`);
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
		return new Promise((success) => {
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

/**
 * @typedef {Object} FileSpec
 * @property {string} path
 * @property {string} type
 * @property {number} size
 * @property {number} time
 */

class FileItem extends HTMLElement {
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

	// 修改以下属性时注意同步到拖动排序拷贝元素。

	get options() { return this._options; }
	set options(value) { this._options = value; }
	get file() { return this._file; }
	set file(value) { this._file = value; }

	get spec() { return this._spec; }

	/**
	 * 
	 * @param {{}} spec 
	 * @param {boolean} preview 
	 */
	setSpec(spec, preview) {
		if (!spec) { throw new Error('错误的文件元信息。'); }

		/** @type {FileSpec} */
		this._spec = spec;
		this._path.innerText = this._spec.path;
		this._path.title = this._spec.path;
		this._size.innerText = `大小：${this._formatFileSize(this._spec.size)}`;

		const fullPath = `/v3/posts/${TaoBlog.post_id}/files/${encodeURIComponent(this._spec.path)}`;

		const previewContainer = this.querySelector('.preview');
		previewContainer.innerHTML = '';
		let elem = null;

		const type = spec.type ?? '';
		const isImage = /^image\//.test(type);
		const isVideo = /^video\//.test(type);

		if(isImage) {
			const img = document.createElement('img');
			if(preview) img.src = fullPath;
			elem = img;
		} else if(isVideo) {
			const video = document.createElement('video');
			if(preview) video.src = fullPath;
			elem = video;
		} else {
			const img = document.createElement('img');
			if(preview) img.src = 'file.png';
			elem = img;
		}

		elem.title = `文件名：${this._spec.path}`;

		previewContainer.appendChild(elem);
	}

	set finished(b) {
		this._progress.innerText = '';
		this.classList.toggle('finished', b);
	}
	get finished() {
		return this.classList.contains('finished');
	}

	set progress(v) {
		const s = v == 100 ? '处理中...' : `上传中：${v}%`;
		this._progress.innerText = s;
	}

	/**
	 * 
	 * @param {string} message 
	 * @param {boolean} showRetry 
	 */
	error(message, showRetry) {
		if (message) {
			this._message.innerText = message;
			this._message.classList.add('error');
			this._message.style.display = 'block';
			this._progress.innerText = '';
			if (this._file && showRetry) {
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

		const f = this._spec;
		const pathEncoded = _encodePathAsURL(f.path);

		// 通通插入成图片，后端根据 type 自动修改插入类型。
		return `![](${pathEncoded})\n`;
	}
}

class MyFileList extends HTMLElement {
	connectedCallback() {
		this.innerHTML = `<ol></ol>`
		this._list = this.querySelector('ol');

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
				this._onRetry && this._onRetry(fi);
				return;
			}

			const li = target.closest('li');
			if (!li) { return; }
			li.classList.toggle('selected');

			console.log('选中的文件：', this.getSelectedItems());
			if (this._onSelectionChange) {
				this._onSelectionChange(this.getSelectedItems());
			}
		});

		this._sortable = new Sortable(
			this._list,
			{
				handle: '.preview',
				animation: 150,
				onClone: function(event) {
					const old = event.item.firstElementChild;
					/** @type {FileItem} */
					const cloned = event.clone.firstElementChild;
					cloned.setSpec(old.spec, true);
					cloned.options = old.options;
					cloned.file = old.file;
					console.log('拷贝元素', event);
				},
				onEnd: () => {
					this._onUserSorted?.();
				},
			},
		)
	}

	/**
	 * @returns {FileItem[]}
	 */
	getSelectedItems() {
		return this.querySelectorAll('li.selected file-list-item');
	}

	/**
	 * 
	 * @param {(a: FileItem, b: FileItem) => number} cmp 
	 */
	sort(cmp) {
		/** @type {HTMLLIElement[]} */
		const items = Array.from(this._list.children);
		const sorted = items.sort((a,b)=>{
			/** @type {FileItem} */
			const fi1 = a.querySelector('file-list-item');
			/** @type {FileItem} */
			const fi2 = b.querySelector('file-list-item');
			return cmp(fi1, fi2);
		}).map(li => li.dataset.id);
		this._sortable.sort(sorted);
	}

	/**
	 * 选中指定文件路径的文件。
	 * @param {string} path 
	 * @returns {boolean}
	 */
	select(path) {
		// 取消任何选择。
		Array.from(this.items).forEach(fi => {
			const li = fi.closest('li');
			li.classList.remove('selected');
		});

		const anchor = Array.from(this.items).find(fi => fi.spec.path == path);
		if(anchor) {
			anchor.scrollIntoView();

			const li = anchor.closest('li');
			li.classList.add('selected');

			console.log('选中的文件：', this.getSelectedItems());
			if (this._onSelectionChange) {
				this._onSelectionChange(this.getSelectedItems());
			}
		}
	}

	/**
	 * 选中所有文件（不包含隐藏的）。
	 */
	selectAll() {
		const items = this.querySelectorAll('li:not(.hidden)');
		items.forEach(item => item.classList.add('selected'));
		console.log('选中的文件：', this.getSelectedItems());
		if (this._onSelectionChange) {
			this._onSelectionChange(this.getSelectedItems());
		}
	}

	clearSelection() {
		this.querySelectorAll('li.selected').forEach(li => {
			li.classList.remove('selected');
		});
		this._onSelectionChange && this._onSelectionChange([]);
	}

	/**
	 * 注册选中文件变化时的回调函数。
	 * @param {(selected: NodeListOf<FileItem>) => void} callback - 当文件选择状态发生变化时调用，参数为当前被选中的 FileItem 节点列表。
	 */
	onSelectionChange(callback) {
		this._onSelectionChange = callback;
	}
	onEditCaption(callback) {
		this._onEditCaption = callback;
	}
	/**
	 * @param {(fi: FileItem) => {}} callback 
	 */
	onRetry(callback) {
		this._onRetry = callback;
	}
	onUserSorted(callback) {
		this._onUserSorted = callback;
	}

	/**
	 * 
	 * @param {*} spec 
	 * @param {File} file 
	 * @param {{keepPos: Boolean}} options 
	 * @returns {FileItem}
	 */
	addNew(file, spec, options) {
		/** @type {NodeListOf<FileItem>} */
		const items = this._list.querySelectorAll('file-list-item');
		const existed = Array.from(items).filter(fi => fi.spec.path == spec.path);
		if (existed?.length > 0) { return existed[0]; }
		return this._insert(file, spec, options, false);
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
	 * @param {boolean} preview 
	 * @returns 
	 */
	_insert(file, spec, options, preview) {
		const li = document.createElement('li');
		li.dataset.id = spec.path;
		const fi = new FileItem();
		li.appendChild(fi);
		this._list.appendChild(li);
		fi.setSpec(spec, preview);
		fi.file = file;
		fi.options = options;
		return fi;
	}

	// 如果同一张图片上传了两遍，则可能由前一张图片被转换了格式，导致出现两条记录。
	// 所以不能直接改 spec，需要去重。
	/**
	 * 
	 * @param {FileItem} fi 
	 * @param {*} spec 
	 */
	updateSpec(fi, spec) {
		const fis = this.querySelectorAll('file-list-item');
		fis.forEach(f => {
			if (f._spec.path == spec.path && f != fi) {
				this.removeFile(f);
			}
		});
		fi.setSpec(spec, true);
		fi.closest('li').dataset.id = spec.path;
	}

	set files(list) {
		this._list.innerHTML = '';
		list.forEach(f => {
			const fi = this._insert(null, f, {}, true);
			fi.finished = true;
		});
	}

	/**
	 * @returns {NodeListOf<FileItem>}
	 */
	get items() {
		return this._list.querySelectorAll('file-list-item');
	}

	/**
	 * 
	 * @param {(fi: FileItem) => boolean} predicate
	 */
	filterItems(predicate) {
		this.items.forEach(fi => {
			const show = predicate(fi);
			const li = fi.closest('li');
			li.classList.toggle('hidden', !show);
		});
	}
}

customElements.define('file-list', MyFileList);
customElements.define('file-list-item', FileItem);

class FileCreationDialog {
	/**
	 * @typedef {Object} FileCreationDialogOptions
	 * @property {(type: string, path: string, data: string) => Promise<boolean>} onSave
	 * 
	 * @param {FileCreationDialogOptions} options 
	 */
	constructor(options) {
		this._options = options;

		/** @type {HTMLDialogElement} */
		this._dialog = document.querySelector('dialog[name=create-file-dialog]');
		/** @type {HTMLFormElement} */
		this._form = this._dialog.querySelector('form');
		/** @type {HTMLSelectElement} */
		this._type = this._dialog.querySelector('select[name=type]');
		/** @type {HTMLInputElement} */
		this._name = this._dialog.querySelector('input[name=name]');

		this._form.addEventListener('submit', async (e)=>{
			e.preventDefault();

			let path = this._name.value;
			let type = 'text/plain';
			let data = '';

			if(path.trim() == '') {
				alert('文件名不能为空。');
				return;
			}

			switch (this._type.value) {
			default:
				alert('未知文件类型。');
				return
			case 'text':
				type = 'text/plain';
				break;
			case 'table':
				path += '.table';
				data = '<table><tbody><tr><td>&nbsp;</td><td>&nbsp;</td></tr><tr><td>&nbsp;</td><td>&nbsp;</td></tr></tbody></table>';
				break;
			case 'drawio':
				path += '.drawio';
				data = '<mxGraphModel dx="1380" dy="912" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1169" pageHeight="827" math="0" shadow="0"><root><mxCell id="0" /><mxCell id="1" parent="0" /></root></mxGraphModel>';
				break;
			case 'tldraw':
				path += '.tldraw';
				data = '{}';
				break;
			}

			if(await this._options.onSave(type, path, data)) {
				this.close();
			}
		});
	}

	close() {
		this._dialog.close();
	}

	showModal() {
		this._type.value = 'table';
		this._name.value = '';
		this._dialog.showModal();
	}
}

class FileManagerDialog {
	/**
	 * 
	 * @param {{
	 *      onInsert: (selected: NodeListOf<FileItem>) => void,
	 *      onRetryUploadFile: (fi: FileItem) => void,
	 *      onDelete: (selected: NodeListOf<FileItem>) => void,
	 *      onChooseFiles: () => void,
	 *      onCreateFile: (type: string, path: string, data: string) => Promise<boolean>,
	 *      onEditFile: (path: string) => void,
	 *      onRefreshList: () => void,
	 *      onShowUnused: (b: boolean) => void,
	 * }} options 
	 */
	constructor(options) {
		this._options = options;
		/** @type {HTMLDialogElement} */
		this._dialog = document.querySelector('dialog[name="file-manager"]');
		/** @type {MyFileList} */
		this._fileList = this._dialog.querySelector('file-list');

		this._fileCreationDialog = new FileCreationDialog({
			onSave: async (type, path, data) => {
				const success =  await this._options.onCreateFile(type, path, data);
				if(success) {
					this._fileCreationDialog.close();
					this._fileList.select(path);
					this.close();
				}
			},
		});

		this._dialog.querySelector('.insert').addEventListener('click', e => {
			e.stopPropagation();
			e.preventDefault();
			const selected = this._fileList.getSelectedItems();
			if (selected.length <= 0) { return; }
			if(Array.from(selected).some(fi => !fi.finished)) {
				alert('选中了未完成处理的文件。');
				return;
			}
			this._options?.onInsert?.(selected);
		});
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
		this._dialog.querySelector('button.tiled').addEventListener('click', (e)=>{
			e.preventDefault();
			e.stopPropagation();
			this.switchView();
		});
		this._dialog.querySelector('button.refresh').addEventListener('click', e => {
			e.preventDefault();
			e.stopPropagation();
			this._options.onRefreshList?.();
		});
		this._dialog.querySelector('button.filtered').addEventListener('click', e => {
			e.preventDefault();
			e.stopPropagation();
			this._toggleShowUnused(e.target);
		});
		this._dialog.querySelector('.delete').addEventListener('click', ()=>{
			const selected = this._fileList.getSelectedItems();
			this._options?.onDelete?.(selected);
		});
		this._dialog.querySelector('.upload').addEventListener('click', ()=>{
			this._options?.onChooseFiles?.();
		});
		this._dialog.querySelector('button.create').addEventListener('click', () => {
			this._fileCreationDialog.showModal();
		});
		this._dialog.querySelector('button.edit').addEventListener('click', ()=>{
			this._options.onEditFile(this._fileList.getSelectedItems()[0].spec.path);
		});

		this._fileList.onSelectionChange(selected => {
			const btnInsert = this._dialog.querySelector('.insert');
			const btnDelete = this._dialog.querySelector('.delete');
			const btnSelectNone = this._dialog.querySelector('.select-none');
			/** @type {HTMLButtonElement} */
			const btnEdit = this._dialog.querySelector('.edit');

			btnInsert.disabled = selected.length <= 0;
			btnDelete.disabled = selected.length <= 0;
			btnSelectNone.disabled = selected.length <= 0;

			let canEdit = false;
			if (selected.length == 1) {
				const fi = selected[0];
				if(fi.spec.path.endsWith('.table')) {
					canEdit = true;
				}
				if(fi.spec.path.endsWith('.drawio')) {
					canEdit = true;
				}
				if(fi.spec.path.endsWith('.tldraw')) {
					canEdit = true;
				}
			}
			btnEdit.disabled = !canEdit;
		});

		this._fileList.onEditCaption(async fi => {
			/** @type {HTMLDialogElement} */
			const dialog = this._dialog.querySelector('dialog[name="file-source-dialog"]');
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
		this._fileList.onRetry(this._options.onRetryUploadFile);
		this._fileList.onUserSorted(()=>{
			/** @type {HTMLSelectElement} */
			const select = this._dialog.querySelector('select.sort');
			select.selectedIndex = 0;
		});

		this._dialog.querySelector('select.sort').addEventListener('change', e => {
			this._sortFiles();
		});
	}
	close() {
		this._dialog.close();
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
	switchView() {
		const name = 'view-tiled';
		this._fileList.classList.toggle(name);
	}
	/**
	 * 
	 * @param {HTMLButtonElement} btn 
	 */
	_toggleShowUnused(btn) {
		const unused = btn.dataset.unused == '1';
		this._options.onShowUnused(!unused);
		btn.dataset.unused = unused ? '0' : '1';
		btn.textContent = !unused ? '显示全部' : '显示未用';
	}

	updateUsedFilesList() {
		const btn = this._dialog.querySelector('button.filtered');
		const unused = btn.dataset.unused == '1';
		this._options.onShowUnused(unused);
	}

	/**
	 * 
	 * @param {(fi: FileItem) => boolean} predicate
	 */
	filterItems(predicate) {
		this._fileList.filterItems(predicate);
	}

	set files(list) {
		console.log(list);
		this._fileList.files = list;
		this.updateUsedFilesList();
		this._sortFiles();
	}

	showUploadFile() {
		this._options?.onChooseFiles?.();
	}

	/**
	 * 
	 * @param {FileItem} fi 
	 */
	removeFile(fi) {
		return this._fileList.removeFile(fi);
	}

	removeFileByPath(path) {
		const items = Array.from(this._fileList.items);
		const item = items.find(fi => fi.spec.path == path);
		if(item) return this.removeFile(item);
		return false;
	}

	/**
	 * 
	 * @param {File} file 
	 * @param {{
	 *      path: string,
	 *      size: number,
	 *      type: string,
	 * }} spec 
	 */
	createFile(file, spec, options) {
		const ret = this._fileList.addNew(file, spec, options);
		this._sortFiles();
		return ret;
	}

	updateFile(fi, spec) {
		const ret = this._fileList.updateSpec(fi, spec);
		this._sortFiles();
		return ret;
	}

	_sortFiles() {
		const rule = this._dialog.querySelector('select.sort').value;
		/** @type {(a: FileItem, b: FileItem) => number} */
		let cmp = null;
		switch(rule) {
		case 'name_asc':
			cmp = (a, b) => a.spec.path.localeCompare(b.spec.path);
			break;
		case 'name_desc':
			cmp = (a, b) => -(a.spec.path.localeCompare(b.spec.path));
			break;
		case 'time_asc':
			cmp = (a, b) => a.spec.time - b.spec.time;
			break;
		case 'time_desc':
			cmp = (a, b) => -(a.spec.time - b.spec.time);
			break;
		}
		if(!cmp) { return; }
		this._fileList.sort(cmp);
	}
}

class Tab {
	/**
	 * 
	 * @param {HTMLDivElement} dom 
	 * @param {string} name 
	 * @param {{onClose: ()=> void, onSelect: () => void}} options 
	 */
	constructor(name, dom, options) {
		this._options = options;
		this._dom = dom;

		/** @type {HTMLSpanElement} */
		this._name = this._dom.querySelector('.name');
		/** @type {HTMLSpanElement} */
		this._close = this._dom.querySelector('.close');

		this._name.textContent = name;
		this._close.addEventListener('click', (e)=>{
			e.preventDefault();
			e.stopImmediatePropagation();
			this._options.onClose();
		});
		this._dom.addEventListener('click', (e)=>{
			e.preventDefault();
			e.stopImmediatePropagation();
			this._options.onSelect();
		});
	}

	get name() {
		return this._name.textContent;
	}
}

class Editor extends HTMLElement {
	constructor() {
		super();

		/** @type {string} */
		this._name = "";
	}
	
	connectedCallback() { }

	set name(value) {
		this._name = value;
	}
	get name() {
		return this._name;
	}

	/**
	 * 
	 * @param {HTMLElement} editor 
	 */
	embed(editor) {
		this.appendChild(editor);
	}

	/**
	 * @param {boolean} value 
	 */
	set fullscreen(value) {
		this.classList.toggle('fullscreen', value);
	}
}
customElements.define('my-editor', Editor);

class TableEditor extends HTMLElement {
	constructor() {
		super();

		this._table = null;
	}

	async connectedCallback() {
		/** @type {HTMLTemplateElement} */
		const tmpl = document.getElementById('tmpl-table-editor');
		const cloned = tmpl.content.cloneNode(true);
		this.appendChild(cloned.firstElementChild);

		if(typeof JavaScriptTableEditor == 'undefined') {
			const url = 'https://unpkg.com/javascript-table-editor@1.0.7/dist/table.iife.min.js';
			const script = document.createElement('script');
			script.src = url;
			document.head.appendChild(script);
			const timer = setInterval(async ()=>{
				if(typeof JavaScriptTableEditor != 'undefined') {
					clearInterval(timer);
					await this._init();
				}
			}, 1000);
		} else {
			await this._init();
		}
	}

	async _init() {
		const placeholder =  this.querySelector('.placeholder');
		/** @type {EventTarget} */
		const table = new JavaScriptTableEditor.Table(placeholder);
		const content = await this._file.text();
		table.use(content);
		this._table = table;

		this._table.addEventListener('change', () => {
			this._onChange?.(this._table.getContent());
		});

		this.querySelector('.toolbar').addEventListener('click', (e) => {
			if(e.target.tagName == 'BUTTON') {
				const name = e.target.className;
				this._table[name]?.();
			}
		});
	}

	/**
	 * 
	 * @param {(content: string) => void} content 
	 */
	onChange(content) {
		this._onChange = content;
	}

	/**
	 * @param {File} value 
	 */
	set file(value) {
		this._file = value;
	}
}
customElements.define('table-editor', TableEditor);

// [Embed mode](https://www.drawio.com/doc/faq/embed-mode)
class DrawioEditor extends HTMLElement {
	constructor() {
		super();

		/** @type {HTMLIFrameElement} */
		this._iframe = null;
	}

	/**
	 * @param {File} value 
	 */
	set file(value) {
		this._file = value;
	}

	async connectedCallback() {
		const frame = document.createElement('iframe');

		// embed drawio 有个 bug，即便当前是 dark 模式，进入 iframe 仍然是 light，但是切日夜切换却是正常的。
		// 这里通过媒体查询设定一个初始值，设定后仍然能自动日夜切换。
		// ui=dark/light 是查 gpt 得知的，官方文档没有写。
		const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		frame.src = `https://embed.diagrams.net/?configure=1&embed=1&spin=1&libraries=1&proto=json&ui=${isDark ? 'dark' : 'light' }`;
		frame.style.width = '100%';
		frame.style.minHeight = '85dvh';
		frame.style.border = 'none';
		this.appendChild(frame);

		window.addEventListener('message', this._messageHandler);
	}
	disconnectedCallback() {
		window.removeEventListener('message', this._messageHandler);
		console.log('移除事件');
	}

	/**
	 * 
	 * @param {MessageEvent} e 
	 * @returns 
	 */
	_messageHandler = async (e) => {
		console.log('收到消息：', e);
		const m = JSON.parse(e.data);
		if(m.event == 'configure') {
			e.source.postMessage(JSON.stringify({
				action: 'configure',
				config: {
					css: `
					`,
				},
			}), e.origin);
			return;
		}
		if(m.event == 'init') {
			e.source.postMessage(JSON.stringify({
				action: 'load',
				xml: await this._file.text(),
				title: this._file.name,
			}),e.origin);
			return;
		}
		if(m.event == 'save') {
			e.source.postMessage(JSON.stringify({
				action: 'export',
				format: 'svg',
				border: 10,
			}), e.origin);
			return;
		}
		if(m.event == 'export') {
			// console.log(m.data);
			const xmlName = this._file.name;
			const xmlFile = new File([m.xml], xmlName);

			const svgName = xmlName + '.svg';
			const index = m.data.indexOf(';base64,');
			const base64Content = m.data.slice(index + ';base64,'.length);
			const bytes = Uint8Array.fromBase64(base64Content);
			const svgFile = new File([bytes], svgName);

			e.source.postMessage(JSON.stringify({
				action: 'spinner',
				message: '正在保存...',
				show: true,
				enabled: false,
			}), e.origin);

			try {
				await this._onSave(xmlFile, svgFile);
			} finally {
				e.source.postMessage(JSON.stringify({
					action: 'spinner',
					show: false,
				}), e.origin);
			}

			return;
		}
		if(m.event == 'exit') {
			this._onClose();
			return;
		}
	}

	/**
	 * 
	 * @param {(xmlFile: File, svgFile: File)=>Promise<boolean>} callback 
	 */
	onSave(callback) {
		this._onSave = callback;
	}
	onClose(callback) {
		this._onClose = callback;
	}
}
customElements.define('drawio-editor', DrawioEditor);

class TldrawEditor extends HTMLElement {
	constructor() {
		super();
	}

	/**
	 * @param {File} value 
	 */
	set file(value) {
		this._file = value;
	}

	async connectedCallback() {
		if(typeof renderTldrawComponent == 'undefined') {
			const style = document.createElement('link');
			style.rel = 'stylesheet';
			style.href = '/admin/tldraw.min.css';
			document.head.appendChild(style);

			const script = document.createElement('script');
			script.src = '/admin/tldraw.min.js';
			script.type = 'module';
			document.body.appendChild(script);

			const timer = setInterval(async () => {
				if(typeof renderTldrawComponent != 'undefined') {
					clearInterval(timer);
					await this._init();
				}
			}, 1000);
		} else {
			await this._init();
		}
	}

	async _init() {
		const placeholder = document.createElement('div');
		this.appendChild(placeholder);
		renderTldrawComponent(placeholder, {
			snapshotJSON: await this._file.text(),
			saveSnapshot: async (state, light, dark) => {
				return await this._onSave(state, light, dark);
			},
			close: async () => {
				this._onClose();
			},
		});
	}

	/**
	 * 
	 * @param {(state: string, light: Blob, dark: Blob) => Promise<boolean>} callback 
	 */
	onSave(callback) {
		this._onSave = callback;
	}
	onClose(callback) {
		this._onClose = callback;
	}
}
customElements.define('tldraw-editor', TldrawEditor);

class TabsManager {
	/**
	 * 
	 * @param {{
	 *  onClose: (tab: Tab)=>void,
	 *  getFile: (path: string) => Promise<File>,
	 *  saveFiles: (files: [File]) => Promise<boolean>,
	 * }} options 
	 */
	constructor(options) {
		this._options = options;

		/** @type {HTMLDivElement} */
		this._tabs = document.querySelector('#tabs');
		/** @type {HTMLDivElement} */
		this._editors = document.querySelector('#editors');

		/** @type {HTMLTemplateElement} */
		this._tabTmpl = document.querySelector('#tab-template');
		// /** @type {HTMLTemplateElement} */
		// this._editorTmpl = document.querySelector('#editor-template');
	}

	_select(name) {
		Array.from(this._tabs.children).forEach(_div => {
			/** @type {HTMLDivElement} */
			const div = _div;
			/** @type {Tab} */
			const tab = _div._tab;

			div.classList.toggle('selected', tab.name == name);
		});
		Array.from(this._editors.children).forEach(div => {
			if(div instanceof HTMLDivElement && name == '文章') {
				div.classList.add('selected');
				return;
			}

			/** @type {Editor} */
			const editor = div;
			editor.classList.toggle('selected', editor.name == name);
		});
	}

	/**
	 * @returns {boolean}
	 */
	isOpening(name) {
		const found = Array.from(this._tabs.children).find(div => div._tab.name == name);
		return !!found;
	}

	async open(name) {
		if(this.isOpening(name)) {
			this._select(name);
			this._toggleTabs();
			return;
		}

		const cloned = this._tabTmpl.content.cloneNode(true);
		const child = cloned.firstElementChild;
		const tab = new Tab(name, child, {
			onClose: () => {
				this._options.onClose(tab);
			},
			onSelect: () => {
				this._select(name);
			},
		});
		child._tab = tab;
		this._tabs.appendChild(child);
			
		if(name != '文章') {
			const editor = new Editor();
			editor.name = name;
			const embed = await this._openEmbed(name, tab);
			if (embed instanceof TldrawEditor) {
				editor.fullscreen = true;
			}
			if (embed instanceof DrawioEditor) {
				editor.fullscreen = true;
			}
			editor.embed(embed);
			this._editors.appendChild(editor);
		}

		this._select(name);
		this._toggleTabs();
	}

	/**
	 * 
	 * @param {string} path 
	 * @param {Tab} tab 
	 * @returns {Promise<HTMLElement>}
	 */
	async _openEmbed(path, tab) {
		if(path.endsWith('.table')) {
			const file = await this._options.getFile(path);
			const editor = new TableEditor();
			editor.file = file;
			editor.onChange((content) => {
				this._options.saveFiles([new File([content], file.name)]);
			});
			return editor;
		}
		if(path.endsWith('.drawio')) {
			const file = await this._options.getFile(path);
			const editor = new DrawioEditor();
			editor.file = file;
			editor.onSave(async (xmlFile, svgFile) => {
				svgFile.parentPath = xmlFile.name;
				return await this._options.saveFiles([xmlFile, svgFile]);
			});
			editor.onClose(()=>{
				this._options.onClose(tab);
			});
			return editor;
		}
		if(path.endsWith('.tldraw')) {
			const file = await this._options.getFile(path);
			const editor = new TldrawEditor();
			editor.file = file;
			editor.onSave((state, light, dark) => {
				const stateFile = new File([state], path);
				const lightFile = new File([light], path + '.light.svg');
				const darkFile  = new File([dark],  path + '.dark.svg');
				lightFile.parentPath = stateFile.name;
				darkFile.parentPath  = stateFile.name;
				return this._options.saveFiles([stateFile, lightFile, darkFile]);
			});
			editor.onClose(()=>{
				this._options.onClose(tab);
			});
			return editor;
		}

		return document.createElement('div');
	}

	close(name) {
		if(name == '文章') { return; }

		const tab = Array.from(this._tabs.children).find(div => div._tab.name == name);
		if(tab) { tab.remove(); }

		const editor = Array.from(this._editors.children).find(_editor => {
			/** @type {Editor} */
			const editor = _editor;
			return editor.name == name;
		});
		if(editor) { editor.remove(); }

		this._toggleTabs();
	}

	_toggleTabs() {
		const many = this._tabs.children.length > 1;
		this._tabs.classList.toggle('only', !many);
	}
}

class PostFormUI {
	constructor() {
		this._api = new PostManagementAPI();
		/** @type {HTMLFormElement} */
		this._form = document.querySelector('#main');
		/** @type {HTMLDivElement} */
		this._editorContainer = document.querySelector('#editor-container');
		this._previewCallbackReturned = true;
		/** @type {HTMLInputElement} */
		this._files = this._form.querySelector('#files');
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
						if (!(e instanceof Response && e.status == 404)) {
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

		this._form.querySelector('p.file-manager-button .all').addEventListener('click', () => {
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
				this.editorCommands = new TinyMDE.CommandBar({
					element: document.getElementById('command-container'),
					editor: this.editor,
					commands: [
						{
							name: `insertTaskItem`,
							title: `插入任务`,
							innerHTML: `☑️ 任务`,
							action: editor => {
								editor.paste('- [ ] ');
							},
						},
						{
							name: `divider`,
							title: `插入当时时间分割线`,
							innerHTML: `✂️ 分隔符`,
							action: editor => {
								const date = new Date();
								const year = date.getFullYear();
								const month = (date.getMonth() + 1).toString().padStart(2, "0");
								const day = date.getDate().toString().padStart(2, "0");
								const hours = date.getHours().toString().padStart(2, "0");
								const minutes = date.getMinutes().toString().padStart(2, "0");
								const seconds = date.getSeconds().toString().padStart(2, "0");
								
								const formatted = `\n--- ${year}-${month}-${day} ${hours}:${minutes}:${seconds} ---\n\n`;
								editor.paste(formatted);
							},
						},
					],
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

		const submit = this._form.querySelector('input[type=submit]');
		submit.addEventListener('click', async (e)=>{
			e.preventDefault();
			e.stopPropagation();
			await this._handleSubmit(submit);
		});

		/** @type {HTMLInputElement} */
		const fullscreenCheckbox = this._form.elements['fullscreen'];
		const exitDiv = this._editorContainer.querySelector('.exit-fullscreen');
		fullscreenCheckbox.addEventListener('change', e => {
			const checked = fullscreenCheckbox.checked;
			this._editorContainer.classList.toggle('stretch', checked);
			exitDiv.style.display = checked ? 'block' : 'none';
		});
		exitDiv.querySelector('button').addEventListener('click', e => {
			this._editorContainer.classList.remove('stretch');
			exitDiv.style.display = 'none';
			fullscreenCheckbox.checked = false;
		});

		// 这行代码用于允许接收 drop 事件。
		// [HTML Drag and Drop API - Web APIs | MDN](https://developer.mozilla.org/en-US/docs/Web/API/HTML_Drag_and_Drop_API#drop_target)
		document.addEventListener('dragover', e => e.preventDefault());
		document.addEventListener('drop', e => {
			e.preventDefault();
			this._handleSelectFiles(e.dataTransfer.files);
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
			if(wv.width < 500 && wv.height < 500) {
				ec.classList.add('stretch');
				ec.style.height = `${wv.height}px`;
			} else {
				ec.classList.remove('stretch');
				ec.style.removeProperty('height');
			}
		});
	}

	initFrom(p) {
		this.time = p.date * 1000;
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
		this._contentChanged = b;

		if(window.parent != window) {
			window.parent.postMessage({
				name: 'dirty',
				id: TaoBlog.post_id,
				dirty: b,
			}, '*');
		}
	}

	/**
	 * 
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

	async _updatePost() {
		let p = TaoBlog.posts[TaoBlog.post_id];
		p.metas.geo = this.geo;
		p.metas.toc = !!+this.toc;
		p.metas.text_indent = this.autoIndent;

		const newPost = {
			id: TaoBlog.post_id,
			date: this.time,
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
		selected.forEach(fi => {
			const text = fi.getInsertionText();
			if (this.editor) {
				this.editor.paste(text);
			} else {
				// TODO 插入到选中位置。
				this.elemSource.value += text;
			}
		});
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

		autoSave = autoSave && window.parent != window;

		try {
			let rsp = await PostManagementAPI.previewPost(
				id, post.source_type, content, autoSave, post.modified,
			);

			this.setPreview(rsp.html, true);
			this.setDiff(rsp.diff);
			this._updateReferencedFiles(rsp.paths || []);
			this._pathsInUse = rsp.paths || [];
			this._fileManager.updateUsedFilesList();

			if(autoSave) {
				this._setContentChanged(false);

				post.modified = rsp.updated_at;
				post.title = rsp.title

				window.parent.postMessage({
					name: 'saved',
					id: id,
					title: post.title,
					updatedAt: post.modified,
				}, '*');
			}
		} catch (e) {
			this.setPreview(e, false);
		}
	};

	/**
	 * 
	 * @param {string[]} paths 
	 */
	_updateReferencedFiles(paths) {
		const buttonsParent = document.querySelector('p.file-manager-button');
		buttonsParent.querySelectorAll('.file').forEach(b => b.remove());

		// note: 路径是有重复的，需要拖动去重
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
		this._previewCallbackReturned = true;
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
				if (this._previewCallbackReturned == false) { return; }
				clearTimeout(debouncing);
				debouncing = setTimeout(() => {
					callback(e.content);
				}, 1500);
			});
		} else {
			// textarea
			this.elemSource.addEventListener('input', ()=>{
				this._setContentChanged(true);
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
			if (f.size > (10 << 20)) {
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

document.addEventListener('DOMContentLoaded', () => {
	const post = TaoBlog.posts[TaoBlog.post_id];
	const form = new PostFormUI();
	form.initFrom(post);
}, {once: true});
