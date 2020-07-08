(function() {
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
