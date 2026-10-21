/**
 * 
 * @param {HTMLImageElement} img 
 */
async function decryptFile(img) {
	const decrypt = async src => {
		try {
			const specResponse = await fetch(src);
			if (!specResponse.ok) {
				throw new Error(`获取解密信息失败：HTTP ${specResponse.status}`);
			}
			const spec = await specResponse.json();
			const dataResponse = await fetch(spec.src);
			if (!dataResponse.ok) {
				throw new Error(`获取加密文件失败：HTTP ${dataResponse.status}`);
			}
			const data = await dataResponse.arrayBuffer();
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
			throw new Error('图片解码失败。', { cause: e });
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
				const [url, descriptor] = item.trim().split(/\s+/, 2);
				return { url, descriptor };
			});
		};
		const sources = img.parentElement.querySelectorAll('source');
		await Promise.all(Array.from(sources, async source => {
			const parsedSrcset = parseSrcset(source.srcset);
			const updated = await Promise.all(parsedSrcset.map(async set => {
				let url;
				try {
					url = await decrypt(set.url);
				} catch {
					url = set.url;
				}
				return set.descriptor ? `${url} ${set.descriptor}` : url;
			}));
			source.srcset = updated.join(', ');
		}));
	}
}

function fixVideoCache(video) {
	if (video.onerror) {
		let url = new URL(video.src);

		// 有可能灯箱的 URL 是从原始图片拷过来（已经解码过了）。
		if(url.protocol == 'blob:') { return; }

		url.searchParams.set('_', '');
		video.src = url.toString();
	}
}
