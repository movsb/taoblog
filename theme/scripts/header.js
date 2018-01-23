
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
        var listners = this._get_map(module, name);
        listners.push(callback);
    };

    EventDispatcher.prototype.remove = function(module, name, callback) {
        var listners = this._get_map(module, name);
        var index = listners.indexOf(callback);
        if(index != -1) {
            listners.splice(index, 1);
        }
    };

    EventDispatcher.prototype.dispatch = function(module, name)
    {
        var listners = this._get_map(module, name);
        Array.prototype.splice.call(arguments, 0, 2);
        var args = Array.from(arguments);
        listners.forEach(function(callback) {
            callback.apply(null, args);
        });
    };

    return {
        events: new EventDispatcher(),
    };
}

var TaoBlog = window.TaoBlog = new __TaoBlog();

