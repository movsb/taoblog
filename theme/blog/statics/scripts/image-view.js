class ImageView {
	constructor() {}

	init() {
		this._initView();
		
		this._rootDiv   = document.getElementById('img-view');
		this._elemImg   = this._rootDiv.querySelector(':scope > img');
		this._elemInfo  = this._rootDiv.querySelector(':scope .info');
		this._elemTip   = this._rootDiv.querySelector(':scope .tip');

		this._imageIndex = -1;
		this._scale = 1;
		this._transform = {};

		this._countDown = 3;
		setInterval(function () {
			if (--this._countDown == 0) {
				this._elemInfo.style.display = 'none';
			}
		}.bind(this), 1000);

		this._keyHandlerAdded = false;

		this._boundEvents = {};

		this._initBindings();
		
		let images = document.querySelectorAll('.entry img');
		let self = this;
		images.forEach(function(img) {
			img.addEventListener('click', self._onPostImageClick.bind(self));
		});
		this._images = images;

		this._retina = false;
	}
	_onPostImageClick(e) {
		this._locateIndex(e.target);
		this.viewImage(e.target);
	}
	_keyHandler(e) {
		var handled = true;

		if (e.keyCode == 27 || e.keyCode == 32) {
			// Escape or Space to exit
			this.viewImage(null);
		} else if (e.keyCode == 37) {
			// left to previous
			this.nextImage(-1);
		} else if (e.keyCode == 39) {
			// right to next
			this.nextImage(1);
		} else if (e.keyCode == 38) {
			// up to rotate anticlockwisely
			this.rotateImage(false);
		} else if (e.keyCode == 40) {
			// down to rotate clockwisely
			this.rotateImage(true);
		} else {
			handled = false;
		}

		if (handled) {
			e.preventDefault();
		}
	}
	_locateIndex(img) {
		for (var i = 0; i < this._images.length; i++) {
			if (img == this._images[i]) {
				this._imageIndex = i;
				break;
			}
		}
	}
	_showInfo(rawWidth, rawHeight, scale) {
		let s = `第 ${this._imageIndex+1}/${this._images.length} 张`
		+ `，原始尺寸：${rawWidth}x${rawHeight}，缩放比例：${Math.round(scale*100)}%`;
		
		this._elemInfo.innerText = s;
		this._elemInfo.style.display = 'block';
		this._countDown = 3;
	}
	nextImage(dir) {
		this.rotateImage(undefined); // clear rotation first
		this._imageIndex += dir > 0 ? 1 : -1;
		if (this._imageIndex > this._images.length - 1) {
			this._imageIndex = this._images.length - 1;
		}
		if (this._imageIndex < 0) {
			this._imageIndex = 0;
		}
		if (this._imageIndex < this._images.length) {
			this.viewImage(this._images[this._imageIndex]);
		}
	}
	transforms() {
		var s = '';
		for (var p in this._transform) {
			s += p + '(' + this._transform[p] + ') ';
		}
		return s;
	}
	rotateImage(fwd) {
		let img = this._elemImg;
		if (fwd == undefined) {
			this._transform = {};
			img.style.transform = this.transforms();
			this._degree = 0;
			this._scale = 1;
		} else {
			this._degree += !!fwd ? +90 : -90;
			img.setAttribute('data-busy', '1');
			img.style.transition = 'transform 0.3s linear';
			this._transform.rotateZ = `${this._degree}deg`;
			img.style.transform = this.transforms();
		}
	}
	viewImage(img) {
		if (img != null) {
			// tips
			let tip_times_total = 1;
			let tip_times = +this._elemImg.getAttribute('data-times') || 0;

			if (tip_times == 0) {
				this._elemTip.innerText = '左键拖动，上中下键旋转，滚动缩放，左右切换；单击图片或空白区域退出。';
			}

			if (tip_times < tip_times_total) {
				tip_times++;
				this._elemImg.setAttribute('data-times', tip_times);
			} else if (tip_times == tip_times_total) {
				this._elemTip.style.display = 'none';
			}
			
			this._elemImg.addEventListener('load', function() {
				this._rootDiv.style.display = 'block';

				// console.log(this._rootDiv.clientWidth);
				this._rawWidth = this._elemImg.naturalWidth;
				this._rawHeight = this._elemImg.naturalHeight;

				let initScale = this._retina ? 0.5 : 1;
				let initWidth = this._rawWidth * initScale;
				let initHeight = this._rawHeight * initScale;

				{
					let scaleWidth  = this._rootDiv.clientWidth  / initWidth;
					let scaleHeight = this._rootDiv.clientHeight / initHeight;

					// if smaller than container
					if (scaleWidth >= 1 && scaleHeight >= 1) {
						initScale *= 1;
					}

					// if larger than container, scale to fit
					else {
						initScale *= Math.min(scaleWidth, scaleHeight);
					}

					initWidth = this._rawWidth * initScale;
					initHeight = this._rawHeight * initScale;
				}

				let img = this._elemImg;
				img.style.left      = `${(this._rootDiv.clientWidth  - initWidth) /2}px`;
				img.style.top       = `${(this._rootDiv.clientHeight - initHeight)/2}px`;
				img.style.width     = `${initWidth}px`;
				img.style.height    = `${initHeight}px`;
				img.style.display   = 'block';
				
				let body =document.body;
				body.style.maxHeight = window.innerHeight + 'px';
				body.style.overflow = 'hidden';

				this._showInfo(this._rawWidth, this._rawHeight, initScale);
			}.bind(this));

			this._elemImg.onerror = function () {
				this._showInfo(0, 0, 0);
			}.bind(this);

			let src = img.src;
			if (img.classList.contains('transparent')) {
				this._elemImg.classList.add('transparent');
			} else {
				this._elemImg.classList.remove('transparent');
			}
			this._retina = src.indexOf('@2x.') != -1;
			this._elemImg.src = src;
		} else {
			// 以下两行清除因拖动导致的设置
			this._elemImg.style.left = '0';
			this._elemImg.style.top = '0';
			document.body.style.maxHeight = 'none';
			document.body.style.overflow = 'auto';

			// 清除旋转
			this._transform = {};
			this._elemImg.style.transform = this.transforms();

			this._dragging = false;
			this._scale = 1;
			this._degree = 0;
			this._elemImg.style.display = 'none';
			this._rootDiv.style.display = 'none';

			// 清除透明
			this._elemImg.classList.remove('transparent');
		}

		if (img != null) {
			if (!this._keyHandlerAdded) {
				window.addEventListener('keydown', this._boundEvents.keyHandler);
				this._keyHandlerAdded = true;
			}
		} else {
			if (this._keyHandlerAdded) {
				window.removeEventListener('keydown', this._boundEvents.keyHandler);
				this._keyHandlerAdded = false;
			}
		}
	}

	_initView() {
		let html = `
<div class="img-view" id="img-view">
	<img />
	<div class="tool">
		<div class="info"></div>
		<div class="tip"></div>
	</div>
`;
		let div = document.createElement('div');
		div.innerHTML = html;
		document.body.appendChild(div.firstElementChild);
	}

	_initBindings() {
		this._boundEvents.keyHandler = this._keyHandler.bind(this);

		let imgHandlers = {
			'mousedown':        this._onImgMouseDown,
			'mousemove':        this._onImgMouseMove,
			'mouseup':          this._onImgMouseUp,
			'click':            this._onImgClick,
			'dblclick':         this._onImgDblClick,
			'transitionend':    this._onTransitionEnd,
		};
		for (let key in imgHandlers) {
			this._elemImg.addEventListener(key, imgHandlers[key].bind(this));
		}
		
		let divHandlers = {
			'wheel':        this._onDivMouseWheel,
			'mousemove':    this._onDivMouseMove,
			'click':        this._onDivClick,
		};
		for (let key in divHandlers) {
			this._rootDiv.addEventListener(key, divHandlers[key].bind(this));
		}
	}
	_onImgMouseDown(e) {
		if (this._elemImg.getAttribute('data-busy') == '1') {
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
		} else if (e.which == 2) { // middle button
			this.rotateImage(true);
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
			this.viewImage(null);
			console.log('hide because of small move');
		}

		e.preventDefault();
		e.stopPropagation();
		return false;
	}
	_onImgDblClick(e) {
		this.viewImage(null);
		e.preventDefault();
		return false;
	}
	_onTransitionEnd(e) {
		this._elemImg.style.transition = '';
		this._elemImg.setAttribute('data-busy', '');
	}
	_onDivMouseWheel(e) {
		if (this._elemImg.getAttribute('data-busy') == '1') {
			e.preventDefault();
			return false;
		}

		// https://developer.mozilla.org/en-US/docs/Web/API/Element/wheel_event
		// 根据 delta 值（正、负）求得允许范围内的新的缩放值。
		this._scale += e.deltaY * -0.01;
		this._scale = Math.min(Math.max(.25, this._scale), 5);

		this._transform.scale = this._scale;
		this._elemImg.style.transform = this.transforms();

		this._showInfo(this._rawWidth, this._rawHeight, this._scale);

		e.preventDefault();
		return false;
	}
	_onDivMouseMove(e) {
		if (this._dragging) {
			let left = this._coordX + e.clientX - this._offsetX + 'px';
			let top = this._coordY + e.clientY - this._offsetY + 'px';
			this._elemImg.style.left = left;
			this._elemImg.style.top = top;
			// console.log('left & top: ', left, top);
			return false;
		}

		return true;
	}
	_onDivClick(e) {
		this.viewImage(null);
	}
}

var imgView = new ImageView();
imgView.init();
