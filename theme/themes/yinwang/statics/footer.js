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
    var s = document.createElement('script');
    s.src = '/plugins/mathjax/mathjax.js';
    s.async = true;
    document.body.appendChild(s);
}
})();
