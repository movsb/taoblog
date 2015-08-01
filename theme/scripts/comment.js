document.write(function(){/*
	<!--评论标题 -->
	<h3 id="comment-title">
		<i class="fa fa-mr fa-comments"></i>评论<span class="no-sel">(<span class="loaded">0</span>/<span class="total">0</span>)</span>
	</h3>

	<!-- 评论列表  -->
	<ol id="comment-list">

	</ol>

	<!-- 评论功能区  -->
	<div class="comment-func">
		<span id="post-comment" class="post-comment no-sel btn">发表评论</span>
		<span id="load-comments" class="load-comments no-sel btn">加载评论</span>
		&nbsp;<span id="loading-status"></span>
		<input type="hidden" id="post-id" name="post-id" value="" />
	</div>

	<!-- 评论框 -->
	<div id="comment-form-div" class="normal">
		<div class="comment-form-div-1">
			<div class="no-sel">
				<div class="closebtn" title="关闭">
					<i class="fa fa-times"></i>
				</div>
				<div class="maxbtn" title="最大化">
					<i class="fa fa-plus"></i>
				</div>

				<div class="comment-title">
					<span>评论</span>
				</div>
			</div>

			<form id="comment-form" action="/admin/comment.php">
				<div class="fields">
					<div class="field">
						<label>昵称</label>
						<input type="text" name="author" required/>
						<span class="needed">必填</span>
					</div>
					<div class="field">
						<label>邮箱</label>
						<input type="text" name="email" required/>
						<span class="needed">必填(不公开)</span>
					</div>
					<div class="field">
						<label>网址</label>
						<input type="text" name="url" />
					</div>
					<div style="display: none;">
						<input id="comment-form-post-id" type="hidden" name="post_id" value="" />
						<input id="comment-form-parent"  type="hidden" name="parent" value="" />
						<input type="hidden" name="do" value="post-cmt" />
					</div>
				</div>

				<div class="comment-content">
					<label style="position: absolute;">评论</label>
					<label style="visibility: hidden;">评论</label> <!-- 别问我，我想静静 -->
					<textarea id="comment-content" name="content" wrap="off" required></textarea>
				</div>

				<div class="comment-submit">
					<input type="submit" id="comment-submit" class="no-sel" value="发表评论" />
					&nbsp;<span id="submitting-status"><span>
				</div>
			</form>
		</div>
		<div class="comment-form-div-2">
			<div class="toolbar no-sel" style="color: #C8C8C8; font-size: 24px;">
				<span>评论: </span>
				<div class="right">
					<span>字体: </span>
					<span class="font-dec" title="减小字号"><i class="fa fa-minus"></i></span>
					<span class="font-inc" title="增大字号"><i class="fa fa-plus"></i></span>&nbsp;
					<span class="close" title="还原"><i class="fa fa-times"></i></span>
				</div>
			</div>
			<div class="textarea-wrapper">
				<textarea id="comment-content-2" wrap="off"></textarea>
			</div>
		</div>
	</div>
*/}.toString().slice(14,-3));

$('#post-id').val(_post_id);

// 加载评论总数
$.post(
	'/admin/comment.php',
	'do=get-count&post_id=' + $('#post-id').val(),
	function(data) {
		$('#comment-title .total').text(data);
	}
);

// 无公害评论内容
function sanitize_content(c) {
    c = c.replace(/&/g, '&amp;');
    c = c.replace(/</g, '&lt;');
    c = c.replace(/>/g, '&gt;');
    c = c.replace(/'/g, '&#39;');
    c = c.replace(/"/g, '&#34;');

    return c;
}

function comment_avatar(eh,sz) {
	return '/theme/avatar.php?' + encodeURIComponent(eh + '?s=' + sz);
}

function comment_friendly_date(d) {
	var year = d.substring(0,4);
	var start = year == (new Date()).getFullYear() ? 5 : 0;
	return d.substring(start, 16);
}

// 从评论生成html内容
function comment_item(cmt) {
	var s = '';
	s += '<li style="display: none;" class="comment-li" id="comment-' + cmt.id + '">\n';
	s += '<div class="comment-avatar">'
		+ '<img src="' + comment_avatar(cmt.avatar, 48) + '" width="48px" height="48px"/>'
		+ '</div>\n';
	s += '<div class="comment-meta">\n';

	if(cmt.is_admin) {
		s += '<span class="author">楼主 </span>';
	} else {
		var nickname;
		if(typeof cmt.url == 'string' && cmt.url.length) {
			if(!cmt.url.match(/^https?:\/\//i))
				cmt.url = 'http://' + cmt.url;

			nickname = '<a rel="nofollow" target="_blank" href="' + cmt.url + '">' + cmt.author + '</a>'
		} else {
			nickname = cmt.author;
		}

		s += '<span class="nickname">' + nickname + '</span>\n';
	}

	s += '<span class="date">' + comment_friendly_date(cmt.date) + '</span>\n</div>\n';
	s += '<div class="comment-content">' + sanitize_content(cmt.content) + '</div>\n';
	s += '<div class="reply-to no-sel" style="margin-left: 54px;"><a style="cursor: pointer;" onclick="comment_reply_to('+cmt.id+');return false;">回复</a></div>';
	s += '</li>';

	return s;
}

// 弹出回复框
function comment_reply_to(p){
	$('#comment-form-post-id').val($('#post-id').val());
	$('#comment-form-parent').val(p);

	// 设置已保存的作者/邮箱/网址,其实只需要在页面加载完成后设置一次即可，是嘛？
	if(window.localStorage) {
		$('#comment-form input[name=author]').val(localStorage.getItem('cmt_author'));
		$('#comment-form input[name=email]').val(localStorage.getItem('cmt_email'));
		$('#comment-form input[name=url]').val(localStorage.getItem('cmt_url'));
	}

	$('#comment-form-div').fadeIn();
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
$('#load-comments').click(function() {
	if($(this).attr('loading') === 'true')
		return;

	var load = this;
	$(this).attr('loading', 'true');
	$('#loading-status').html('<i class="fa fa-spin fa-spinner"></i><span>加载中...</span>');
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

			$('#loading-status').text('');

			if(cmts.length == 0) {
				$('#loading-status').html('<i class="fa fa-info-circle"></i><span>没有了！</span>');
				setTimeout(function(){
						$('#loading-status').text('');
					},
					1500
				);
			} else {
				cmt_loaded += cmts.length;
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

$('#load-comments').click();

function cmt_save_info() {
	if(window.localStorage) {
		localStorage.setItem('cmt_author', $('#comment-form input[name=author]').val());
		localStorage.setItem('cmt_email', $('#comment-form input[name=email]').val());
		localStorage.setItem('cmt_url', $('#comment-form input[name=url]').val());
	} else {
		if(typeof show_tips == 'function') {
			show_tips('抱歉，你的浏览器不支持 localStorage，评论者的相关信息将无法正确地保存以便后续使用。');
		}
	}
}

// Ajax评论提交
$('#comment-submit').click(function() {
	var timeout = 1500;

	$(this).attr('disabled', 'disabled');
	$('#submitting-status').html('<i class="fa fa-spin fa-spinner"></i>正在提交...');
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
				$('#submitting-status').html('<i class="fa fa-mr fa-info-circle"></i>评论成功！');
				setTimeout(function() {
						$('#comment-form-div').fadeOut();
						$('#submitting-status').text('');
						cmt_save_info();
					},
					timeout
				);
			} else {
				$('#submitting-status').html('<i class="fa fa-mr fa-info-circle"></i>' + data.error);
				setTimeout(function() {
						$('#submitting-status').text('');
					},
					timeout
				);
			}
		},
		'json'
	)
	.fail(function(xhr, sta, e){
		$('#submitting-status').text('未知错误！');
	})
	.always(setTimeout(function(){
			$('#submitting-status').text('');
			$('#comment-submit').removeAttr('disabled');
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
		ta.css('font-size', Math.max(8,parseFloat(ta.css('font-size'))/1.2) + 'px');
	});

	$('#comment-form-div.maximized .toolbar .close').click(function(){
		$('#comment-form-div').removeClass('maximized').attr('style','').addClass('normal');
		$('#comment-content').val($('#comment-content-2').val());
		$('#comment-form-div.normal').show();
	});

});

