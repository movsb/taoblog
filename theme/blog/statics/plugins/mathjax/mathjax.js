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

	// jQuery adds _ parameter to skip cache
	let s = document.createElement('script');
	s.src= 'https://cdnjs.cloudflare.com/ajax/libs/mathjax/2.7.4/MathJax.js';
	s.async = true;
	document.body.appendChild(s);

    let mathjaxTimer;
    mathjaxTimer = setInterval(function(){
        if(!window.MathJax || !MathJax.isReady) return;
        clearInterval(mathjaxTimer);

		console.log('mathjax loaded');
		
		let p = document.createElement('p');
		p.innerText = '$a$';

        MathJax.Hub.Typeset(p, function(){
			console.log('typeset warming-up');

			// for old-styled maths
            document.querySelectorAll('code:not([class*="lang"])').forEach(function(e) {
				let html = e.innerHTML;
                if(html.startsWith('$') && html.endsWith('$')) {
					let parent = document.createElement('div');
					parent.innerHTML = html.startsWith('$$') ? '<div></div>' : '<span></span>';
					let child = parent.firstElementChild;
					child.style.margin = '3px';
					child.innerHTML = html;
                    MathJax.Hub.Typeset(child);
                    // console.log('Typeset: ', e);
					e.replaceWith(child);
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
