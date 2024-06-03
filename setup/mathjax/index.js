// 参考：
// https://github.com/mathjax/MathJax-demos-node#MathJax-demos-node
// https://github.com/sparanoid/mathjax-node-server/blob/master/index.js

const http = require('http');
const url = require('url');

require('mathjax').init({
	loader: {
		load: ['input/tex', 'output/svg']
	},
}).then((MathJax) => {
	const server = http.createServer((r, w) => {
		MathJax.tex2svgPromise(String.raw`a`, {
			display: true,
		}).then((node) => {
			const adaptor = MathJax.startup.adaptor;
			let html = adaptor.outerHTML(node);
			w.setHeader('Content-Type', 'image/svg+xml');
			w.end(html);
		});
	});
	server.listen(31415);
}).catch(err => console.log(err));
