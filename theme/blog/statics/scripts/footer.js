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

(function() {
	if (TaoBlog && TaoBlog.vim) {
		TaoBlog.vim.bind('a', async ()=>{
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
		});
	}
})();

// 不知道为什么 Firefox On iOS 会会长数字解析成电话号码，但是 Safari 并不会。
// 按照网上所有人说的添加 `<meta name="format-detection" content="telephone=no">` 也无济于事。
// 草草草🌿️🌿️🌿️！！！
setTimeout(()=>{
	document.querySelectorAll('a[x-apple-data-detectors]').forEach(a => {
		a.replaceWith(a.textContent);
	});
}, 1000);
