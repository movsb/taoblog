class ImageViewUI {
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

		this._viewImage(img);
		this._initBindings();

		if (!this._isMobileDevice()) {
			this._boundKeyHandler = this._keyHandler.bind(this);
			document.body.addEventListener('keydown', this._boundKeyHandler);
		}
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

	_viewImage(img) {
		// 对于 xhtml 来说保留原始大小写
		// https://developer.mozilla.org/en-US/docs/Web/API/Element/tagName
		if (img.tagName == 'svg') {
			let svg = img.cloneNode(true);
			if (img.classList.contains('transparent')) {
				this.root.classList.add('transparent');
			}
			this.obj = svg;
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
		}
	}

	_hideImage() {
		if (this._isMobileDevice()) {
			let wrapper = document.getElementById('wrapper');
			wrapper.style.display = 'block';
			window.scrollTo({left: 0, top: this._scrollY});
		} else {
			let body = document.body;
			body.style.maxHeight = 'none';
			body.style.overflow = 'auto';
			body.removeEventListener('keydown', this._boundKeyHandler);
		}
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
			rawWidth = obj.naturalWidth;
			rawHeight = obj.naturalHeight;
		} else {
			rawWidth = (parseInt(obj.style.width) || parseInt(obj.getAttribute('width'))) ?? 100;
			rawHeight = (parseInt(obj.style.height) || parseInt(obj.getAttribute('height'))) ?? 100;
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

		// 这个是 Zoom.js 的 Bugs。
		// 本来在桌面浏览器上对 body hidden 就够了，
		// 但是移动设备上会出现当页面有滚动时，图片预览缩放功能乱跑。
		if (this._isMobileDevice()) {
			this._scrollY = window.scrollY;
			let wrapper = document.getElementById('wrapper');
			wrapper.style.display = 'none';
		} else {
			let body = document.body;
			body.style.maxHeight = window.innerHeight + 'px';
			body.style.overflow = 'hidden';
		}
	}
	_isMobileDevice() {
		return 'ontouchstart' in window || /iPhone|iPad|Android|XiaoMi/.test(navigator.userAgent);
	}
	_initBindings() {
		if (this._isMobileDevice()) {
			let zoom = new Zoom(this.obj, {
				rotate: false,
				// minZoom: 0.25,
				// maxZoom: 5,
			});
			this.root.addEventListener('touchstart', (e)=> {
				if (e.touches.length === 1) {
					if (this.mayBeDoubleTap) {
						clearTimeout(this.mayBeDoubleTap);
						this.mayBeDoubleTap = null;
						this.show(false);
						e.preventDefault();
						e.stopPropagation();
					} else {
						this.mayBeDoubleTap = setTimeout(()=>this.mayBeDoubleTap=null,250);
					}
				}
			});
		} else {
			let moveHandler = this._onImgMouseMove.bind(this);
			
			// https://stackoverflow.com/a/52839734/3628322
			this.obj.addEventListener('mousedown', (e)=> {
				// console.log('down...');
				// 点击 svg 的其它地方不要拖，方便复制文本。
				if (this.obj.tagName == 'svg' && e.target.tagName != 'svg') {
					// this._onImgMouseDown(e);
					console.log('not clicking on svg root node.')
					this._offsetX = e.clientX;
					this._offsetY = e.clientY;
					return;
				}
				this._onImgMouseDown(e);
				window.addEventListener('mousemove', moveHandler, true);
				window.addEventListener('mouseup', (e) => {
					// console.log('up...');
					this._onImgMouseUp(e);
					window.removeEventListener('mousemove', moveHandler, true);
				}, { once: true });
			}, true);

			let imgHandlers = {
				'click':            this._onImgClick,
				'transitionend':    this._onTransitionEnd,
			};
			for (let key in imgHandlers) {
				this.obj.addEventListener(key, imgHandlers[key].bind(this), true);
			}
			
			let divHandlers = {
				'wheel':        this._onDivMouseWheel,
				'click':        this._onDivClick,
			};
			for (let key in divHandlers) {
				this.root.addEventListener(key, divHandlers[key].bind(this));
			}
		}
	}
	_onImgMouseDown(e) {
		if (this.obj.getAttribute('data-busy') == '1') {
			e.preventDefault();
			return false;
		}

		// http://stackoverflow.com/a/2725963
		if (e.which == 1) { // left button
			this._offsetX = e.clientX;
			this._offsetY = e.clientY;

			this._coordX = parseInt(this.obj.style.left);
			this._coordY = parseInt(this.obj.style.top);

			this._dragging = true;
		}

		e.preventDefault();
		return false;
	}
	_onImgMouseMove(e) {
		console.log('moving');
		if (!this._dragging) {
			return false;
		}

		let left = this._coordX + e.clientX - this._offsetX + 'px';
		let top = this._coordY + e.clientY - this._offsetY + 'px';
		
		// console.log({left, top});
		this.obj.style.left = left;
		this.obj.style.top = top;

		e.preventDefault();
		return false;
	}
	_onImgMouseUp(e) {
		this._dragging = false;
		e.preventDefault();
		console.log('exit dragging');
		return false;
	}
	_onImgClick(e) {
		// console.log('img: click');
		if (this.obj.tagName == 'svg' && e.target.tagName != 'svg') {
			console.log('clicking on text nodes.');
			e.preventDefault();
			e.stopPropagation();
			return;
		}
		let smallMove = false;
		{
			let horz = Math.abs(e.clientX - this._offsetX);
			let vert = Math.abs(e.clientY - this._offsetY);

			if (horz <= 1 && vert <= 1) {
				smallMove = true;
			}
		}

		if (smallMove) {
			this.show(false);
			console.log('hide because of small move');
		}

		e.preventDefault();
		e.stopPropagation();
		return false;
	}
	_onTransitionEnd(e) {
		this.obj.style.transition = '';
		this.obj.setAttribute('data-busy', '');
	}
	_onDivMouseWheel(e) {
		if (this.obj.getAttribute('data-busy') == '1') {
			e.preventDefault();
			return false;
		}

		// https://developer.mozilla.org/en-US/docs/Web/API/Element/wheel_event
		// 根据 delta 值（正、负）求得允许范围内的新的缩放值。
		this._scale += e.deltaY * -0.01;
		this._scale = Math.min(Math.max(.25, this._scale), 10);

		this.obj.style.transform = `scale(${this._scale})`;

		e.preventDefault();
		return false;
	}
	_onDivClick(e) {
		this.show(false);
	}
}

class ImageView {
	constructor() {
		let images = document.querySelectorAll('.entry img');
		images.forEach(img => img.addEventListener('click', e => this.show(e.target)));
		let svgs = document.querySelectorAll('.entry svg');
		svgs.forEach(img => img.addEventListener('click', e => {
			// 仅点空白处才显示图片，否则可能是复制文本。
			if (e.target.tagName == 'svg') {
				this.show(e.currentTarget);
			}
		}));
	}
	
	show(img) {
		new ImageViewUI(img);
	}
}

TaoBlog.imgView = new ImageView();
