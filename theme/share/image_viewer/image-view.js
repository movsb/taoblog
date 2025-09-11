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

			// 实况照片本身要宽高限制，不适于设置 100%，包一层。
			if(clone.classList.contains('live-photo')) {
				const div = document.createElement('div');
				div.classList.add('live-photo-wrapper');
				div.appendChild(clone);
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
	}
	hide() {
		this.root.style.display = 'none';
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
});
