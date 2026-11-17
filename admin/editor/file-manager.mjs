import { FilesManager } from './files.mjs';

/** @import { FileSpec } from './files.mjs' */

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

		const fullPath = `/${TaoBlog.post_id}/${encodeURIComponent(this._spec.path)}`;

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
				/** @type {FileItem} */
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

export class FileManagerDialog {
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
		// 全屏编辑的时候被盖住了。
		// 全局设置好像问题不大？
		this._dialog.style.zIndex = 999;
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
