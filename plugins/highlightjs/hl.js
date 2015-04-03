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
});

