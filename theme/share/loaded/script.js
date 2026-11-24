// 按一级子元素的完整边界折叠长碎念。
document.addEventListener('DOMContentLoaded', () => {
	document.querySelectorAll('.tweets article').forEach(article => {
		const content = article.querySelector('.content');
		const toggle = article.querySelector('.tweet-toggle');
		if (!content || !toggle) { return; }
		let expanded = false;
		const update = () => {
			const children = Array.from(content.children);
			children.forEach(child => child.classList.remove('tweet-overflow'));
			const top = content.getBoundingClientRect().top;
			const limit = document.documentElement.clientHeight * .7;
			// 用实际位置累计高度，包含段落之间折叠后的 margin。
			const cutoff = children.findIndex(child => {
				if (!child.getClientRects().length) { return false; }
				return child.getBoundingClientRect().bottom - top > limit;
			});
			toggle.hidden = cutoff < 0;
			if (!expanded && cutoff >= 0) {
				children.slice(cutoff).forEach(child => child.classList.add('tweet-overflow'));
			}
			toggle.setAttribute('aria-expanded', String(expanded));
			toggle.textContent = expanded ? '收起' : '展开剩余内容';
		};
		toggle.addEventListener('click', event => {
			event.preventDefault();
			expanded = !expanded;
			update();
			if (!expanded && article.getBoundingClientRect().top < 0) {
				article.scrollIntoView({ block: 'start' });
			}
		});
		update();
		new ResizeObserver(update).observe(content);
		content.addEventListener('load', update, true);
		window.addEventListener('resize', update);
		document.fonts.ready.then(update);
	});
}, {once: true});

// 自动更新时间相对时间。
(function() {

function all() {
	let times = document.querySelectorAll('time[data-unix]');
	let stamps = Array.from(times).map(t => ({
		unix: parseInt(t.dataset.unix),
		timezone: t.dataset.timezone,
	}));
	let latest = 0;
	stamps.forEach(t => { if (t.unix > latest) latest = t.unix; });
	return { times, stamps, latest };
}

async function format(stamps) {
	let path = '/v3/utils/time/format';
	let formatted = undefined;
	const timezone = TimeWithZone.getTimezone();
	try {
		let rsp = await fetch(path, {
			method: 'POST',
			body: JSON.stringify({
				times: stamps,
				device: timezone,
			}),
		});
		if (!rsp.ok) {
			console.log(rsp.statusText);
			return;
		}
		rsp = await rsp.json();
		formatted = rsp.formatted;
	} catch (e) { console.log(e); return }
	if (!formatted) { return; }
	return formatted;
}

let update = async function() {
	let { times, stamps, latest } = all();
	if (!times.length) { return; }
	let formatted = await format(stamps);
	if (!formatted) { return; }
	times.forEach((t, i) => {
		const f = formatted[i];
		t.innerText = f.friendly;
		let title = `服务器时间：${f.server}`;
		if (f.device && f.device != f.server) {
			title = `${title}\n浏览器时间：${f.device}`;
		}
		if (f.original && f.original != f.server) {
			title = `${title}\n评论者时间：${f.original}`;
		}
		t.title = title;
	});
	let current =  Math.floor(new Date().getTime()/1000);
	let diff = current - latest;
	setTimeout(update, diff<60 ? 10000 : 60000)
}

setTimeout(update, 3000);

})();

// 不知道为什么 Firefox On iOS 会会长数字解析成电话号码，但是 Safari 并不会。
// 按照网上所有人说的添加 `<meta name="format-detection" content="telephone=no">` 也无济于事。
// 草草草🌿️🌿️🌿️！！！
//
// 原始HTML是没有这些a的，不知道浏览器是什么时候悄悄增加的。
// 出现过页面加载完成但是没有被替换的情况，所以延迟1秒再替换。
document.addEventListener('DOMContentLoaded', () => {
	setTimeout(()=>{
		document.querySelectorAll('a[x-apple-data-detectors]').forEach(a => {
			a.replaceWith(a.textContent);
		});
	}, 1000);
}, {once: true});

// 草稿管理时嵌入页面，总是以宽屏显示。
document.addEventListener('DOMContentLoaded', () => {
	if(window.parent != window && window.parent.location.pathname.endsWith('/drafts')) {
		document.body.classList.add('wide');
		document.body.classList.add('embed-drafts');
	}
}, {once: true});
