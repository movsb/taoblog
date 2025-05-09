async function decodeFile(id) {
	let img = document.querySelector(`img[data-id=${id}],video[data-id=${id}]`);
	if (!img) { return; }
	const newURL = await (await fetch(img.src)).text();
	const data = await (await fetch(newURL)).arrayBuffer();
	try {
		const key = await crypto.subtle.importKey(
			'raw', Uint8Array.fromBase64(img.dataset.key),
			{ name: 'AES-GCM'},
			false, ['decrypt']
		);
		let decrypted = await crypto.subtle.decrypt(
			{
				name: 'AES-GCM',
				iv: Uint8Array.fromBase64(img.dataset.nonce),
			},
			key,
			data,
		);
		let obj = URL.createObjectURL(new Blob([decrypted]));
		img.src = obj;
		['data-id','data-key','data-nonce','onerror'].forEach(e => { img.removeAttribute(e); });
	} catch(e) {
		console.log(e);
		console.log(id, img.src, img.dataset.key, img.dataset.nonce, new Uint8Array(data).toBase64());
	}
}
