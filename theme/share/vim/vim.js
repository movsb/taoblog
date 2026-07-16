class __Vim {
	static isMac = navigator.platform.includes('Mac');

	constructor() {
		this.maps  = {};    // 按键绑定映射
		this.tree  = {};    // TRIE 搜索树
		this.stack = [];    // 按键栈
		this.timer = null;  // 定时清理掉无效的按键

		this.init();
	}

	init() {
		// 'keypress' 事件被废弃了，改用 'keydown' 事件。
		document.body.addEventListener('keydown', (function (e) {
			if (this.timer) {
				clearInterval(this.timer);
				this.timer = null;
			}

			if (e.target.tagName != 'BODY') {
				this.stack = [];
				return;
			}

			this.stack.push(e.key);
			this.trigger(e);

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
				location = '/';
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

		for (let keys in this.maps) {
			this.bind(keys, this.maps[keys]);
		}
	}

	/**
	 * 绑定按键映射。
	 * 
	 * 格式：`[ca]q`，表示：ctrl + alt + q。
	 * 
	 * 中括号内的可以省略；大小写直接用大、小写字母表示。
	 * 
	 * @param {string} keys 
	 * @param {function} handler 
	 */
	bind(keys, handler) {
		let ctrl = false, alt = false;

		if(keys[0] == '[') {
			let i = keys.indexOf(']');
			if(i == -1) {
				throw new Error('Invalid key binding: ' + keys);
			}
			for(let j = 1; j < i; j++) {
				switch(keys[j]) {
					case 'c':
						ctrl = true;
						break;
					case 'a':
						alt = true;
						break;
					default:
						throw new Error('Invalid key binding: ' + keys);
				}
			}
			keys = keys.slice(i + 1);
		}

		let node = this.tree;
		for (let i in keys) {
			if (!node[keys[i]]) {
				node[keys[i]] = {};
			}
			node = node[keys[i]];
		}
		node._handler = handler;
		node._ctrl = ctrl;
		node._alt = alt;
	}

	/**
	 * 
	 * @param {KeyboardEvent} e 
	 * @returns 
	 */
	trigger(e) {
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
		 if (!node._handler) {
			return;
		}

		// 修饰键需要匹配，否则清空。
		if(node._ctrl == e.ctrlKey && (__Vim.isMac ? node._alt == e.metaKey : node._alt == e.altKey)) {
			e.preventDefault();
			e.stopPropagation();

			console.log('triggering:', node);
			node._handler.call(this);
		}

		this.stack = [];
	}
}

document.addEventListener('DOMContentLoaded', function() {
	const vim = (TaoBlog||window).vim = new __Vim();

	const edit = document.querySelector('.edit-button > a');
	if (edit) { vim.bind('e', edit.click.bind(edit)); }

	vim.bind('a', async ()=>{
		try {
			let wa = new WebAuthn();
			await wa.login();
			location.reload();
		} catch(e) {
			if (e instanceof DOMException && ["NotAllowedError", "AbortError"].includes(e.name)) {
				console.log('已取消操作。');
				return;
			}
			alert(e);
		}
	});

	vim.bind('n', () => {
		if(location.pathname != '/admin/editor') {
			location.href='/admin/editor?new=1&type=markdown';
		}
	});
}, {once: true});
