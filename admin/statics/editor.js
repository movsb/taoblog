class PostManagementAPI
{
	constructor() { }

	// 创建一条文章。
	async createPost(source, time) {
		let path = `/v3/posts`;
		let rsp = await fetch(path, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				date: time,
				type: 'tweet',
				status: 'public',
				source: source,
				source_type: 'markdown',
				status: 'public',
			}),
		});
		if (!rsp.ok) {
			throw new Error('发表失败：' + await rsp.text());
		}
		let c = await rsp.json();
		console.log(c);
		return c;
	}

	// 更新/“编辑”一条已有评论。
	// 返回更新后的评论项。
	// 参数：id        - 评论编号
	// 参数：source    - 评论 markdown 原文
	async updatePost(id, modified, source, created) {
		let path = `/v3/posts/${id}`;
		let rsp = await fetch(path, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				post: {
					source_type: 'markdown',
					source: source,
					date: created,
					modified: modified,
				},
				update_mask: 'source,sourceType,created'
			})
		});
		if (!rsp.ok) { throw new Error('更新失败：' + await rsp.text()); }
		let c = await rsp.json();
		console.log(c);
		return c;
	}

	// 文章预览
	async previewPost(id, source) {
		let path = `/v3/posts:preview`;
		let rsp = await fetch(path, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				id: id,
				markdown: source,
			})
		});
		if (!rsp.ok) {
			throw new Error(await rsp.text());
		}
		return await rsp.json();
	}
}

class PostFormUI {
	constructor() {
		this._form = document.querySelector('form');
		this._previewCallbackReturned = true;

		this.editor = new TinyMDE.Editor({
			element: document.querySelector('#editor-container'),
			textarea: document.querySelector('#editor-container textarea'),
		});
	}

	get elemSource()    { return this._form['source'];  }
	get elemTime()      { return this._form['time'];    }
	get elemPreviewContainer() { return this._form.querySelector('#preview-container'); }

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

	set source(v)   { this.elemSource.value = v;        }
	setPreview(v, ok)  {
		if (!ok) {
			this.elemPreviewContainer.innerText = v;
		} else {
			this.elemPreviewContainer.innerHTML = v;
		}
		this._previewCallbackReturned = true;
	}

	/**
	 * @param {any[]} list
	 */
	set files(list) {
		console.log(list);
		let ol = this._form.querySelector('.files .list');
		ol.innerHTML = '';
		list.files.forEach(f => {
			let li = document.createElement('li');
			let name = document.createElement('span');
			name.innerText = f.path;
			li.appendChild(name);
			ol.appendChild(li);
		});
	}

	submit(callback) {
		let submit = document.querySelector('input[type=submit]');
		submit.addEventListener('click', (e) => {
			e.preventDefault();
			e.stopPropagation();
			callback();
		});
	}

	drop(callback) {
		window.addEventListener('dragenter', e => {
			e.stopPropagation();
			e.preventDefault();
		}, false);
		window.addEventListener('dragover', e => {
			e.stopPropagation();
			e.preventDefault();
			e.dataTransfer.dropEffect = 'copy';
		}, false);
		window.addEventListener('drop', e => {
			e.stopPropagation();
			e.preventDefault();
			// TODO 换 items，可以知道是否为目录。
			console.log(e.dataTransfer.files);
			callback(e.dataTransfer.files);
		}, false);
	}

	// debounced
	sourceChanged(callback) {
		let debouncing = undefined;
		if (this.editor) {
			this.editor.addEventListener('change', (e)=>{
				if (this._previewCallbackReturned == false) { return; }
				clearTimeout(debouncing);
				debouncing = setTimeout(() => {
					callback(e.content);
				}, 500);
			});
		} else {
			this.elemSource.addEventListener('input', (e)=>{
				if (this._previewCallbackReturned == false) { return; }
				clearTimeout(debouncing);
				debouncing = setTimeout(() => {
					callback(this.elemSource.value);
				}, 500);
			});
		}
	}
}

class FilesManager {
	constructor(id) {
		if (!id) { throw new Error('无效文章编号。'); }
		this._post_id = +id;
	}
	connect() {
		// 火狐官网文档异常小节明明写了 url 如果不是 ws:// 或者 wss:// 会有异常，结果火狐没抛，
		// 谷歌浏览器抛了，真有你的🔥🦊！
		// https://developer.mozilla.org/en-US/docs/Web/API/WebSocket/WebSocket#url
		const prefix = (location.protocol == 'https:' ? 'wss://' : 'ws://') + location.host;
		this._ws = new WebSocket(`${prefix}/v3/posts/${this._post_id}/files`);
		return new Promise((resolve, reject) => {
			this._ws.onclose = () => { console.log('ws closed'); reject("ws closed"); };
			this._ws.onerror = (e) => { console.log(e); reject(e); };
			this._ws.onmessage = console.log;
			this._ws.onopen = () => resolve(this);
		});
	}
	close() {
		this._ws && this._ws.close();
	}
	
	_promise(data, callback) {
		if (this._ws.readyState != WebSocket.OPEN) {
			throw new Error('WebSocket 连接不正确。');
		}
		this._ws.send(JSON.stringify(data));
		return new Promise((resolve, reject) => {
			this._ws.onmessage = (m) => {
				console.log(m);
				resolve(callback(JSON.parse(m.data)));
				this._ws.onmessage = console.log;
			};
			this._ws.onerror = (e) => {
				console.error(e);
				reject(e);
				this._ws.onerror = alert;
			};
		});
	}
	
	// 列举所有的文件列表。
	// NOTE: 返回的文件用 path 代表 name。
	// 因为后端其实是支持目录的，只是前端上传的时候暂不允许。
	// 用 name 表示 path 容易误解。
	async list() {
		let data = { list_files: {}};
		return this._promise(data, obj => obj?.list_files);
	}

	// 创建一个文件。
	// f: <input type="file" /> 中拿来的文件。
	async create(f) {
		let r = new FileReader();
		r.readAsDataURL(f);
		return new Promise((resolve, reject) => {
			r.onerror = (e) => {
				console.error('读文件失败：', r.error);
				reject(r.error);
			}
			r.onload = async () => {
				const data = {
					write_file: {
						spec: {
							path: f.name,
							mode: 0o644,
							size: f.size,
							time: Math.floor(f.lastModified/1000),
						},
						data: r.result.slice(r.result.indexOf(',')+1),
					},
				};
				try {
					resolve(await this._promise(data, obj => obj.write_file));
				} catch(e) {
					reject(e);
				}
			};
		});
	}
	
	// 删除一个文件。
	async delete(path) {
		
	}
}

let postAPI = new PostManagementAPI();
let formUI = new PostFormUI();
formUI.submit(async () => {
	try {
		let post = undefined;
		if (window._post_id > 0) {
			post = await postAPI.updatePost(window._post_id, _modified, formUI.source, formUI.time);
		} else {
			post = await postAPI.createPost(formUI.source, formUI.time);
		}
		alert('成功。');
		window.location = `/${post.id}/`;
	} catch(e) {
		alert(e);
	}
});
if (typeof _created == 'number') {
	formUI.time = _created*1000;
}
formUI.drop(async files => {
	if (!window._post_id) {
		alert('新建文章暂不支持上传文件，请先发表。');
		return;
	}
	if (files.length <= 0) { return; }
	if (files.length > 1) {
		// TODO 其实接口完全允许多文件上传。
		alert('目前仅支持单文件上传哦。');
		return;
	}

	let fm;
	try {
		fm = new FilesManager(+window._post_id);
		await fm.connect();
	} catch(e) {
		alert(e);
		return;
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
			await fm.create(f);
			alert(`文件 ${f.name} 上传成功。`);
		} catch(e) {
			alert(`文件 ${f.name} 上传失败：${e}`);
			return;
		}
		try {
			let list = await fm.list();
			// 奇怪，不是说 lambda 不会改变 this 吗？为什么变成 window 了……
			// 导致我的不得不用 formUI，而不是 this。
			formUI.files = list;
		} catch(e) {
			alert(e);
		}
		console.log(this);
	});
});
let updatePreview = async (content) => {
	try {
		let rsp = await postAPI.previewPost(+window._post_id, content);
		formUI.setPreview(rsp.html, true);
	} catch (e) {
		formUI.setPreview(e, false);
	}
};
formUI.sourceChanged(async (content) => {
	await updatePreview(content);
});
updatePreview(formUI.source);
(async function() {
	if (!window._post_id) {
		console.log('新建文章不支持查询文件列表。');
		return;
	}
	let fm = new FilesManager(+window._post_id);
	try {
		await fm.connect();
		let list = await fm.list();
		formUI.files = list;
		fm.close();
	} catch(e) {
		alert(e);
	}
})();
