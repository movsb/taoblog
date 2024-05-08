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

// 代码高亮
// TODO 早期代码兼容：
// <pre class="code" lang="cpp"></pre>
// 处理成 `lang-xxx` 形式。

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

TaoBlog.fn.getUserID = function() {
	let matches = /taoblog\.user_id=(\d+)/.exec(document.cookie);
	if (matches && matches.length == 2) {
		return +matches[1];
	}
	return 0;
};

TaoBlog.userID = TaoBlog.fn.getUserID();
