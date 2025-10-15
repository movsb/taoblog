/**
 * 
 * @param {Response} rsp 
 */
async function throwAPIError(rsp) {
	let exception = await rsp.text();
	try { exception = JSON.parse(exception); }
	catch {}
	throw exception.message ?? exception;
}

/**
 * 尝试解析 JSON 响应，如果失败则抛出异常。
 * @param {Response} rsp 
 */
async function decodeResponse(rsp) {
	if(!rsp.ok) {
		await throwAPIError(rsp);
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
	return new Promise((resolve) => {
		r.onload = () => {
			resolve(r.result);
		};
		r.readAsDataURL(f);
	});
}
