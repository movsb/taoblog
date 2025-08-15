class ImageViewDesktop {
	constructor(img) {
		let html = `<div class="img-view" id="img-view" style="display: none"></div>\n`;
		
		let old = document.getElementById('img-view');
		if (old) { old.remove(); }

		let div = document.createElement('div');
		div.innerHTML = html;
		document.body.appendChild(div.firstElementChild);
		
		this.root = document.getElementById('img-view');

		this.show(true);
	
		this._scale = 1;

		this.viewImage(img);
		this.initBindings();

		this._boundKeyHandler = this._keyHandler.bind(this);
		document.body.addEventListener('keydown', this._boundKeyHandler);
	}
	
	show(yes) {
		if (yes) {
			if (TaoBlog && TaoBlog.fn && TaoBlog.fn.fadeIn) {
				TaoBlog.fn.fadeIn(this.root);
			} else{
				this.root.style.display = 'block';
			}
		} else {
			if (TaoBlog && TaoBlog.fn && TaoBlog.fn.fadeOut) {
				TaoBlog.fn.fadeOut(this.root);
			} else{
				this.root.style.display = 'none';
			}
			this._hideImage();
		}
	}

	viewImage(img) {
		// 对于 xhtml 来说保留原始大小写
		// https://developer.mozilla.org/en-US/docs/Web/API/Element/tagName
		if (img.tagName == 'svg') {
			let svg = img.cloneNode(true);
			if (img.classList.contains('transparent')) {
				this.root.classList.add('transparent');
			}
			this.obj = svg;
			this.ref = img;
			this.root.appendChild(svg);
			setTimeout(()=>this._onImgLoad(), 0);
		} else {
			if (img.classList.contains('transparent')) {
				this.root.classList.add('transparent');
			}
			this.obj = document.createElement('img');
			this.obj.addEventListener('load', this._onImgLoad.bind(this));
			this.root.appendChild(this.obj);
			this.obj.src = img.src;
			this.ref = img;
			this.initMetadata(img.dataset.metadata);
		}
	}

	_hideImage() {
		let body = document.body;
		body.style.overflow = 'auto';
		body.removeEventListener('keydown', this._boundKeyHandler);
	}

	_keyHandler(e) {
		if (e.keyCode == 27) {
			this.show(false);
			e.preventDefault();
			e.stopPropagation();
		}
	}
	_onImgLoad() {
		let rawWidth, rawHeight;
		
		const obj = this.obj;

		if (obj.tagName == 'IMG') {
			rawWidth = obj.naturalWidth || parseInt(this.ref.style.width) || parseInt(this.ref.getAttribute('width')) || 300;
			rawHeight = obj.naturalHeight || parseInt(this.ref.style.height) || parseInt(this.ref.getAttribute('height')) || 300;
		} else if (obj.tagName == 'svg') {
			this.obj.style.opacity = .01;
			this.obj.style.display = 'block';
			let {width, height } = obj.getBBox();
			rawWidth = width || 300;
			rawHeight = height || 300;
		} else {
			console.error('未处理的类型。');
		}

		let initScale = 1;
		let initWidth = rawWidth * initScale;
		let initHeight = rawHeight * initScale;

		{
			let scaleWidth  = this.root.clientWidth  / initWidth;
			let scaleHeight = this.root.clientHeight / initHeight;

			// if smaller than container
			if (scaleWidth >= 1 && scaleHeight >= 1) {
				initScale *= 1;
			}

			// if larger than container, scale to fit
			else {
				initScale *= Math.min(scaleWidth, scaleHeight);
			}

			initWidth = rawWidth * initScale;
			initHeight = rawHeight * initScale;
		}

		let style = this.obj.style;
		style.left      = `${(this.root.clientWidth  - initWidth) /2}px`;
		style.top       = `${(this.root.clientHeight - initHeight)/2}px`;
		style.width     = `${initWidth}px`;
		style.height    = `${initHeight}px`;

		if (TaoBlog && TaoBlog.fn && TaoBlog.fn.fadeIn) {
			TaoBlog.fn.fadeIn(this.obj);
		} else {
			style.display   = 'block';
		}

		let body = document.body;
		body.style.overflow = 'hidden';
	}
	initBindings() {
		// 单击图片时，缩放到原始大小。
		// 或者，关闭预览。
		this.obj.addEventListener('click', (e)=> {
			if(this._scale != 1) {
				this.scale = 1;
				this._scale = 1;
			} else {
				this.show(false);
			}
			e.preventDefault();
			e.stopPropagation();
			return false;
		});

		this.root.addEventListener('click', (e)=> {
			// 点击空白处关闭预览。
			this.show(false);
		});

		this.obj.addEventListener('mousemove', (e)=>{
			if (e.buttons != 0) { return; }
			// console.log('moving:', e, e.offsetX, e.offsetY);
			// 以一种奇怪的方式实现了缩放鼠标位置并根据鼠标位置移动图片位置。
			this.obj.style.transformOrigin = `${e.offsetX}px ${e.offsetY}px`;
		});
		
		// 禁止拖动图片。
		this.obj.addEventListener('mousedown', (e)=> {
			e.preventDefault();
			e.stopPropagation();
		});

		this.root.addEventListener('wheel', (e)=>{
			// https://developer.mozilla.org/en-US/docs/Web/API/Element/wheel_event
			// 根据 delta 值（正、负）求得允许范围内的新的缩放值。
			this._scale += e.deltaY * -0.01;
			this._scale = Math.min(Math.max(1, this._scale), 10);
			this.scale = this._scale;

			e.preventDefault();
			return false;
		});
	}

	set scale(value) {
		this.obj.style.transform = `scale(${value})`;
	}

	initMetadata(metadata) {
		let md;
		try {
			md = JSON.parse(metadata);
			if (md.length <= 0 || md.length&1 > 0) {
				return;
			}
		} catch {
			return;
		}
		let title = '';
		for(let i=0; i<md.length; i+=2) {
			title += `${md[i+0]}：${md[i+1]}\n`;
		}
		this.obj.title = title;
	}
}

class ImageViewMobile {
	constructor(images) {
		this.numberOfImages = images.length;
		const div = document.createElement('div');
		div.innerHTML = `<div id="img-view-mobile"></div>`;
		this.root = div.firstElementChild;
		images.forEach((img, index)  => {
			const clone = img.cloneNode(true);
			this.root.appendChild(clone);
			img.addEventListener('click', (e) => {
				this.view(index);
				e.preventDefault();
				e.stopPropagation();
			});
			clone.addEventListener('click', (e) => {
				this.root.style.display = 'none';
				e.preventDefault();
				e.stopPropagation();
			});
		});
		document.body.appendChild(this.root);
	}
	view(index) {
		console.log('viewing image:', index);
		this.root.style.opacity = 0;
		this.root.style.display = 'flex';
		const width = this.root.clientWidth;
		this.root.scrollLeft = width * index;
		this.root.style.opacity = 1;
	}
}

class ImageView {
	constructor() {
		if (this.isMobileDevice()) {
			this.initMobile();
		} else {
			this.initDesktop();
		}
	}

	isMobileDevice() {
		return 'ontouchstart' in window || /iPhone|iPad|Android|XiaoMi/.test(navigator.userAgent);
	}

	initDesktop() {
		let images = document.querySelectorAll('.entry img:not(.no-zoom):not(.static)');
		images.forEach(img => img.addEventListener('click', e => {
			if (!img.complete) {
				console.log('图片未完成加载：', img);
				return;
			}
			new ImageViewDesktop(img);
		}));
	}

	initMobile() {
		let images = document.querySelectorAll('.entry img:not(.no-zoom):not(.static)');
		new ImageViewMobile(images)
	}
}

document.addEventListener('DOMContentLoaded', () => {
	(TaoBlog||window).imgView = new ImageView();
});
