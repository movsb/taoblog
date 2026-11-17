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
			const url = 'table.iife.min.js';
			const script = document.createElement('script');
			script.src = url;
			document.head.appendChild(script);
			const timer = setInterval(async ()=>{
				if(typeof JavaScriptTableEditor != 'undefined') {
					clearInterval(timer);
					await this._init();
				}
			}, 200);
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
			this._onChange?.(this._table.getJSON());
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
		if (this._iframe) {
			window.addEventListener('message', this._messageHandler);
			return;
		}
		const frame = document.createElement('iframe');
		this._iframe = frame;

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
		if (e.origin !== 'https://embed.diagrams.net'
			|| !this._iframe || !e.source || e.source !== this._iframe.contentWindow
			|| typeof e.data !== 'string') {
			return;
		}
		console.log('收到消息：', e);
		let m;
		try {
			m = JSON.parse(e.data);
		} catch {
			return;
		}
		if (!m || typeof m !== 'object' || Array.isArray(m) || typeof m.event !== 'string') {
			return;
		}
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
			if (typeof m.xml !== 'string' || typeof m.data !== 'string') { return; }
			const index = m.data.indexOf(';base64,');
			if (index < 0 || !m.data.startsWith('data:image/svg+xml')) { return; }
			const base64Content = m.data.slice(index + ';base64,'.length);
			let bytes;
			try {
				bytes = Uint8Array.fromBase64(base64Content);
			} catch {
				return;
			}
			const xmlName = this._file.name;
			const xmlFile = new File([m.xml], xmlName);

			const svgName = xmlName + '.svg';
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

export class TabsManager {
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
