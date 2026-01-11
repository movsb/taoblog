/**
 * 
 * @param {Response} rsp 
 */
async function decodeAPIError(rsp) {
	let exception = await rsp.text();
	try { exception = JSON.parse(exception); }
	catch {}
	return new Error(
		exception.message ?? exception,
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
	return new Promise((resolve) => {
		r.onload = () => {
			resolve(r.result);
		};
		r.readAsDataURL(f);
	});
}
