function __TaoBlog()
{
    return {
        fn: {},
		posts: {},
    };
}

var TaoBlog = window.TaoBlog = new __TaoBlog();

TaoBlog.fn.parseCookies = function() {
	return Object.fromEntries(
	  document.cookie
		.split('; ')
		.map(cookie => cookie.split('=').map(decodeURIComponent))
	);
}
TaoBlog.fn.getUserID = function() {
	return +(TaoBlog.fn.parseCookies()['taoblog.user_id'] || 0);
};
TaoBlog.fn.getNickname = function() {
	return TaoBlog.fn.parseCookies()['taoblog.nickname'] || '';
};

if (TaoBlog.fn.getUserID() > 0) {
	document.addEventListener('DOMContentLoaded', ()=>{
		document.body.classList.add('signed-in');
	}, {once: true});
}

class TimeWithZone {
	constructor(timestamp, zone) {
		const now = new Date();
		if (typeof timestamp != 'number') {
			timestamp = now.getTime() / 1000;
			zone = TimeWithZone.getTimezone();
		} else if (typeof zone != 'string' || zone == '') {
			zone = TimeWithZone.getTimezone();
		}
		this._timestamp = timestamp;
		this._zone = zone;
	}
	
	get time() { return this._timestamp; }
	get zone() { return this._zone;      }
	
	toJSON() {
		return new Date(this._timestamp*1000).toJSON();
	}

	static getTimezone() {
		try {
			return Intl.DateTimeFormat().resolvedOptions().timeZone;
		} catch {
			return '';
		}
	}
}

class GeoLink extends HTMLElement {
	connectedCallback() {
		const a = document.createElement('a');
		this.init(a);
		this.appendChild(a);
	}
	/**
	 * 
	 * @param {HTMLAnchorElement} a 
	 */
	init(a) {
		a.innerText = this.getAttribute('name');

		const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
		const maybeChina = /Asia\/(Shanghai|Beijing|Chongqing)/.test(timezone);

		const [lat,lon] = JSON.parse(this.getAttribute(maybeChina ? 'gcj02' : 'wgs84'));

		// 怎样在浏览器里面调用打开系统地图的链接。
		// 增加了通过时区判断是否在中国，决定是否回退到谷歌打开。
		// 对应中国，是 GCJ02 坐标。否则为 WGS04 坐标。
		// [对国内地图坐标系统的一些观察 - 陪她去流浪](https://blog.twofei.com/1967/)
		const isIOS = /iPad|iPhone|iPod|Macintosh/.test(navigator.userAgent);
		const isAndroid = /Android/.test(navigator.userAgent);

		if (isIOS) {
			// 在 ios/firefox 上不工作，换成 http 链接。
			// a.href = `maps://?q=${lat},${lon}`;
			// [Adopting unified Maps URLs | Apple Developer Documentation](https://developer.apple.com/documentation/mapkit/unified-map-urls)
			// 还是好像还是不能工作，垃圾 firefox。
			// a.href = `https://maps.apple.com?query=${lat},${lon}`;
			//
			// 新版地址有变化。
			// https://developer.apple.com/documentation/mapkit/unified-map-urls#Show-the-place-card-of-a-POI
			// 参考这个地址：
			// https://maps.apple.com/place?coordinate=40.864791,-73.931723&name=My%20Favorite%20Place.
			// 问题是：官方文档并没有给出 name 的介绍。
			const url = new URL('https://maps.apple.com/place');
			const query = new URLSearchParams();
			query.set('coordinate', `${lat},${lon}`);
			query.set('name', this.getAttribute('name'));
			url.search = '?' + query.toString();
			a.href = url.toString();
		} else if (isAndroid) {
			a.href = `geo:${lat},${lon}`;
		} else {
			const url = china
				? `https://uri.amap.com/marker?position=${lon},${lat}`
				// t=k：卫星图，ChatGPT 说的，谷歌官方没找到文档。
				: `https://www.google.com/maps?q=${lat},${lon}&t=k`
				;
			a.href = url;
		}
	}
}
customElements.define('geo-link', GeoLink);
