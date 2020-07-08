(function() {
	window.MathJax = {
		jax: ["input/TeX", "output/HTML-CSS"],
		extensions: ["tex2jax.js"],
		tex2jax: {
			skipTags: [],
			inlineMath: [['$', '$'], ['\\(', '\\)']],
			displayMath: [['$$', '$$'], ['\\[', '\\]']]
		},
		skipStartupTypeset: true,
		menuSettings: {
			zoom: 'Double-Click'
		}
	};

	var $body = $('body');
    // jQuery adds _ parameter to skip cache
    var s = document.createElement('script');
    s.src= 'https://cdnjs.cloudflare.com/ajax/libs/mathjax/2.7.4/MathJax.js';
    s.async = true;
    $body[0].appendChild(s);

    var mathjaxTimer;
    mathjaxTimer = setInterval(function(){
        if(!window.MathJax || !MathJax.isReady) return;
        clearInterval(mathjaxTimer);

		console.log('mathjax loaded');

        MathJax.Hub.Typeset($('<p>$a$</p>').get(0), function(){
			console.log('typeset warming-up');

			// for old-styled maths
            $('code:not([class*="lang"])').each(function(_, e) {
                var html = $(e).html();
                if(html.startsWith('$') && html.endsWith('$')) {
                    var wrap = $(html.startsWith('$$') ? '<div/>' : '<span/>')
                        .css('margin', '3px')
                        .html(html)[0];
                    MathJax.Hub.Typeset(wrap);
                    // console.log('Typeset: ', e);
                    $(e).replaceWith(wrap);
                }
			});

			// for new markdown-generated maths.
			var maths = document.querySelectorAll('.math.inline,.math.display');
			maths.forEach(function(elem) {
				MathJax.Hub.Typeset(elem);
			});
        });
    }, 1000);
})();
