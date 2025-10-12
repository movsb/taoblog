/**
 * @typedef {Object} DraftPost
 * @property {number} id
 * @property {string} title
 * @property {number} date - Unix timestamp
 */

class SidebarManager {
	/**
	 * 
	 * @param {string} selector 
	 * @param {string} expand 
	 * @param {{
	 *  onClick: (id: number) => void
	 *  onCollapse() => void
	 * }} options
	 */
	constructor(selector, expand, options) {
		this._options = options || {};
		/** @type {HTMLDivElement} */
		this._sidebar = document.querySelector(selector);
		/** @type {HTMLImageElement} */
		this._expand = document.querySelector(expand);
		/** @type {HTMLOListElement} */
		this._list = this._sidebar.querySelector('ol');
		/** @type {HTMLTemplateElement} */
		this._template = document.getElementById('draft-template');

		this._expand.addEventListener('click', () => {
			this._options.onCollapse?.();
		});

		this._sidebar.addEventListener('blur', () => {
			this._options.onCollapse?.();
		});

		/** @type {HTMLLIElement | null} */
		this._lastSelected = null;

		this.focus();
	}

	focus() {
		const match = window.matchMedia('(max-width: 599px)').matches;
		if(match) {
			this._sidebar.tabIndex = 0;
			this._sidebar.focus();
		} else {
			this._sidebar.removeAttribute('tabIndex');
		}
	}
	blur() {
		this._sidebar.blur();
	}

	/**
	 * 
	 * @param {DraftPost[]} drafts 
	 */
	init(drafts) {
		drafts.forEach(draft => {
			/** @type {DocumentFragment} */
			const doc = this._template.content.cloneNode(true);
			const li = doc.firstElementChild;
			li.dataset.id = draft.id;
			li.addEventListener('click', this._handleClick);
			const titleSpan = li.querySelector('.title');
			titleSpan.textContent = draft.title;
			titleSpan.title = draft.title;
			const dateSpan = li.querySelector('.date');
			const date = new Date(draft.modified * 1000);
			dateSpan.textContent = date.toLocaleString();
			this._list.appendChild(li);
		});
	}

	/**
	 * 
	 * @param {string} title 
	 * @param {number} updatedAt 
	 */
	handleSaved(id, title, updatedAt) {
		const li = this._list.querySelector(`li[data-id="${id}"]`);
		const titleSpan = li.querySelector('.title');
		titleSpan.textContent = title;
		titleSpan.title = title;
		const dateSpan = li.querySelector('.date');
		const date = new Date(updatedAt * 1000);
		dateSpan.textContent = date.toLocaleString();
	}
	handleDirty(id, dirty) {
		const li = this._list.querySelector(`li[data-id="${id}"]`);
		const titleSpan = li.querySelector('.title');
		titleSpan.classList.toggle('dirty', dirty);
	}

	/**
	 * 
	 * @param {MouseEvent} e 
	 */
	_handleClick = (e) => {
		/** @type {HTMLLIElement} */
		const li = e.currentTarget;

		if (this._lastSelected) {
			if(this._lastSelected == li) { return; }
			this._lastSelected.classList.remove('selected');
		}

		li.classList.add('selected');
		this._lastSelected = li;

		this._options.onClick?.(+li.dataset.id);

		const match = window.matchMedia('(max-width: 599px)').matches;
		if(match) { this._options.onCollapse(); }
	}
}

class EditorAreaManager {
	constructor(selector) {
		/** @type {HTMLDivElement} */
		this._editorArea = document.querySelector(selector);

		/** @type {HTMLIFrameElement | null} */
		this._currentDocument = null;
	}

	showDraft(id) {
		/** @type {HTMLIFrameElement | null} */
		const existed = this._find(id);

		if(this._currentDocument) {
			if(this._currentDocument == existed) {
				return;
			}
			if(this._currentDocument.dataset.dirty == '1') {
				this._currentDocument.style.display = 'none';
			} else {
				this._currentDocument.remove();
			}
			this._currentDocument = null;
		}

		if(existed) {
			existed.style.display = 'block';
			this._currentDocument = existed;
			return;
		}

		const frame = document.createElement('iframe');
		frame.src = `editor?id=${id}`;
		frame.dataset.id = id;
		this._editorArea.appendChild(frame);
		this._currentDocument = frame;
	}

	/**
	 * 
	 * @param {number} id 
	 * @returns {HTMLIFrameElement| null}
	 */
	_find(id) {
		return this._editorArea.querySelector(`iframe[data-id="${id}"]`);
	}

	handleSaved(id, title, updatedAt) { }
	handleDirty(id, dirty) {
		const frame = this._find(id);
		frame.dataset.dirty = dirty ? '1' : '0';
	}
}

const editorAreaManager = new EditorAreaManager('#editor-area');

const sidebarManager = new SidebarManager('#sidebar', 'img.expand', {
	onClick: (id) => {
		console.log('click', id);
		editorAreaManager.showDraft(id);
	},
	onCollapse: () => {
		const w = document.querySelector('.wrapper');
		if(w.classList.toggle('collapsed')) {
			sidebarManager.blur();
		} else {
			sidebarManager.focus();
		}
	},
});

document.addEventListener('DOMContentLoaded', () => {
	sidebarManager.init(drafts);

	window.addEventListener('message', (e) => {
		if(e.data.name == 'saved') {
			const { id, title, updatedAt } = e.data;
			sidebarManager.handleSaved(id, title, updatedAt);
			editorAreaManager.handleSaved(id, title, updatedAt);
			return;
		}
		if(e.data.name == 'dirty') {
			const { id, dirty } = e.data;
			sidebarManager.handleDirty(id, dirty);
			editorAreaManager.handleDirty(id, dirty);
		}
	});
}, {once: true});
