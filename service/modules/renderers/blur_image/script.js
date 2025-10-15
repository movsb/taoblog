document.addEventListener('DOMContentLoaded', ()=>{
	const scale = (a,b,l,u) => {
		if(a>b) {
			return [u, Math.floor(Math.max(b/(a/u), l))];
		} else {
			return [Math.floor(Math.max(a/(b/u), l)), u];
		}
	};
	/** @type {NodeListOf<HTMLImageElement>} */
	const images = document.querySelectorAll('img[data-blurhash]:not(.static)');
	images.forEach(async img => {
		const removeBlurhash = () => {
			img.removeAttribute('data-blurhash');
		};

		if(img.complete) { return removeBlurhash(); }

		let [width, height] = [parseInt(img.width), parseInt(img.height)];
		if (width <=0 || height <= 0) { return removeBlurhash(); }

		[width, height] = scale(width, height, 1, 32);

		const url = await createBlurImageFromHash(img.dataset.blurhash, width, height, width, height);

		// createBlurImageFromHash 有等待，返回的时候可能已经 complete 了。
		if(img.complete) { return removeBlurhash(); }

		img.style.backgroundImage = `url("${url}")`;
		img.style.backgroundRepeat = 'no-repeat';
		img.style.backgroundSize = 'cover';

		img.addEventListener('load', ()=>{
			img.style.removeProperty('background-image');
			img.style.removeProperty('background-repeat');
			img.style.removeProperty('background-size');
			img.removeAttribute('data-blurhash');
		});
	});
}, {once: true});
