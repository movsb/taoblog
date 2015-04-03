$.fn.center = function () {
    this.css("position","fixed");
    this.css("top", Math.max(0, ($(window).height() - $(this).outerHeight()) / 2) + 'px');
	this.css("left", Math.max(0, ($(window).width() - $(this).outerWidth()) / 2) + 'px');
	return this;
}

// 加载评论总数
$.post(
	'/admin/comment.php',
	'do=get-count&post_id=' + $('#post-id').val(),
	function(data) {
		$('#comment-title .total').text(data);
	}
);

// 评论框
$('#comment-form-div').keyup(function(e){
	if(e.keyCode==27) {
		if($('#comment-form-div .comment-form-div-data input[name="maximized"]').val() === 'true'){
			$('#comment-content').val($('#comment-content-2').val());
			$('#comment-content-2').hide();
			$('#comment-form-div .comment-form-div-1').show();
			$('#comment-form-div .comment-form-div-data input[name="maximized"]').val('false');
		} else {
			$(this).fadeOut();
		}
	}
});
// 无公害评论内容
function sanitize_content(c) {
    c = c.replace(/&/g, '&amp;');
    c = c.replace(/</g, '&lt;');
    c = c.replace(/>/g, '&gt;');
    c = c.replace(/'/g, '&#39;');
    c = c.replace(/"/g, '&#34;');

    return c;
}

// 从评论生成html内容
function comment_item(cmt) {
	var s = '';
	s += '<li style="display: none;" class="comment-li" id="comment-' + cmt.id + '">\n';
	s += '<div class="comment-avatar" onclick="location.hash=\'#comment-'+cmt.id + '\';">'
		+ '<img src="' + cmt.avatar + '" width="48px" height="48px"/>'
		+ '</div>\n';
	s += '<div class="comment-meta">\n';

	if(cmt.author == 'anhbk@qq.com') {
		s += '<span style="color: red;" class="author">楼主 </span>';
	} else {
		s += '<span class="nickname">' + cmt.author + '</span>\n';
	}

	s += '<p class="date">' + cmt.date + '</p>\n</div>\n';
	s += '<div class="comment-content">' + sanitize_content(cmt.content) + '</div>\n';
	s += '<div class="reply-to" style="margin-left: 54px;"><a style="cursor: pointer;" onclick="comment_reply_to('+cmt.id+');return false;">回复</a></div>';
	s += '</li>';

	return s;
}

// 弹出回复框
function comment_reply_to(p){
	$('#comment-form-post-id').val($('#post-id').val());
	$('#comment-form-parent').val(p);
	$('#comment-form-div').center().fadeIn();
}

// 为上一级评论添加div
function comment_add_reply_div(id){
	if($('#comment-'+id+' .comment-replies').length === 0){
		$('#comment-'+id).append('<div class="comment-replies" id="comment-reply-'+id+'"><ol></ol></div>');
	}
}

// 发表评论按钮
$('#post-comment').click(function(){
	comment_reply_to(0);
});

// 保存 加载进度/加载总数
var cmt_loaded = 0;
var cmt_loaded_ch = 0;

function comment_append_children(ch, p) {
	for(var i=0; i<ch.length; i++){
		if(!ch[i]) continue;

		if(ch[i].parent == p) {
			$('#comment-reply-'+p + ' ol:first').append(comment_item(ch[i]));
			$('#comment-'+ch[i].id).fadeIn();
			comment_add_reply_div(ch[i].id);
			comment_append_children(ch, ch[i].id);
			delete ch[i];
		}
	}
}

// 加载按钮
$('#load-comments .load').click(function() {
	if($(this).attr('loading') === 'true')
		return;

	var load = this;
	$(this).attr('loading', 'true');
	$('#load-comments .loading').show();
	$.post(
		'/admin/comment.php',
		'do=get-cmt&count=5&offset=' + cmt_loaded
			+ '&post_id=' + $('#post-id').val(),
		function(data) {
			var cmts = data.cmts || [];
			var ch_count = 0;
			for(var i=0; i<cmts.length; i++){
				$('#comment-list').append(comment_item(cmts[i]));
				$('#comment-'+cmts[i].id).fadeIn();
				comment_add_reply_div(cmts[i].id);

				comment_append_children(cmts[i].children, cmts[i].id);

				ch_count += cmts[i].children.length;
			}

			$('#load-comments .loading').hide();

			if(cmts.length == 0) {
				$('#load-comments .none').show();
				setTimeout(function(){
						$('#load-comments .none').hide();
					},
					1500
				);
			} else {
				cmt_loaded += cmts.length;
				location.hash = '#comment-' + cmts[0].id;
			}
			$(load).removeAttr('loading');

			cmt_loaded_ch += ch_count;
			$('#comment-title .loaded').text(cmt_loaded + cmt_loaded_ch);
		},
		'json'
	)
	.always(setTimeout(function(){
			$(load).removeAttr('loading');
	},1500));
});

// Ajax评论提交
$('#comment-submit').click(function() {
	var timeout = 1500;

	$(this).attr('disabled', 'disabled');
	$('#comment-form .comment-submit .submitting').show();
	$.post(
		$('#comment-form')[0].action,
		$('#comment-form').serialize()+'&return_cmt=1',
		function(data){
			if(data.errno == 'success') {
				var parent = $('#comment-form input[name="parent"]').val();
				if(parent == 0) {
					$('#comment-list').append(comment_item(data.cmt));
					// 没有父评论，避免二次加载。
					cmt_loaded ++;

				} else {
					comment_add_reply_div(parent);
					$('#comment-reply-'+parent + ' ol:first').append(comment_item(data.cmt));
				}
				$('#comment-'+data.cmt.id).fadeIn();
				$('#comment-content').val('');
				$('#comment-form .comment-submit .succeeded span').text('评论成功！');
				$('#comment-form .comment-submit .succeeded').show();
				setTimeout(function() {
						$('#comment-form-div').fadeOut();
						$('#comment-form .comment-submit .succeeded').hide();
					},
					timeout
				);
			} else {
				console.log(data.error);
				$('#comment-form .comment-submit .failed span').text(data.error);
				$('#comment-form .comment-submit .failed').show();
				setTimeout(function() {
						$('.form-submit .failed').hide();
					},
					timeout
				);
			}

			$('#comment-form .comment-submit .submitting').hide();
		},
		'json'
	)
	.fail(function(xhr, sta, e){
		$('#comment-form .comment-submit .submitting').hide();
		var info = $('#comment-form .comment-submit .failed span');
		if(xhr.status == '409'){
			info.text('请不要过快地提交，或提交相同的评论！');
		} else {
			info.text('未知错误！');
		}
		$('#comment-form .comment-submit .failed').show();
	})
	.always(setTimeout(function(){
			$('#comment-form .comment-submit .submitting').hide();
			$('#comment-submit').removeAttr('disabled');
			$('#comment-form .comment-submit .failed').hide();
		},
		timeout
	));
	return false;
});

// 评论输入框允许TAB键
function enableTabIndent(t,e){
	if(e.keyCode === 9){
		var start = t.selectionStart;
		var end = t.selectionEnd;

		var that = $(t);

		var value = that.val();
		var before = value.substring(0, start);
		var after = value.substring(end);
		var selArray = value.substring(start, end).split('\n');

		var isIndent = !e.shiftKey;

		if(isIndent){
			if(start === end || selArray.length === 1){
				that.val(before + '\t' + after);
				t.selectionStart = t.selectionEnd = start + 1;
			} else {
				var sel = '\t' + selArray.join('\n\t');
				that.val(before + sel + after);
				t.selectionStart = start + 1;
				t.selectionEnd = end + selArray.length; 
			}
		} else {
			var reduceEnd = 0;
			var reduceStart = false;

			if(selArray.length > 1) {
				selArray.forEach(function(x, i, a){
					if(i>0 && x.length>0 &&  x[0]==='\t'){
						a[i] = x.substring(1);
						reduceEnd++;
					}
				});
				sel = selArray.join('\n');
			} else {
				sel = selArray[0];
			}

			var b1 = '',b2 = '';
			if(before.length){
				var npos = before.lastIndexOf('\n');
				if(npos !== -1){
					b1 = before.substring(0, npos+1);
					b2 = before.substring(npos+1);
				} else {
					b1 = '';
					b2 = before;
				}

				sel = b2 + sel;
			}

			if(sel.length && sel[0]==='\t'){
				sel = sel.substring(1);
				reduceStart = true;
			}

			that.val(b1 + sel + after);
			t.selectionStart = start + (reduceStart ? -1 : 0);
			t.selectionEnd = end - (reduceEnd + (reduceStart ? 1 : 0));
		}
		return true;
	}
	return false;
}

$('#comment-content').keydown(function(e){
	if(enableTabIndent(this, e)){
		e.preventDefault();
	}
});

$('#comment-content-2').keydown(function(e){
	if(enableTabIndent(this, e)){
		e.preventDefault();
	}
});

if(/comment-/.test(location.hash)) {
	alert('请点击“加载评论”来查看评论！');
}



$('#comment-form-div .closebtn').click(function(){
		$('#comment-form-div').fadeOut();
});

$('#comment-form-div .maxbtn').click(function(){
	$('#comment-content-2').val($('#comment-content').val());
	$('#comment-form-div').removeClass('normal').attr('style','').addClass('maximized');
	$('#comment-form-div.maximized .toolbar .font-inc').click(function(){
		var ta = $('#comment-content-2');
		ta.css('font-size', parseFloat(ta.css('font-size'))*1.2 + 'px');
	});

	$('#comment-form-div.maximized .toolbar .font-dec').click(function(){
		var ta = $('#comment-content-2');
		ta.css('font-size', Math.max(0,parseFloat(ta.css('font-size'))/1.2) + 'px');
	});

	$('#comment-form-div.maximized .toolbar .close').click(function(){
		$('#comment-form-div').removeClass('maximized').attr('style','').addClass('normal');
		$('#comment-content').val($('#comment-content-2').val());
		$('#comment-form-div.normal').center().show();
	});

});

