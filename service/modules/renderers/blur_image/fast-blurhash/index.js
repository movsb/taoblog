import { decodeBlurHash } from './fast-blurhash'

/**
 * 
 * @param {Uint8ClampedArray} pixels 
 * @returns {Promise<Blob>}
 */
async function createBlob(pixels, width, height) {
	const canvas = document.createElement('canvas');
	canvas.width = width;
	canvas.height = height;
	const ctx = canvas.getContext('2d');
	const imageData = ctx.createImageData(width, height);
	imageData.data.set(pixels);
	ctx.putImageData(imageData, 0, 0);
	return new Promise((resolve, reject) => {
		canvas.toBlob(blob => {
			if(!blob) {
				return reject('图片转换失败。');
			}
			return resolve(blob);
		});
	});
}

/**
 * 
 * @param {string} hash 
 */
async function decodeBlurHashToBlob(hash, decodeWidth, decodeHeight, renderWidth, renderHeight) {
	const pixels = decodeBlurHash(hash, decodeWidth, decodeHeight, 1);
	const blob = await createBlob(pixels, renderWidth, renderHeight);
	return URL.createObjectURL(blob);
}

window.createBlurImageFromHash = decodeBlurHashToBlob;
