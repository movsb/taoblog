// TODO 在 _footer.html 有引用到，用于在页面上提供登录按钮。
class WebAuthnAPI {
	constructor(){}

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
				modified: p.modified,
				modified_timezone: p.modified_timezone,
				type: p.type ?? 'tweet',
				status: p.status ?? 'public',
				source: p.source,
				metas: p.metas,
				source_type: 'markdown',
				top: p.top,
			},
			update_mask: 'source,sourceType,date,type,status,modifiedTimezone,metas',
			get_post_options: {
				with_user_perms: true,
			},
		};

		if(obj.post.status == 'partial') {
			obj.update_user_perms = true;
			obj.user_perms = extra.users;
		}

		obj.update_top = true;

		let rsp = await fetch(path, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(obj),
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
			let exception = await rsp.text();
			try { 
				exception = JSON.parse(exception);
				exception = exception.message ?? exception;
			}
			catch {}
			throw exception;
		}
		return await rsp.json();
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

	async getTopPosts() {
		let path = `/v3/posts:top`;
		let rsp = await fetch(path);
		if (!rsp.ok) {
			throw new Error('获取置顶文章失败：' + await rsp.text());
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
}
