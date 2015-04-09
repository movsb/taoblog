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

