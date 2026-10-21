/**
 * 
 * @param {Response} rsp 
 */
async function decodeAPIError(rsp) {
	let exception = await rsp.text();
	try { exception = JSON.parse(exception); }
	catch {}
	const message = typeof exception?.message === 'string' ? exception.message
		: typeof exception === 'string' ? exception : '';
	return new Error(
		message.trim() || `HTTP ${rsp.status}${rsp.statusText ? ' ' + rsp.statusText : ''}`,
		{ cause: rsp },
	);
}

/**
 * 尝试解析 JSON 响应，如果失败则抛出异常。
 * 会自动判断响应是否 ok。
 * @param {Response} rsp 
 */
async function decodeResponse(rsp) {
	if(!rsp.ok) {
		throw await decodeAPIError(rsp);
	}
	return rsp.json();
}

/**
 * 
 * @param {File} f 
 * @returns {Promise<string>}
 */
async function readFileAsDataURL(f) {
	const r = new FileReader();
	return new Promise((resolve, reject) => {
		r.onload = () => {
			resolve(r.result);
		};
		r.onerror = () => {
			reject(r.error ?? new Error('文件读取失败。'));
		};
		r.onabort = () => {
			reject(new DOMException('文件读取已取消。', 'AbortError'));
		};
		r.readAsDataURL(f);
	});
}
