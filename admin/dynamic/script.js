class WebAuthnAPI {
	constructor(){
		// 非 HTTPS 页面这个类没有被导出（localhost除外）。
		if(typeof PublicKeyCredential == 'undefined' && location.protocol == 'http:' && location.hostname != 'localhost') {
			throw new Error('只能在 HTTPS 安全连接下完成此操作。');
		}
	}

	async beginRegistration() {
		let path = `/admin/login/webauthn/register:begin`;
		let rsp = await fetch(path, {
			method: 'POST',
		});
		if (!rsp.ok) {
			throw new Error('开始注册失败：' + await rsp.text());
		}
		let options = await rsp.json();
		options.publicKey = PublicKeyCredential.parseCreationOptionsFromJSON(options.publicKey);
		return options;
	}
	async finishRegistration(options) {
		let path = `/admin/login/webauthn/register:finish`;
		let rsp = await fetch(path, {
			method: 'POST',
			body: JSON.stringify(options),
		});
		if (!rsp.ok) {
			throw new Error('结束注册失败：' + await rsp.text());
		}
	}
	async beginLogin() {
		let path = `/admin/login/webauthn/login:begin`;
		let rsp = await fetch(path, {
			method: 'POST',
		});
		if (!rsp.ok) {
			throw new Error('开始登录失败：' + await rsp.text());
		}
		let options = await rsp.json();
		let challenge = options.publicKey.challenge;
		options.publicKey = PublicKeyCredential.parseRequestOptionsFromJSON(options.publicKey);
		return { options, challenge };
	}
	async finishLogin(options, challenge) {
		let path = `/admin/login/webauthn/login:finish?challenge=${challenge}`;
		let rsp = await fetch(path, {
			method: 'POST',
			body: JSON.stringify(options),
		});
		if (!rsp.ok) {
			throw new Error('结束登录失败：' + await rsp.text());
		}
	}
}

class WebAuthn {
	constructor() {
		this.api = new WebAuthnAPI();
	}

	async register() {
		let options = await this.api.beginRegistration();
		console.log('开始注册返回：', options);
		options.publicKey.pubKeyCredParams = [
			{
				alg: -8,    // ed25519，我的最爱！❤️
				type: 'public-key',
			},
		];
		console.log('创建参数：', options);
		let credential = await navigator.credentials.create(options);
		if (!credential) { return null; }
		console.log('本地创建凭证：', credential);
		await this.api.finishRegistration(credential.toJSON());
		console.log('注册成功。');
	}

	async login() {
		let { options, challenge } = await this.api.beginLogin();
		console.log('开始登录返回：', options);
		if(options?.publicKey?.rpId != location.host) {
			throw new Error('站点配置的主页域名与当前访问域名不匹配。\n\n请使用用户名和密码登录后，修改站点主页地址后再尝试操作。');
		}
		let credential = await navigator.credentials.get(options);
		console.log('本地获取凭证：', credential);
		await this.api.finishLogin(credential.toJSON(), challenge);
		console.log('登录成功。');
	}
}

class PostManagementAPI
{
	constructor() { }

	// 更新/“编辑”文章。
	// 返回更新后的。
	async updatePost(p, extra) {
		let path = `/v3/posts/${p.id}`;
		let obj = {
			post: {
				date: p.date,
				date_timezone: p.date_timezone,
				modified: p.modified,
				modified_timezone: p.modified_timezone,
				type: p.type ?? 'tweet',
				status: p.status ?? 'public',
				source: p.source,
				metas: p.metas,
				source_type: p.source_type,
				top: p.top,
				category: p.category ?? 0,
			},
			update_mask: 'source,sourceType,date,dateTimezone,type,status,modifiedTimezone,metas',
			get_post_options: {
				with_user_perms: true,
			},
		};

		if(obj.post.status == 'partial') {
			obj.update_user_perms = true;
			obj.user_perms = extra.users;
			obj.send_user_notify = extra.send_user_notify;
		}

		obj.update_top = true;
		obj.update_category = true;

		let rsp = await fetch(path, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(obj),
		});
		return await decodeResponse(rsp);
	}

	// 文章预览
	/**
	 * @typedef {Object} PreviewResponse
	 * @property {string} html
	 * @property {string} diff
	 * @property {string[]} paths
	 * @property {string} title
	 * @property {number} updated_at
	 * 
	 * @param {number} id 
	 * @param {string} type 
	 * @param {string} source 
	 * @param {boolean} autoSave 
	 * @param {number} modifiedAt 
	 * @returns {Promise<PreviewResponse>}
	 */
	static async previewPost(id, type, source, autoSave, modifiedAt) {
		let path = `/v3/posts:preview`;
		let rsp = await fetch(path, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				id: id,
				type: type,
				source: source,
				save: autoSave ?? false,
				modified_at: modifiedAt ?? 0,
			})
		});
		return await decodeResponse(rsp);
	}

	// 任务列表/待办事项列表更新
	// id       文章编号
	// modified 文章修改时间
	// checks   待标记为“完成”的任务列表
	// unchecks 待标记为“未完成”的任务列表
	// NOTE：同时支持评论和文章——加速合并进程。
	async checkTaskListItems(id, isPost, modified, checks, unchecks) {
		let path = `/v3/${isPost?'posts':'comments'}/${id}/tasks:check`;
		let body = {
			modification_time: modified,
			checks: checks,
			unchecks: unchecks,
		};
		let rsp = await fetch(path, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify(body),
		});
		if (!rsp.ok) {
			let exception = await rsp.text();
			try { exception = JSON.parse(exception); }
			catch {}
			throw exception;
		}
		return await rsp.json();
	}

	async reorderTopPosts(ids) {
		let path = `/v3/posts:top`;
		let rsp = await fetch(path, {
			method: 'PATCH',
			body: JSON.stringify({ ids }),
		});
		if (!rsp.ok) {
			throw new Error('保存置顶文章失败：' + await rsp.text());
		}
	}

	static async createCategory(obj) {
		let path = `/v3/categories`;
		let rsp = await fetch(path, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify(obj),
		});
		if (!rsp.ok) {
			throw new Error('创建分类失败：' + await rsp.text());
		}
		return await rsp.json();
	}

	static async updateCategory(id, obj) {
		let path = `/v3/categories/${id}`;
		let rsp = await fetch(path, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify(obj),
		});
		if (!rsp.ok) {
			throw new Error('更新分类失败：' + await rsp.text());
		}
		return await rsp.json();
	}

	static async updateUserSettings(obj, masks) {
		masks.settings = obj;
		let path = '/v3/settings';
		let rsp = await fetch(path, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify(masks),
		});
		if (!rsp.ok) {
			throw new Error('更新设置失败：' + await rsp.text());
		}
		return await rsp.json();
	}
}
