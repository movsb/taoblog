/**
 * @typedef {Object} FileSpec
 * @property {string} path
 * @property {string} type
 * @property {number} size
 * @property {number} time
 */

export class FilesManager {
	constructor(id) {
		if (!id) { throw new Error('无效文章编号。'); }
		this._post_id = +id;
	}

	/**
	 * 编码每一级文件名，保留目录分隔符，避免 #、? 等字符变成 URL 片段或查询。
	 * @param {string} path
	 */
	_fileURL(path) {
		const encodedPath = path.split('/').map(segment => encodeURIComponent(segment)).join('/');
		return `/v3/posts/${this._post_id}/files/${encodedPath}`;
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
		return (await decodeResponse(rsp)).files;
	}

	async delete(path) {
		const url = this._fileURL(path);
		let rsp = await fetch(url, {
			method: 'DELETE',
		});
		return await decodeResponse(rsp);
	}

	async updateCaption(path, caption) {
		const url = `${this._fileURL(path)}:caption`;
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
		const url = this._fileURL(path);
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
				failure(new DOMException('上传已取消。', 'AbortError'));
			});
			xhr.addEventListener('error', ()=>{
				failure(new Error('上传失败。', { cause: xhr }));
			});
			xhr.addEventListener('timeout', ()=>{
				failure(new Error('上传超时。', { cause: xhr }));
			});
			xhr.addEventListener('load', ()=> {
				if(xhr.readyState != XMLHttpRequest.DONE) {
					failure(new Error('上传响应状态异常。', { cause: xhr }));
					return;
				}
				if(xhr.status >= 200 && xhr.status < 300) {
					try {
						success(JSON.parse(xhr.responseText));
					} catch (error) {
						failure(new Error('上传响应格式错误。', { cause: error }));
					}
					return;
				}
				console.log(xhr);
				failure(new Error(`上传失败：HTTP ${xhr.status}`, { cause: xhr }));
			});
			xhr.upload.addEventListener('progress', (e)=> {
				if (e.lengthComputable && e.total > 0) {
					progress?.(Math.round(e.loaded / e.total * 100));
				}
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
