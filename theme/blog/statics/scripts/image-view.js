class ImageViewUI {
	constructor(img) {
		let html = `
<div class="img-view" id="img-view">
	<img />
</div>
`;
		
		let old = document.getElementById('img-view');
		if (old) { old.remove(); }

		let div = document.createElement('div');
		div.innerHTML = html;
		document.body.appendChild(div.firstElementChild);
	
		this.root = document.getElementById('img-view');
		this.img = document.querySelector('#img-view img');

		this._scale = 1;

		this._initBindings();
		
		this._viewImage(img);

		this._boundKeyHandler = this._keyHandler.bind(this);
		document.body.addEventListener('keydown', this._boundKeyHandler);
	}

	_viewImage(img) {
		let src = img.src;
		if (img.classList.contains('transparent')) {
			this.img.classList.add('transparent');
		}
		this.img.src = src;	
	}
	_hideImage() {
		document.body.style.maxHeight = 'none';
		document.body.style.overflow = 'auto';
		document.body.removeEventListener('keydown', this._boundKeyHandler);
		this.root.remove();
	}

	_keyHandler(e) {
		if (e.keyCode == 27) {
			this._hideImage();
			e.preventDefault();
			e.stopPropagation();
		}
	}
	_onImgLoad() {
		let rawWidth = this.img.naturalWidth;
		let rawHeight = this.img.naturalHeight;

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

		let style = this.img.style;
		style.left      = `${(this.root.clientWidth  - initWidth) /2}px`;
		style.top       = `${(this.root.clientHeight - initHeight)/2}px`;
		style.width     = `${initWidth}px`;
		style.height    = `${initHeight}px`;
		style.display   = 'block';
		
		let body =document.body;
		body.style.maxHeight = window.innerHeight + 'px';
		body.style.overflow = 'hidden';
	}
	_initBindings() {
		this.img.addEventListener('load', this._onImgLoad.bind(this));

		if ('ontouchstart' in window || /iPhone|iPad|Android|XiaoMi/.test(navigator.userAgent)) {
			let zoom = new Zoom(this.img, {
				rotate: false,
				// minZoom: 0.25,
				// maxZoom: 5,
			});
			this.root.addEventListener('touchstart', (e)=> {
				if (e.touches.length === 1) {
					if (this.mayBeDoubleTap) {
						clearTimeout(this.mayBeDoubleTap);
						this.mayBeDoubleTap = null;
						this._hideImage();
						e.preventDefault();
						e.stopPropagation();
					} else {
						this.mayBeDoubleTap = setTimeout(()=>this.mayBeDoubleTap=null,250);
					}
				}
			});
		} else {
			let imgHandlers = {
				'mousedown':        this._onImgMouseDown,
				'mousemove':        this._onImgMouseMove,
				'mouseup':          this._onImgMouseUp,
				'click':            this._onImgClick,
				'transitionend':    this._onTransitionEnd,
			};
			for (let key in imgHandlers) {
				this.img.addEventListener(key, imgHandlers[key].bind(this));
			}
			
			let divHandlers = {
				'wheel':        this._onDivMouseWheel,
				'mousemove':    this._onDivMouseMove,
				'click':        this._onDivClick,
			};
			for (let key in divHandlers) {
				this.root.addEventListener(key, divHandlers[key].bind(this));
			}
		}
	}
	_onImgMouseDown(e) {
		if (this.img.getAttribute('data-busy') == '1') {
			e.preventDefault();
			return false;
		}

		let target = e.target;

		// http://stackoverflow.com/a/2725963
		if (e.which == 1) { // left button
			this._offsetX = e.clientX;
			this._offsetY = e.clientY;

			this._coordX = parseInt(target.style.left);
			this._coordY = parseInt(target.style.top);

			this._dragging = true;
		}

		e.preventDefault();
		return false;
	}
	_onImgMouseMove(e) {
		if (!this._dragging) {
			return false;
		}

		let left = this._coordX + e.clientX - this._offsetX + 'px';
		let top = this._coordY + e.clientY - this._offsetY + 'px';

		let target = e.target;
		target.style.left = left;
		target.style.top = top;

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
		console.log('img: click');
		let smallMove = false;
		{
			let horz = Math.abs(e.clientX - this._offsetX);
			let vert = Math.abs(e.clientY - this._offsetY);

			if (horz <= 1 && vert <= 1) {
				smallMove = true;
			}
		}

		if (smallMove) {
			this._hideImage();
			console.log('hide because of small move');
		}

		e.preventDefault();
		e.stopPropagation();
		return false;
	}
	_onTransitionEnd(e) {
		this.img.style.transition = '';
		this.img.setAttribute('data-busy', '');
	}
	_onDivMouseWheel(e) {
		if (this.img.getAttribute('data-busy') == '1') {
			e.preventDefault();
			return false;
		}

		// https://developer.mozilla.org/en-US/docs/Web/API/Element/wheel_event
		// 根据 delta 值（正、负）求得允许范围内的新的缩放值。
		this._scale += e.deltaY * -0.01;
		this._scale = Math.min(Math.max(.25, this._scale), 10);

		this.img.style.transform = `scale(${this._scale})`;

		e.preventDefault();
		return false;
	}
	_onDivMouseMove(e) {
		if (this._dragging) {
			let left = this._coordX + e.clientX - this._offsetX + 'px';
			let top = this._coordY + e.clientY - this._offsetY + 'px';
			this.img.style.left = left;
			this.img.style.top = top;
			// console.log('left & top: ', left, top);
			return false;
		}

		return true;
	}
	_onDivClick(e) {
		this._hideImage(null);
	}
}

class ImageView {
	constructor() {
		let images = document.querySelectorAll('.entry img');
		images.forEach(img => img.addEventListener('click', e => this.show(e.target)));
	}
	
	show(img) {
		new ImageViewUI(img);
	}
}

TaoBlog.imgView = new ImageView();
