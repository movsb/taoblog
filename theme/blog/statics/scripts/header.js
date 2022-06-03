
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
