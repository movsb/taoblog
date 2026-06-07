/**
 * @typedef {{}} TaoBlog
 */

__TaoBlog.prototype.parseCookies = function() {
	return Object.fromEntries(
	  document.cookie
		.split('; ')
		.map(cookie => cookie.split('=').map(decodeURIComponent))
	);
}
__TaoBlog.prototype.getUserID = function() {
	return +(TaoBlog.parseCookies()['taoblog.user_id'] || 0);
};
__TaoBlog.prototype.getNickname = function() {
	return TaoBlog.parseCookies()['taoblog.nickname'] || '';
};

if (TaoBlog.getUserID() > 0) {
	document.addEventListener('DOMContentLoaded', ()=>{
		document.body.classList.add('signed-in');
	}, {once: true});
}

// Cookie 已绑定到具体的登录IP地址，可能会经常变。
// 所以在文档激活后，定期检查Cookie是否有效，若无效，则提示登录。
let checkingSession = false;
async function onPageActive() {
	if(document.visibilityState != 'visible') return;
	if(!location.pathname.startsWith('/admin/')) return;
	if(location.pathname.startsWith('/admin/login') || location.pathname.startsWith('/admin/logout')) return;

	// 触发得太快容易导致登录界面被频繁弹出，等页面稳定后再检查。
	setTimeout(async ()=>{
		if(checkingSession) return;
		checkingSession = true;

		// 当前登录的用户ID，如果没有登录，则为0。
		const userID = TaoBlog.getUserID();

		try {
			const response = await fetch('/admin/session');
			const badSession = response.status === 401;
			const serverUserID = response.ok ? (await response.json()).user_id : 0;

			if (badSession || userID == 0 || serverUserID !== userID) {
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
			// 网络错误，暂时无法验证登录状态，不做任何处理。
			console.error('无法验证登录状态:', error);
		} finally {
			checkingSession = false;
		}
	}, 2000);
}
document.addEventListener('visibilitychange', onPageActive);
// window.addEventListener('focus', onPageActive);
window.addEventListener('pageshow', onPageActive);
