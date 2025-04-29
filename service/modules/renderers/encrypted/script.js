async function decodeFile(id) {
	let img = document.querySelector(`img[data-id=${id}]`);
	if (!img) { return; }
	let enc = JSON.parse(img.dataset.encryption);
	let data = await (await fetch(img.src)).arrayBuffer();
	const key = await crypto.subtle.importKey(
		'raw', Uint8Array.fromBase64(enc.Key),
		{ name: 'AES-GCM'},
		false, ['decrypt']
	);
	let decrypted = await crypto.subtle.decrypt(
		{
			name: 'AES-GCM',
			iv: Uint8Array.fromBase64(enc.Nonce),
		},
		key,
		data,
	);
	let obj = URL.createObjectURL(new Blob([decrypted]));
	img.src = obj;
}
