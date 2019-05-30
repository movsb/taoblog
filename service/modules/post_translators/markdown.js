var fs = require('fs');
var marked = require('marked');
var sizeOf = require('image-size');

var renderer = new marked.Renderer();
renderer.image = function(href, alt, text) {
    function old() {
        return marked.Renderer.prototype.image.call(renderer, href, alt, text);
    }

    if(!/^[^/]+$/.test(href)) {
        return old();
    }

    href = escape(href);
    alt = alt || "";
    var d = {};
    try {
        var dir = process.argv[2];
        var file = dir + '/' + href;
        d  = sizeOf(file);
    } catch(err) {
        return old();
    }
    return `<img data-src="${href}" alt="${alt}" width="${d.width}px" height="${d.height}px" />`;
};

marked.setOptions({
    renderer: renderer,
    langPrefix: 'language-'
});

var data = fs.readFileSync(0, 'utf-8');
var html = marked(data);
process.stdout.write(html);
