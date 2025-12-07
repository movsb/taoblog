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
