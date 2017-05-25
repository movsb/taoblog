
function __TaoBlog()
{
    function EventDispatcher()
    {
        this._event_maps = {};
    }

    EventDispatcher.prototype._get_map = function(name) {
        return this._event_maps[name] = this._event_maps[name] || [];
    };

    EventDispatcher.prototype.add = function(name, callback) {
        var listners = this._get_map(name);
        listners.push(callback);
    };

    EventDispatcher.prototype.remove = function(name, callback) {
        var listners = this._get_map(name);
        var index = listners.indexOf(callback);
        if(index != -1) {
            listners.splice(index, 1);
        }
    };

    EventDispatcher.prototype.dispatch = function(name)
    {
        var listners = this._get_map(name);
        Array.prototype.splice.call(arguments, 0, 1);
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

