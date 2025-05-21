async function decryptFile(img) {
	const spec = await (await fetch(img.src)).json();
	const data = await (await fetch(spec.src)).arrayBuffer();
	try {
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
		let obj = URL.createObjectURL(new Blob([decrypted]));
		img.src = obj;
		['onerror','onended'].forEach(e => { img.removeAttribute(e); });
	} catch(e) {
		console.log(img, spec, new Uint8Array(data).toBase64());
	}
}

function fixVideoCache(video) {
	if (video.onerror) {
		let url = new URL(video.src);
		url.searchParams.set('_', '');
		video.src = url.toString();
	}
}
