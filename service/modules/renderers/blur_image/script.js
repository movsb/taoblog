/**
 * @import { thumbHashToDataURL } from "./thumbhash/thumbhash";
 */

document.addEventListener('DOMContentLoaded', ()=>{
	/** @type {NodeListOf<HTMLImageElement>} */
	const images = document.querySelectorAll('img[data-thumb-hash]:not(.static)');
	images.forEach(async img => {
		const removeHash = () => {
			img.removeAttribute('data-thumb-hash');
		};

		if(img.complete) { return removeHash(); }

		const bytes = Uint8Array.fromBase64(img.dataset.thumbHash);
		const dataURL = thumbHashToDataURL(bytes);

		img.style.backgroundImage = `url("${dataURL}")`;
		img.style.backgroundRepeat = 'no-repeat';
		img.style.backgroundSize = 'cover';

		img.addEventListener('load', ()=>{
			img.style.removeProperty('background-image');
			img.style.removeProperty('background-repeat');
			img.style.removeProperty('background-size');

			removeHash();
		});
	});
}, {once: true});
