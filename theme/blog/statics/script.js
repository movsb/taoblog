function __TaoBlog()
{
    return {
        fn: {},
		posts: {},
    };
}

var TaoBlog = window.TaoBlog = new __TaoBlog();

TaoBlog.fn.fadeIn = function(elem, callback) {
	return TaoBlog.fn._fadeIn(elem, callback, 'fade-in');
};
TaoBlog.fn.fadeIn95 = function(elem, callback) {
	return TaoBlog.fn._fadeIn(elem, callback, 'fade-in-95');
};
TaoBlog.fn._fadeIn = function(elem, callback, name) {
	elem.classList.remove('fade-out');
	elem.classList.remove('fade-out-95');
	elem.style.display = 'block';
	if (typeof callback == 'function') {
		elem.addEventListener('animationend', function(event) {
			// console.log('fade-in animationend');
			callback();
		}, { once: true});
	}
	elem.classList.add(name);
};
TaoBlog.fn.fadeOut = function(elem, callback, name) {
	return TaoBlog.fn._fadeOut(elem, callback, 'fade-out');
}
TaoBlog.fn.fadeOut95 = function(elem, callback, name) {
	return TaoBlog.fn._fadeOut(elem, callback, 'fade-out-95');
}
TaoBlog.fn._fadeOut = function(elem, callback, name) {
	elem.classList.remove('fade-in');
	elem.classList.remove('fade-in-95');
	elem.addEventListener('animationend', function(event) {
		// console.log('fade-out animationend');
		elem.style.display = 'none';
		if (typeof callback == 'function') {
			callback();
		}
	}, { once: true});
	elem.classList.add(name);
};

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
	});
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
	constructor() {
		super();
		this.addEventListener('click', this.navigate);
	}
	connectedCallback() {
		this.classList.add('like-a');
		this.style.color = `inherit`;
		this.style.cursor = 'pointer';
	}
	navigate(event) {
		const longitude = this.getAttribute('longitude');
		const latitude = this.getAttribute('latitude');
		this.openMap(longitude, latitude);
	}
	// 怎样在浏览器里面调用打开系统地图的链接。
	// 增加了通过时区判断是否在中国，决定是否回退到谷歌打开。
	openMap(lon, lat) {
		const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent);
		const isAndroid = /Android/.test(navigator.userAgent);

		if (isIOS) {
			// Apple Maps
			window.location.href = `maps://?q=${lat},${lon}`;
		} else if (isAndroid) {
			// Android Maps
			window.location.href = `geo:${lat},${lon}`;
		} else {
			const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
			const maybeChina = /Asia\/(Shanghai|Beijing|Chongqing)/.test(timezone);
			const encodedTitle = encodeURIComponent(this.innerText);
			const url = maybeChina
				? `https://map.baidu.com/?latlng=${lat},${lon}&title=${encodedTitle}`
				: `https://www.google.com/maps?q=${lat},${lon}`
				;
			window.open(url, '_blank');
		}
	}
}
customElements.define('geo-link', GeoLink);
