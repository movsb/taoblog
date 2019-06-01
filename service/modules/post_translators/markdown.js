var fs = require('fs');
var marked = require('marked');
var sizeOf = require('image-size');

function encodeAttr(s) {
    // TODO encode
    return s;
}

var renderer = new marked.Renderer();
renderer.image = function(href, alt, text) {
    function old() {
        return marked.Renderer.prototype.image.call(renderer, href, alt, text);
    }

    if(!/^[^/]+$/.test(href)) {
        return old();
    }

    var d = {};
    try {
        var dir = process.argv[2];
        var file = dir + '/' + href;
        d  = sizeOf(file);
    } catch(err) {
        // console.warn(err);
    }

    href = encodeAttr(href);
    alt = encodeAttr(alt || href);

    if(d.width && d.height) {
        return `<img data-src="${href}" alt="${alt}" width="${d.width}px" height="${d.height}px" />`;
    } else {
        return `<img data-src="${href}" alt="${alt}" />`;
    }
};

marked.setOptions({
    renderer: renderer,
    langPrefix: 'language-'
});

var data = fs.readFileSync(0, 'utf8');
var html = marked(data);
process.stdout.write(html);
