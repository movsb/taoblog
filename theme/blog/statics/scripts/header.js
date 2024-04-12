function __TaoBlog()
{
    function EventDispatcher()
    {
        this._event_maps = {};
    }

    EventDispatcher.prototype._get_map = function(module, name) {
        this._event_maps[module] = this._event_maps[module] || {};
        this._event_maps[module][name] = this._event_maps[module][name] || [];
        return this._event_maps[module][name];
    };

    EventDispatcher.prototype.add = function(module, name, callback) {
        var listeners = this._get_map(module, name);
        listeners.push(callback);
    };

    EventDispatcher.prototype.remove = function(module, name, callback) {
        var listeners = this._get_map(module, name);
        var index = listeners.indexOf(callback);
        if(index != -1) {
            listeners.splice(index, 1);
        }
    };

    EventDispatcher.prototype.dispatch = function(module, name)
    {
        var listeners = this._get_map(module, name);
        Array.prototype.splice.call(arguments, 0, 2);
        var args = Array.from(arguments);
        listeners.forEach(function(callback) {
            callback.apply(null, args);
        });
    };

    return {
        events: new EventDispatcher(),
        fn: {}
    };
}

var TaoBlog = window.TaoBlog = new __TaoBlog();

// Add anchors to external links and open it in new tab instead
TaoBlog.fn.externAnchor = function() {
	let hostname = location.hostname;
	return function(a) {
		if (a.href != "" && a.hostname !== hostname && !a.classList.contains('no-external')) {
			a.setAttribute('target', '_blank');
			a.classList.add('external');
		}
	};
}();

// 代码高亮
TaoBlog.fn.highlight = function(pre) {
	// 早期代码兼容：
	// <pre class="code" lang="cpp"></pre>
	{
		if (pre.classList.contains('code') && pre.getAttribute('lang') != '' && !pre.firstElementChild) {
			let code = document.createElement('code');
			code.innerHTML = pre.innerHTML;
			code.classList.add(`language-${pre.getAttribute('lang')}`);
			pre.innerHTML = '';
			pre.appendChild(code);
		}
	}
	// 必须是 <pre><code class="language-xxx"></code></pre>
	{
		let code = pre.querySelector(':scope > code');
		if (code) {
			let hasLang = false;
			code.classList.forEach(function(name) {
				if (!hasLang && /^language-/.test(name)) {
					hasLang = true;
				}
			});
			if (!hasLang) {
				code.classList.add('language-none');
			}
		}
	}

	Prism.highlightAllUnder(pre);

	// 自动滚动行号
	{
		let lines = pre.querySelector('span.line-numbers-rows');
		if(!lines) { return; }
		let div = document.createElement('div');
		div.classList.add('line-numbers-wrapper');
		let code = lines.parentElement;
		code.appendChild(div);
		lines.remove();
		div.appendChild(lines);
		code.addEventListener('scroll', function() {
			lines.style.top = `-${code.scrollTop}px`;
		});
	}
};

TaoBlog.fn.fadeIn = function(elem, callback) {
	elem.classList.remove('fade-out');
	elem.style.display = 'block';
	if (typeof callback == 'function') {
		elem.addEventListener('animationend', function(event) {
			// console.log('fade-in animationend');
			callback();
		}, { once: true});
	}
	elem.classList.add('fade-in');
};
TaoBlog.fn.fadeOut = function(elem, callback) {
	elem.classList.remove('fade-in');
	elem.addEventListener('animationend', function(event) {
		// console.log('fade-out animationend');
		elem.style.display = 'none';
		if (typeof callback == 'function') {
			callback();
		}
	}, { once: true});
	elem.classList.add('fade-out');
};
