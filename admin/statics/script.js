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
		let decoded = await this.base64decode(
			options.publicKey.challenge,
			options.publicKey.user.id,
		);
		options.publicKey.challenge = decoded[0];
		options.publicKey.user.id = decoded[1];
		return options;
	}
	async finishRegistration(options) {
		let encoded = await this.base64encode(
			options.rawId,
			options.response.clientDataJSON,
			options.response.attestationObject,
		);
		options.rawId = encoded[0];
		options.response.clientDataJSON = encoded[1];
		options.response.attestationObject = encoded[2];

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
		let decoded = await this.base64decode(
			options.publicKey.challenge,
		);
		options.publicKey.challenge = decoded[0];
		return options;
	}
	// challenge: 只是用来查找 session 的，肯定不会用作服务端真实 challenge。想啥呢？
	async finishLogin(options, challenge) {
		let encoded = await this.base64encode(
			options.rawId,
			options.response.authenticatorData,
			options.response.clientDataJSON,
			options.response.signature,
			options.response.userHandle,
			challenge,
		);
		options.rawId = encoded[0];
		options.response.authenticatorData = encoded[1];
		options.response.clientDataJSON = encoded[2];
		options.response.signature = encoded[3];
		options.response.userHandle = encoded[4];
		let path = `/admin/login/webauthn/login:finish?challenge=${encoded[5]}`;
		let rsp = await fetch(path, {
			method: 'POST',
			body: JSON.stringify(options),
		});
		if (!rsp.ok) {
			throw new Error('结束登录失败：' + await rsp.text());
		}
	}

	async base64encode() {
		let path = '/admin/login/webauthn/base64:encode';
		let body = Array.from(arguments).map(a => Array.from(new Uint8Array(a)));
		let rsp = await fetch(path, { method: 'POST', body: JSON.stringify(body)});
		if (!rsp.ok) { throw new Error(await rsp.text()); }
		return await rsp.json();
	}
	async base64decode() {
		let path = '/admin/login/webauthn/base64:decode';
		let rsp = await fetch(path, { method: 'POST', body: JSON.stringify(Array.from(arguments))});
		if (!rsp.ok) { throw new Error(await rsp.text()); }
		return (await rsp.json()).map(a => new Uint8Array(a));
	}
}

class WebAuthn {
	constructor() {
		this.api = new WebAuthnAPI();
	}

	async register() {
		let { publicKey } = await this.api.beginRegistration();
		console.log('开始注册返回：', publicKey);

		let { rp, challenge, user } = publicKey;
		
		let createOptions = {
			challenge: challenge,
			rp: {
				name: rp.name,
				id: rp.id,
			},
			pubKeyCredParams: [
				{
					alg: -8,    // ed25519，我的最爱！❤️
					type: 'public-key',
				},
			],
			user: {
				id: user.id,
				name: user.name,
				displayName: user.displayName,
			},
		};
		console.log('创建参数：', createOptions);

		let credential = null;
		try {
			credential = await navigator.credentials.create({ publicKey: createOptions });
		} catch (e) {
			console.log(e);
			throw e;
		}
		if (!credential) { return null; }
		console.log('本地创建凭证：', credential);

		let finishOptions = {
			id:     credential.id,
			type:   credential.type,
			rawId:  credential.rawId,
			response: {
				clientDataJSON:     credential.response.clientDataJSON,
				attestationObject:  credential.response.attestationObject,
			},
		};
		console.log('结束注册参数：', finishOptions);

		await this.api.finishRegistration(finishOptions);
		console.log('注册成功。');
	}

	async login() {
		let { publicKey } = await this.api.beginLogin();
		console.log('开始登录返回：', publicKey);
		let options = {
			challenge: publicKey.challenge,
		};
		let credential = await navigator.credentials.get({ publicKey: options});
		console.log('本地获取凭证：', credential);

		let finishLoginOptions = {
			id:     credential.id,
			type:   credential.type,
			rawId:  credential.rawId,
			response: {
				clientDataJSON:     credential.response.clientDataJSON,
				authenticatorData:  credential.response.authenticatorData,
				signature:          credential.response.signature,
				userHandle:         credential.response.userHandle,
			},
		};
		console.log('结束登录参数：', finishLoginOptions);
		await this.api.finishLogin(finishLoginOptions, publicKey.challenge);
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
