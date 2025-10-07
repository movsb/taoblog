class API {
	static async listFiles() {
		const rsp = await fetch('/api/files');
		const json = await rsp.json();
		return json;
	}
	static async deleteFile(path) {
		return fetch(path, { method: 'DELETE'});
	}
}

class UI {
	constructor() {
		/** @type {HTMLDivElement} */
		this._preview = document.querySelector('.preview');
		/** @type {HTMLImageElement} */
		this._img = this._preview.querySelector('img');
		/** @type {HTMLVideoElement} */
		this._video = this._preview.querySelector('video');

		/** @type {HTMLLIElement} */
		this._lastActive = null;

		this._preview.querySelector('.toggle').addEventListener('click', e => {
			e.preventDefault();
			this._preview.classList.toggle('full');
		});
	}

	async preview(imageURL, videoURL) {
		this._img.style.display = 'none';
		this._video.style.display = 'none';
		if(imageURL) {
			this._img.style.display = 'block';
			this._img.src = imageURL;
		}
		if(videoURL) {
			this._video.style.display = 'block';
			this._video.src = videoURL;
			await this._video.play();
		}
	}

	setBasic(li, file) {
		li._isImage = true;

		/** @type {HTMLFormElement} */
		const form = li.querySelector('form');

		/** @type {HTMLSpanElement} */
		const name = form.querySelector('.name');
		name.textContent = file.Name;

		/** @type {HTMLSpanElement} */
		const size = form.querySelector('.size1');
		size.textContent = file.Size;
	}

	handleSelect = e => {
		const li = e.target.closest('li') ?? li;
		if (li != this._lastActive) {
			this._lastActive && this._lastActive.classList.remove('selected');
			li.classList.add('selected');
			this._lastActive = li;
			this.preview();
		}
	}

	handleViewRaw = li => {
		const { textContent: name } = li.querySelector('.name');
		const path = `/in/${name}`;
		if(li._isImage) {
			this.preview(path, undefined);
		}
		if(li._isVideo) {
			this.preview(undefined, path);
		}
	}
	handleDelete = async li => {
		const path = li._path; // 含查询。
		if(path) {
			await API.deleteFile(path);
			li._valid = false;
		}
	}

	initImage(li, file) {
		li._isImage = true;

		/** @type {HTMLFormElement} */
		const form = li.querySelector('form');

		form.addEventListener('submit', async (e) => {
			e.preventDefault();

			if(!li._valid) {
				/** @type {HTMLInputElement} */
				const { value } = form.elements['q'];
				const rsp = await fetch('/api/image', {
					method: 'POST',
					body: JSON.stringify({
						Name: file.Name,
						Q: +value,
					}),
				});
				const json = await rsp.json();
				const size2 = form.querySelector('.size2');
				size2.textContent = json.Size;
				const path = `/out/${json.Path}?t=${new Date().getTime()}`;
				this.preview(path, undefined);
				li._path = path;
				li._valid = true;
			} else {
				this.preview(li._path);
			}
		});

		form.querySelector('div.video').style.display = 'none';
	}

	initVideo(li, file) {
		li._isVideo = true;

		/** @type {HTMLFormElement} */
		const form = li.querySelector('form');

		form.addEventListener('submit', async (e) => {
			e.preventDefault();
			
			if(!li._valid) {
				const { value } = form.elements['crf'];
				const rsp = await fetch('/api/video', {
					method: 'POST',
					body: JSON.stringify({
						Name: file.Name,
						CRF: - +value,
					}),
				});
				const json = await rsp.json();
				const size2 = form.querySelector('.size2');
				size2.textContent = json.Size;
				const path = `/out/${json.Path}?t=${new Date().getTime()}`;
				this.preview(undefined, path);
				li._path = path;
				li._valid = true;
			} else {
				this.preview(undefined, li._path);
			}
		});

		form.querySelector('div.image').style.display = 'none';
	}

	async listFiles() {
		const files = await API.listFiles();
		/** @type {HTMLTemplateElement} */
		const tmpl = document.querySelector('#file');
		const ol = document.querySelector('#list');

		files.forEach(file => {
			const clone = tmpl.content.cloneNode(true);

			/** @type {HTMLLIElement} */
			const li = clone.firstElementChild;

			this.setBasic(li, file);

			li.addEventListener('click', this.handleSelect);

			li.querySelectorAll('.raw').forEach(button => {
				button.addEventListener('click', (e) => {
					e.preventDefault();
					e.stopPropagation();
					this.handleViewRaw(li);
				});
			});
			li.querySelectorAll('input[type="range"]').forEach(range => {
				range.addEventListener('change', e => {
					li._valid = false;
				});
			});
			li.querySelector('.delete').addEventListener('click', () => {
				this.handleDelete(li);
			});

			/** @type {string} */
			const type = file.Type;
			if(type.startsWith('image/')) {
				this.initImage(li, file);
			} else if (type.startsWith('video/')) {
				this.initVideo(li, file);
			}

			ol.appendChild(li);
		});
	}
}

document.addEventListener('DOMContentLoaded', async () => {
	const ui = new UI();
	await ui.listFiles();
});
