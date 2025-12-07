/**
 * @import { readFileAsDataURL } from '../dynamic/utils.js';
 */

const userID = TaoBlog.getUserID();

async function register() {
	let wa = new WebAuthn();
	try {
		await wa.register();
		alert('新的通行密钥注册成功。');
	} catch(e) {
		if (e instanceof DOMException && ["NotAllowedError", "AbortError"].includes(e.name)) {
			console.log('已取消操作。');
			return;
		}
		alert(e);
	}
}
async function switchUser() {
	let wa = new WebAuthn();
	try {
		await wa.login();
		location.reload();
	} catch(e) {
		if (e instanceof DOMException && ["NotAllowedError", "AbortError"].includes(e.name)) {
			console.log('已取消操作。');
			return;
		}
		alert(e);
	}
}

document.getElementById('change-nickname').addEventListener('click', async () => {
	const elem = document.getElementById('nickname');
	const current = elem.textContent;
	const name = prompt('请输入新的昵称：', current);
	if(name === null) { return; }
	if(name != current) {
		const rsp = await fetch(`/v3/users/${userID}`, {
			method: 'PATCH',
			body: JSON.stringify({
				user: {
					nickname: name,
				},
				update_nickname: true,
			}),
		})
		if(!rsp.ok) {
			alert('修改失败。');
			return;
		}
		elem.textContent = name;
	}
});

document.getElementById('change-email').addEventListener('click', async () => {
	const elem = document.getElementById('email');
	const current = elem.textContent;
	const name = prompt('请输入新的邮箱：', current.includes('@') ? current : '');
	if(name === null) { return; }
	if(name != current) {
		const rsp = await fetch(`/v3/users/${userID}`, {
			method: 'PATCH',
			body: JSON.stringify({
				user: {
					email: name,
				},
				update_email: true,
			}),
		})
		if(!rsp.ok) {
			alert('修改失败。');
			return;
		}
		elem.textContent = name;
	}
});

/** @type {HTMLImageElement} */
const avatarImg = document.getElementById('avatar');
avatarImg.addEventListener('click', () => {
	/** @type {HTMLInputElement} */
	const file = document.getElementById('avatar-file');
	file.onchange = async () => {
		console.log(file.files);
		if(file.files.length != 1) {
			return;
		}
		const url = await readFileAsDataURL(file.files[0]);
		console.log(url);

		/** @type {HTMLParagraphElement} */
		const status = document.querySelector('#avatar-wrapper .uploading');
		status.style.display = 'block';

		try {
			const rsp = await fetch(`/v3/users/${userID}`, {
				method: `PATCH`,
				body: JSON.stringify({
					user: {
						avatar: url,
					},
					update_avatar: true,
				}),
			});
			if(!rsp.ok) {
				throw await rsp.text();
			}
			const u = new URL(avatarImg.src);
			const s = new URLSearchParams(u.search);
			s.set('_', new Date().getTime());
			u.search = '?' + s.toString();
			avatarImg.src = u.toString();
		} catch (e) {
			alert('更新失败：' + e);
		} finally {
			status.style.display = 'none';
		}
	};
	file.click();
});
