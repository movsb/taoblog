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

// Lazy load images
(function() {
    function setSrc(img) {
        let src = img.getAttribute('data-src');
        img.setAttribute('src', src);
        img.removeAttribute('data-src');
        img.removeAttribute('width');
        img.removeAttribute('height');
        return src;
    }

    let supportIntersectionObserver = !window.IntersectionObserver !== undefined;
    let allDataSrcImgs = document.querySelectorAll('article img[data-src]');

    if(!supportIntersectionObserver) {
        console.warn("browser doesn't support IntersectionObserver");
        allDataSrcImgs.forEach(img => {
            setSrc(img);
        })
        return;
    }

    function onIntersection(entries, observer) {
        entries.forEach(entry => {
            if(entry.intersectionRatio != 0) {
                let img = entry.target;
                let src = setSrc(img);
                observer.unobserve(img);
                console.log("Lazy loading", src);
            }
        });
    }

    let observer = new IntersectionObserver(onIntersection);
    allDataSrcImgs.forEach(img => { observer.observe(img); })
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
