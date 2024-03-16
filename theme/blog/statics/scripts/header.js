
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
TaoBlog.fn.highlight = function(re) {
	var e = $(re);
	var lang = e.attr('lang');
	// https://stackoverflow.com/a/1318091/3628322
	var hasLang = typeof lang !== typeof undefined && lang !== false;
	var hasCode = e.find('>code').length > 0;
	// console.log(re, hasLang, hasCode);
	if(hasLang && !hasCode) {
		let code = $('<code/>').html(e.html());
		code.addClass("language-" + lang);
		e.removeAttr('lang');
		e.html('');
		e.append(code);
		hasCode = true;
	}
	if (re.classList.length == 0) {
		re.classList.add('language-none');
	}
	if(hasCode) {
		console.log('有代码');
		e.removeClass('code');
		// TODO
		// e.addClass('line-numbers');
		Prism.highlightAllUnder(re);
	} else {
		console.log('没有 code', re);
	}

	let lines = re.querySelector('span.line-numbers-rows');
	if(lines === null) {
		console.log('没有行号');
		return;
	}
	let div = document.createElement('div');
	div.classList.add('line-numbers-wrapper');
	let code = lines.parentElement;
	code.appendChild(div);
	lines.remove();
	div.appendChild(lines);
	code.addEventListener('scroll', function() {
		lines.style.top = '-' + code.scrollTop + 'px';
	});
};

TaoBlog.fn.fadeIn = function(elem, callback) {
	elem.classList.remove('fade-out');
	elem.style.display = 'block';
	if (typeof callback == 'function') {
		elem.addEventListener('animationend', function(event) {
			callback();
		}, { once: true});
	}
	elem.classList.add('fade-in');
};
TaoBlog.fn.fadeOut = function(elem, callback) {
	elem.classList.remove('fade-in');
	elem.addEventListener('animationend', function(event) {
		elem.style.display = 'none';
		if (typeof callback == 'function') {
			callback();
		}
	}, { once: true});
	elem.classList.add('fade-out');
};
