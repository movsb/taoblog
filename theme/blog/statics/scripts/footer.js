// TODO 目前的 Markdown 在处理 @2x 图片时无法处理 HTML 标签使用的图片 <img>，只能处理 ![]() 这种。
// 这里使用脚本临时处理一下，后续应该在 Markdown 里面统一处理。
(function(){
	let imgs = document.querySelectorAll('img');
	imgs.forEach(function(img) {
		if(img.src.indexOf('@2x.') >= 0 && img.style.width == '' && img.style.height == '') {
			img.addEventListener('load', function() {
				if (img.naturalWidth > 0 && img.naturalHeight > 0) {
					img.style.width = `${img.naturalWidth/2}px`;
					img.style.height = `${img.naturalHeight/2}px`;
				}
			});
		}
	});
})();

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
	if (diff < 60) { setTimeout(update, 10000); return; }
	setTimeout(update, 60000);
}

setTimeout(update, 3000);

TaoBlog.events.add('comment', 'post', () => { update(); });

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
