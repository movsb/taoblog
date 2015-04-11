if($('div.code').length) {
	$('head').append('<link rel="stylesheet" type="text/css" href="/plugins/highlight/monokai_sublime.css" />');
	// jQuery will cause the underscore(_) query parameter added to the request
	//$('head').append('<script type="text/javascript" src="/plugins/highlight/highlight.min.js"></script>');
	//$.getScript('/plugins/highlight/highlight.min.js');
	(function(){
		var head = document.getElementsByTagName('head')[0];
		var s = document.createElement('script');
		s.src = '/plugins/highlight/highlight.min.js';
		head.appendChild(s);
	})();


	var hljscnt = 0;
	var hljstmr = setInterval(function() {
		if(typeof hljs == 'undefined') {
			if(++hljscnt >= 100) clearInterval(hljstmr);

			return;
		}

		clearInterval(hljstmr);

		$('div.code').each(function(i, e){
			var that = $(this);
			var ho = {};

			// TAB 换空格
			if(that.attr('tabsize')) {
				ho.tabReplace = (new Array(parseInt(that.attr('tabsize'))+1)).join(' ');
			} else {
				ho.tabReplace = '    ';
			}

			// 程序语言
			if(that.attr('lang')) {
				ho.languages = [that.attr('lang')];
			}

			hljs.configure(ho);

			hljs.highlightBlock(e);

			// 行号
			var lines = (e.innerHTML.match(/\n/g) || []).length;
			if(!e.innerHTML.match(/\n$/)) lines += 1;
			var s = '<div class="code-wrap"><table><tr><td class="gutter">';
			for(var i=1; i<lines; i++)
				s += '<span>' + i + '</span><br />';
			s += '<span>' + lines + '</span>';
			s += '</td><td class="code">' + e.outerHTML  + '</td></tr></table></div>';

			$(e).replaceWith(s);
		});
	},
	100);
}


if($('pre.shell').length) {
	$('head').append('<link rel="stylesheet" href="/plugins/highlight/shell.css" />');
}

