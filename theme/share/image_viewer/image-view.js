class ImageView {
	/**
	 * @param {Array<HTMLImageElement|HTMLPictureElement|HTMLDivElement>} objects 
	 */
	constructor(objects) {
		const div = document.createElement('div');
		div.innerHTML = `<div id="img-view"></div>`;
		/** @type {HTMLDivElement} */
		this.root = div.firstElementChild;
		document.body.appendChild(this.root);

		objects.forEach((obj, index)  => {
			let clone = obj.cloneNode(true);

			/** @type {HTMLImageElement} */
			const img = clone.querySelector('img') ?? clone;

			// 不再使用 blur 类。
			img.classList.remove('blur');

			// 移除封面效果。
			img.style.removeProperty('object-fit');
			img.style.removeProperty('aspect-ratio');
			img.style.removeProperty('width');

			// 移除模糊预览图。
			img.style.removeProperty('background-image');
			img.style.removeProperty('background-repeat');
			img.style.removeProperty('background-size');

			this._initMetadata(img.dataset.metadata, img);

			// 实况照片本身要宽高限制，不适于设置 100%，包一层。
			if(clone.classList.contains('live-photo')) {
				// 防止被 live photo js 把克隆的也处理了，因为执行顺序不确定。
				clone.classList.add('clone');

				const div = document.createElement('div');
				div.classList.add('live-photo-wrapper');
				div.appendChild(clone);

				// 预览的时候是全屏的，为了更醒目，把图标提出来。
				/** @type {HTMLDivElement} */
				const icon = clone.querySelector('.icon');
				icon.remove();
				icon.classList.add('live-photo-icon');
				div.appendChild(icon);

				// 一定是在 DOMContentLoaded 里面执行的，执行时脚本已经执行完成，所以函数一定存在。
				livePhotoBindEvents(clone, icon, div);

				clone = div;
			}
			this.root.appendChild(clone);

			const clickable = obj instanceof HTMLImageElement ? obj : obj.querySelector('img');
			clickable.addEventListener('click', (e) => {
				this.view(index, clone);
				e.preventDefault();
				e.stopPropagation();
			});

			clone.addEventListener('click', (e) => {
				this.hide();
				e.preventDefault();
				e.stopPropagation();
			});
		});

		let resizeDebouncer = null;
		window.addEventListener('resize', () => {
			if(resizeDebouncer) clearTimeout(resizeDebouncer);
			resizeDebouncer = setTimeout(()=>{
				this._placeLivePhotos();
				resizeDebouncer = null;
			}, 500);
		});
		this._placeLivePhotos();

		this._keydownHandlerBound = this._keydownHandler.bind(this);
	}
	/**
	 * 
	 * @param {number} index 
	 * @param {HTMLImageElement | HTMLPictureElement | HTMLDivElement} obj 
	 */
	view(index, obj) {
		console.log('viewing object:', index);
		this.root.style.opacity = 0;
		this.root.style.display = 'flex';

		// 真实宽度可能是小数，通过 clientWidth 拿不到。
		// 这样会导致越往后滚的时候偏离越大，这里换用 getComputedStyle。
		const gap = 10;
		// const width = this.root.clientWidth;
		const width = parseFloat(getComputedStyle(this.root).width) 
			|| this.root.clientWidth;

		this.root.scrollLeft = (width+gap) * index;
		// alert(`图片索引：${index}, 滚动：${(width+gap)*index}, 真实：${this.root.scrollLeft}，客户区：${this.root.clientWidth}`);

		this.root.style.opacity = 1;

		if(obj.querySelector('div.live-photo')) {
			/** @type {HTMLVideoElement} */
			const video = obj.querySelector('video');
			video.load();
		}

		document.addEventListener('keydown', this._keydownHandlerBound);
	}
	hide() {
		this.root.style.display = 'none';

		document.removeEventListener('keydown', this._keydownHandlerBound);
	}

	/**
	 * 
	 * @param {KeyboardEvent} e 
	 */
	_keydownHandler(e) {
		console.log('image-view:', e);
		if(e.key == 'Escape') {
			this.hide();
			e.preventDefault();
			e.stopPropagation();
		}
	}

	_placeLivePhotos() {
		const maxWidth = window.innerWidth, maxHeight = window.innerHeight;
		/** @type {HTMLDivElement[]} */
		const livePhotos = this.root.querySelectorAll('div.live-photo');
		livePhotos.forEach(p => {
			const width = parseInt(p.style.width);
			const height = parseInt(p.style.height);
			const scaledWidth = maxHeight / height * width;
			const scaledHeight = maxWidth / width * height;
			let w = 0, h = 0;
			if(scaledWidth > maxWidth) {
				w = maxWidth;
				h = scaledHeight;
			} else if(scaledHeight > maxHeight) {
				w = scaledWidth;
				h = maxHeight;
			}
			if(w && h) {
				p.style.setProperty('width', `${w}px`, 'important');
				p.style.setProperty('height', `${h}px`, 'important');
			}
		});
	}

	_initMetadata(metadata, to) {
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
		to.title = title;
	}
}

document.addEventListener('DOMContentLoaded', () => {
	/** @type {HTMLImageElement[]} */
	let images = document.querySelectorAll('.entry img:not(.static)');
	let objects = Array.from(images).map(img => {
		const picture = img.closest('picture');
		if(picture) return picture;
		const livePhoto = img.closest('div.live-photo');
		if(livePhoto) return livePhoto;
		return img;
	});
	window.TaoBlog.imgView = new ImageView(objects);
}, {once: true});
