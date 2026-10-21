/**
 * @typedef {{}} TaoBlog
 */

__TaoBlog.prototype.parseCookies = function() {
	const entries = [];
	for (const part of document.cookie.split(';')) {
		const cookie = part.trim();
		const separator = cookie.indexOf('=');
		if (separator <= 0) { continue; }
		try {
			entries.push([
				decodeURIComponent(cookie.slice(0, separator)),
				decodeURIComponent(cookie.slice(separator + 1)),
			]);
		} catch {
			// 单个 Cookie 编码异常时，仍继续解析其他 Cookie。
		}
	}
	return Object.fromEntries(entries);
}
__TaoBlog.prototype.getUserID = function() {
	const id = Number(TaoBlog.parseCookies()['taoblog.user_id'] || 0);
	return Number.isSafeInteger(id) && id > 0 ? id : 0;
};
__TaoBlog.prototype.getNickname = function() {
	return TaoBlog.parseCookies()['taoblog.nickname'] || '';
};

if (TaoBlog.getUserID() > 0) {
	document.addEventListener('DOMContentLoaded', ()=>{
		document.body.classList.add('signed-in');
	}, {once: true});
}

// 页面重新激活后检查会话，确认失效时才提示登录。
let checkingSession = false;
let sessionCheckTimer = null;
function onPageActive() {
	clearTimeout(sessionCheckTimer);
	sessionCheckTimer = null;
	const canCheck = () => document.visibilityState === 'visible'
		&& location.pathname.startsWith('/admin/')
		&& !location.pathname.startsWith('/admin/login')
		&& !location.pathname.startsWith('/admin/logout');
	if (!canCheck() || checkingSession) { return; }

	// 触发得太快容易导致登录界面被频繁弹出，等页面稳定后再检查。
	sessionCheckTimer = setTimeout(async ()=>{
		sessionCheckTimer = null;
		if (!canCheck() || checkingSession) { return; }
		checkingSession = true;

		try {
			// 当前登录的用户ID，如果没有登录，则为0。
			const userID = TaoBlog.getUserID();
			const response = await fetch('/admin/session', { cache: 'no-store' });
			let badSession = response.status === 401;
			if (!badSession) {
				if (!response.ok) {
					throw new Error(`会话检查失败：HTTP ${response.status}`);
				}
				const session = await response.json();
				const serverUserID = session?.user_id;
				if (!Number.isSafeInteger(serverUserID) || serverUserID <= 0) {
					throw new Error('会话检查返回了无效的用户编号。');
				}
				badSession = userID === 0 || serverUserID !== userID;
			}

			// 请求期间可能切换了页面、标签页或登录用户，不使用过时的结果提示登录。
			if (!canCheck() || TaoBlog.getUserID() !== userID) { return; }
			if (badSession) {
				// 登录状态无效，提示用户重新登录。
				try {
					let wa = new WebAuthn();
					// BUG: cookie过期后，连同user_id也会一起删除，还是会提示所有用户列表。
					await wa.login({user_id: userID});
				} catch(e) {
					if (e instanceof DOMException && ["NotAllowedError", "AbortError"].includes(e.name)) {
						console.log('已取消操作。');
						return;
					}
					alert(e);
				}
				return;
			}
		} catch (error) {
			// 网络、服务端或响应格式异常时，暂时无法验证登录状态。
			console.error('无法验证登录状态:', error);
		} finally {
			checkingSession = false;
		}
	}, 2000);
}
document.addEventListener('visibilitychange', onPageActive);
// window.addEventListener('focus', onPageActive);
window.addEventListener('pageshow', onPageActive);
