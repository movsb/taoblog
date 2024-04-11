/* pre 的双击全选与全窗口 */
(function() {
    $('.entry pre, .entry code').on('dblclick', function(e) {
        var t = e.target.tagName;
        if(t == 'PRE' || t == 'CODE') {
            var selection = window.getSelection();
            var range = document.createRange();
            range.selectNodeContents(e.target);
            selection.removeAllRanges();
            selection.addRange(range);
            e.preventDefault();
            return;
        }
    });
})();

// 图片来源
(function() {
    function replace(img) {
        var figure = document.createElement('figure');

        var newImg = img.cloneNode(false);
        newImg.removeAttribute('data-name');
        newImg.removeAttribute('data-origin');
        figure.appendChild(newImg);

        var caption = document.createElement('figcaption');
        caption.innerText = '图片来源：';

        var a = document.createElement('a');
        a.innerText = img.getAttribute('data-name') || img.getAttribute('alt');
        a.setAttribute('href', img.getAttribute('data-origin'));
        a.setAttribute('target', '_blank');

        caption.appendChild(a);

        figure.appendChild(caption);

        img.parentNode.insertBefore(figure, img);
        img.parentNode.removeChild(img);
    }

    var imgs = document.querySelectorAll('.entry img[data-origin]');
    imgs.forEach(function(img){ replace(img); });
})();

(function() {
	let anchors = document.querySelectorAll('article a');
	anchors.forEach(a => { TaoBlog.fn.externAnchor(a); });
})();

// 数学公式
(function(){
function hasMath() {
	var math = document.querySelector('.math');
	if (math !== null) {
		return true;
	}

    var has = false;

    // old-style, to-be-removed
	let codes = document.querySelectorAll('code:not([class*="lang"])');
	codes.forEach(function(e) {
		var t = e.innerHTML;
		if(t.startsWith('$') && t.endsWith('$')) {
			has = true;
		}
	});
    return has;
}

if(hasMath()) {
	let s = document.createElement('script');
    s.src = '/plugins/mathjax/mathjax.js';
    s.async = true;
    document.body.appendChild(s);
}
})();

// 代码高亮
(function(){
let s = document.createElement('script');
s.async = false;
s.setAttribute('data-manual', '');
s.onload = function() {
	document.querySelectorAll('pre').forEach(function(pre) {
		TaoBlog.fn.highlight(pre);
	});
};
s.src = '/plugins/highlight/prism.js';
document.body.appendChild(s);

TaoBlog.events.add('comment', 'post', function(jItem) {
	let pres = jItem[0].querySelectorAll(':scope > .comment-content pre');
	pres.forEach(function(pre) {
		TaoBlog.fn.highlight(pre);
	});
	let anchors = jItem[0].querySelectorAll(':scope a');
	anchors.forEach(a => { TaoBlog.fn.externAnchor(a); });
});
})();

// TODO 目前的 Markdown 在处理 @2x 图片时无法处理 HTML 标签使用的图片 <img>，只能处理 ![]() 这种。
// 这里使用脚本临时处理一下，后续应该在 Markdown 里面统一处理。
(function(){
	let imgs = document.querySelectorAll('img');
	imgs.forEach(function(img) {
		if(img.src.indexOf('@2x.') >= 0 && img.style.width == '' && img.style.height == '') {
			img.addEventListener('load', function() {
				if (img.naturalWidth > 0 && img.naturalHeight > 0) {
					img.style.width = `${img.naturalWidth/2}px`;
					img.style.height = `${img.naturalHeight/2}px`;
				}
			});
		}
	});
})();

// 简易的 Vim 按键模拟。
class __Vim {
	constructor() {
		this.maps  = {};    // 按键绑定映射
		this.tree  = {};    // TRIE 搜索树
		this.stack = [];    // 按键栈
		this.timer = null;  // 定时清理掉无效的按键

		this.init();
	}

	init() {
		document.body.addEventListener('keypress', (function (e) {
			if (this.timer) {
				clearInterval(this.timer);
				this.timer = null;
			}

			if (e.target.tagName != 'BODY') {
				this.stack = [];
				return;
			}

			this.stack.push(e.key);
			this.trigger();

			if (this.stack.length) {
				this.timer = setInterval(() => {
					if (this.stack.length > 0) {
						this.stack = [];
						console.log('key stack cleared');
					}
				}, 1000);
			}
		}).bind(this));
		
		this.maps = {
			gg: function() {
				window.scrollTo({left: 0, top: 0, behavior: 'smooth'});
			},
			G: function() {
				window.scrollTo({left: 0, top: document.body.scrollHeight, behavior: 'smooth'}); 
			},
			j: function() {
				window.scrollBy({left: 0, top: +150, behavior: 'smooth'});
			},
			k: function() {
				window.scrollBy({left: 0, top: -150, behavior: 'smooth'});
			},
			f: function() {
				if (document.fullscreenElement) {
					document.exitFullscreen();
				} else {
					document.documentElement.requestFullscreen();
				}
			},
			w: function() {
				document.body.classList.toggle("wide");
			},
			r: function() {
				location.reload();
			},
			b: function() {
				location.pathname = '/';
			},
			'?': function() {
				console.log('Vim Help');
				console.table({
					gg: '回到页首',
					G: '回到页尾',
					j: '向下滚动',
					k: '向上滚动',
					f: '进入全屏',
					w: '进入宽屏模式',
					r: '刷新',
					b: '回到首页',
				});
			},
		};

		for (let key in this.maps) {
			let node = this.tree;
			for (let i in key) {
				if (!node[key[i]]) {
					node[key[i]] = {};
				}
				node = node[key[i]];
			}
			node.__handler = this.maps[key];
		}
	}

	trigger() {
		let node = this.tree;
		console.log('stack:', this.stack);

		// 遍历树以寻找匹配按键序列的按键映射/绑定。
		for (let i = 0; i < this.stack.length; i++) {
			let child = node[this.stack[i]];
			if (!child) {
				node = null;
				break;
			}
			node = child;
		}

		// 说明根本没有这个按键映射，
		// 属于无效的按键映射，清空。
		if (!node) {
			console.log('no such key binding:', this.stack);
			this.stack = [];
			return;
		}

		// 按键组合还没有到达最后一个按键。
		 if (!node.__handler) {
			return;
		}

		// console.log(node);
		console.log('triggering:', node);
		node.__handler.call(this);
		this.stack = [];
	}
}

TaoBlog.vim = new __Vim();

// 加载字体
(function() {
	let link = document.createElement('link');
	link.rel = 'stylesheet';
	link.type = 'text/css';
	link.href = 'https://npm.elemecdn.com/lxgw-wenkai-webfont/style.css';
	document.head.appendChild(link);
}) ();
