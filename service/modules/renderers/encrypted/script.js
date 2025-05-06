async function decodeFile(id) {
	let img = document.querySelector(`img[data-id=${id}]`);
	if (!img) { return; }
	let data = await (await fetch(img.src)).arrayBuffer();
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
}
