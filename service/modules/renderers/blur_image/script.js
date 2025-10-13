document.addEventListener('DOMContentLoaded', ()=>{
	const scale = (a,b,l,u) => {
		if(a>b) {
			return [u, Math.floor(Math.max(b/(a/u), l))];
		} else {
			return [Math.floor(Math.max(a/(b/u)), l)];
		}
	};
	/** @type {NodeListOf<HTMLImageElement>} */
	const images = document.querySelectorAll('img[data-blurhash]:not(.static)');
	images.forEach(async img => {
		if(img.complete) { return; }

		let [width, height] = [parseInt(img.width), parseInt(img.height)];
		if (width <=0 || height <= 0) { return; }

		[width, height] = scale(width, height, 1, 32);
		// console.log(width, height);

		const url = await createBlurImageFromHash(img.dataset.blurhash, width, height, width, height);
		img.style.backgroundImage = `url("${url}")`;
		img.style.backgroundRepeat = 'no-repeat';
		img.style.backgroundSize = 'cover';

		img.addEventListener('load', ()=>{
			img.style.removeProperty('background-image');
			img.style.removeProperty('background-repeat');
			img.style.removeProperty('background-size');
		});
	});
});
