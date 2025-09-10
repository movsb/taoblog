/**
 * 
 * @param {HTMLImageElement} img 
 */
async function decryptFile(img) {
	const decrypt = async src => {
		try {
			const spec = await (await fetch(src)).json();
			const data = await (await fetch(spec.src)).arrayBuffer();
			const key = await crypto.subtle.importKey(
				'raw', Uint8Array.fromBase64(spec.key),
				{ name: 'AES-GCM'},
				false, ['decrypt']
			);
			let decrypted = await crypto.subtle.decrypt(
				{
					name: 'AES-GCM',
					iv: Uint8Array.fromBase64(spec.nonce),
				},
				key,
				data,
			);
			const contentType = spec.content_type ?? '';
			return URL.createObjectURL(new Blob([decrypted], { type: contentType }));
		} catch(e) {
			throw new Error('图片解码失败。');
		}
	};

	img.src = await decrypt(img.src);
	['onerror','onended'].forEach(e => { img.removeAttribute(e); });

	// 如果父元素是 <picture>，也需要解析 <source>。
	// [<source>: The Media or Image Source element - HTML | MDN](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/source#srcset)
	if(img.parentElement?.tagName == 'PICTURE') {
		// 不完全合规的解析。
		/**
		 * 
		 * @param {string} srcset 
		 * @returns {Array<{url: string, descriptor: string}>}
		 */
		const parseSrcset = (srcset) => {
			return srcset.split(",").map(item => {
				const [url, descriptor] = item.trim().split(/\s+/, 2)
				return { url, descriptor }
			})
		}
		const sources = img.parentElement.querySelectorAll('source');
		sources.forEach(async source => {
			const parsedSrcset = parseSrcset(source.srcset);
			/** @type {Array<Promise<string> | string} */
			const promises = [];
			parsedSrcset.forEach(async set => {
				try {
					promises.push(decrypt(set.url));
				} catch {
					promises.push(set.url);
				}
			});
			const urls = await Promise.all(promises);
			const updated = [];
			parsedSrcset.forEach((set, index) => {
				let single = `${urls[index]}`;
				if(set.descriptor) single += ` ${set.descriptor}`;
				updated.push(single);
			});
			source.srcset = updated.join(', ');
		});
	}
}

function fixVideoCache(video) {
	if (video.onerror) {
		let url = new URL(video.src);
		url.searchParams.set('_', '');
		video.src = url.toString();
	}
}
