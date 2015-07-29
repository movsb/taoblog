/* 提示框 */
var tips_tmr = null;

function show_tips(so) {
	var tips = $('#tips');
	if(tips.length == 0) {
		$('body').append('<div class="tips" id="tips"></div>');
		tips = $('#tips');
	}

	function dismiss() {
		if(tips_tmr) {
			clearTimeout(tips_tmr);
			tips_tmr = null;
		}

		tips.fadeOut(300);
	}

	if(tips_tmr) {
		clearTimeout(tips_tmr);
		tips_tmr = null;
	}

	if(typeof so == 'string' || typeof so == 'number') {
		tips.html(so);
		tips.css('background-color', 'rgba(255,100,100,0.93');
		tips.fadeIn(300);

		tips_tmr = setTimeout(function() {
			tips.fadeOut(300);
			tips_tmr = null;
		}, 3000);

		return;
	} else if(typeof so == 'object') {
		tips.css('background-color', so.backgroundColor ?
			so.backgroundColor : 'rgba(255,100,100,0.93)');

		var timeout = so.timeout ? so.timeout : 3000;
		var content = so.content ? so.content : '';

		tips.html(content);
		tips.fadeIn(300);

		if(so.click) {
			tips.click(function(e){
				so.click({
					event: e,
					dismiss: dismiss
				});
			});
		}

		tips_tmr = setTimeout(dismiss, timeout);

		return;
	}
}

/* 读取字体大小 */
if(localStorage.getItem('font-size')) {
	document.write('<style>.entry { font-size: ' + localStorage.getItem('font-size') + '; }</style>');
}


