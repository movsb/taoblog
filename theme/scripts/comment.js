document.write(function(){/*
	<!--评论标题 -->
	<h3 id="comment-title">
		评论<span class="no-sel count-wrap"> <span class="loaded">0</span>/<span class="total" itemprop="commentCount">0</span></span>
	</h3>

	<!-- 评论列表  -->
	<ol id="comment-list">

	</ol>

	<!-- 评论功能区  -->
	<div class="comment-func">
		<button id="post-comment" class="post-comment no-sel btn">发表评论</button>
		<button id="load-comments" class="load-comments no-sel btn">加载评论</button>
		&nbsp;<span id="loading-status"></span>
		<input type="hidden" id="post-id" name="post-id" value="" />
	</div>

	<!-- 评论框 -->
	<div id="comment-form-div" class="normal">
		<div class="comment-form-div-1 no-sel">
			<div class="no-sel" class="nc">
                <div class="ncbtns">
                    <img src="/theme/images/close.svg" width="20" height="20" title="隐藏" class="closebtn"/>
                    <img src="/theme/images/question.svg" width="20" height="20" title="支持部分QQ表情，如：

    [笑哭] [小纠结] [无奈]

支持插入代码片段，请使用3个反引号括起来，如：

    ```
    #include <stdio.h>

    int main()
    {
        printf(&quot;Hello, world!&quot;);
        return 0;
    }
    ```
"/>
                </div>
				<div class="comment-title">
					<span>评论</span>
				</div>
			</div>

			<form id="comment-form" action="/admin/comment.php" method="post">
                <div style="display: none;">
                    <input id="comment-form-post-id" type="hidden" name="post_id" value="" />
                    <input id="comment-form-parent"  type="hidden" name="parent" value="" />
                    <input type="hidden" name="do" value="post-cmt" />
                </div>

                <textarea id="comment-content" name="content" wrap="off"></textarea>

                <div class="fields">
                    <input type="text" name="author" placeholder="昵称" />
                    <input type="text" name="email" placeholder="邮箱（不公开）"/>
                    <input type="text" name="url" placeholder="个人站点" />
					<input type="submit" id="comment-submit" class="no-sel" value="发表评论" />
                </div>
			</form>
		</div>
	</div>
*/}.toString().slice(14,-3));

$('#post-id').val(_post_id);

function Comment() {
    this._count  =0;
    this._loaded = 0;       // 已加载评论数
    this._loaded_ch = 0;
}

Comment.prototype.init = function() {
    var self = this;

    // 发表评论按钮
    $('#post-comment').click(function(){
        self.reply_to(0);
    });

    // 加载按钮
    $('#load-comments').click(function() {
        if($(this).attr('loading') === 'true')
            return;

        var load = this;
        $(this).attr('loading', 'true');
        $('#loading-status').html('<i class="fa fa-spin fa-spinner"></i><span>加载中...</span>');

        var pid = $('#post-id').val();

        $.get('/v1/posts/' + pid + '/comments',
            {
                order: 'desc',
                count: 10,
                offset: self._loaded,
            },
            function(cmts) {
                var ch_count = 0;
                for(var i=0; i<cmts.length; i++){
                    $('#comment-list').append(self.gen_comment_item(cmts[i]));
                    $('#comment-'+cmts[i].id).fadeIn();
                    TaoBlog.events.dispatch('comment', 'post', $('#comment-'+cmts[i].id));
                    self.add_reply_div(cmts[i].id);

                    self.append_children(cmts[i].children, cmts[i].id);

                    ch_count += cmts[i].children.length;
                }

                if(typeof(jQuery)=='function' && typeof(jQuery.timeago)=='function')
                    jQuery('.comment-meta .date').timeago();

                $('#loading-status').text('');

                if(cmts.length == 0) {
                    $('#loading-status').html('<i class="fa fa-info-circle"></i><span>没有了！</span>');
                    setTimeout(function(){
                            $('#loading-status').text('');
                        },
                        1500
                    );
                } else {
                    self._loaded += cmts.length;
                }
                $(load).removeAttr('loading');

                self._loaded_ch += ch_count;
                $('#comment-title .loaded').text(self._loaded + self._loaded_ch);
            },
            'json'
        ).fail(function(x) {
            alert(x.responseText);
        })
        .always(setTimeout(function(){
            $(load).removeAttr('loading');
        },1500));
    });

    // Ajax评论提交
    $('#comment-submit').click(function() {
        var timeout = 1500;

        $(this).attr('disabled', 'disabled');
        $('#comment-submit').val('提交中...');
        $.post(
            $('#comment-form')[0].action,
            $('#comment-form').serialize()+'&return_cmt=1',
            function(data){
                if(data.errno == 'success') {
                    var parent = $('#comment-form input[name="parent"]').val();
                    if(parent == 0) {
                        $('#comment-list').prepend(self.gen_comment_item(data.cmt));
                        // 没有父评论，避免二次加载。
                        self._loaded ++;

                    } else {
                        self.add_reply_div(parent);
                        $('#comment-reply-'+parent + ' ol:first').append(self.gen_comment_item(data.cmt));
                    }
                    $('#comment-'+data.cmt.id).fadeIn();
                    TaoBlog.events.dispatch('comment', 'post', $('#comment-'+data.cmt.id));
                    $('#comment-content').val('');
                    $('#comment-form-div').fadeOut();
                    self.save_info();
                } else {
                    alert(data.error);
                }
            },
            'json'
        )
        .fail(
            function(xhr, sta, e){
                alert('未知错误！');
            }
        )
        .always(
            function(){
                $('#comment-submit').val('发表评论');
                $('#comment-submit').removeAttr('disabled');
            }
        );

        return false;
    });

    $('#comment-form-div .closebtn').click(function(){
            $('#comment-form-div').fadeOut();
    });

    // hide comment box when ESC key pressed
    window.addEventListener('keyup', function(e) {
        if(e.keyCode == 27) {
            $('#comment-form-div').fadeOut();
        }
    });

    $(window).on('scroll', function() {
        self.load_essential_comments();
    });

    $(window).on('load', function() {
        self.get_count(function() {
            self.load_essential_comments();
        });
    });
};

Comment.prototype.load_essential_comments = function() {
    if(this._loaded + this._loaded_ch < this._count
        && window.scrollY + window.innerHeight + 50 >= document.body.scrollHeight) 
    {
        this.load_comments();
    }
};

Comment.prototype.get_count = function(callback) {
    var self = this;
    var pid = $('#post-id').val();
    $.get('/v1/posts/' + pid + '/comments:count',
        function(data) {
            self._count = data;
            $('#comment-title .total').text(self._count);
            callback();
        },
        'json'
    );
};

Comment.prototype.gen_avatar = function(eh, sz) {
	return '/theme/avatar.php?' + encodeURIComponent(eh + '?d=mm&s=' + sz);
};

Comment.prototype.emotions = ["狗头", "偷笑", "冷汗", "卖萌", "可爱", "呲牙", "喷血", "嘘", "坏笑", "小纠结", "尴尬", "幽灵", "微笑", "惊喜", "惊恐", "惊讶", "憨笑", "我最美", "托腮", "抠鼻", "拥抱", "撇嘴", "擦汗", "敲打", "斜眼笑", "无奈", "晕", "泪奔", "流汗", "流泪", "玫瑰", "疑问", "笑哭", "衰", "调皮", "阴险", "难过", "骚扰"];

Comment.prototype.normalize_content = function(c) {
    var s = this.h2t(c);
    s = s.replace(/```(\s*(\w+)\s*)?\r?\n([\s\S]+?)```/mg, '<pre class="code"><code class="language-$2">$3</code></pre>');
    s = s.replace(/\[([^\x20-\x7E]{1,3})\]/gm, function(all,alt) {
            if(Comment.prototype.emotions.indexOf(alt) != -1)
                return $('<img/>')
                    .attr('alt', all)
                    .attr('width', '20px')
                    .attr('height', '20px')
                    .css('vertical-align', 'bottom')
                    .attr('src', '/theme/emotions/' + alt + '.png')[0]
                    .outerHTML;
            else
                return all;
        });
    return s;
};

Comment.prototype.friendly_date = function(d) {
	var year = d.substring(0,4);
	var start = year == (new Date()).getFullYear() ? 5 : 0;
	return d.substring(start, 16);
};

// https://stackoverflow.com/a/12034334/3628322
// escape html to text
Comment.prototype.h2t = function(h) {
    var map = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
    };

    return h.replace(/[&<>]/g, function (s) {
        return map[s];
    });
};

// escape html to attribute
Comment.prototype.h2a = function(h) {
    var map = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      "'": '&#39;',
      '"': '&quot;'
    };

    return h.replace(/[&<>'"]/g, function (s) {
        return map[s];
    });
}

Comment.prototype.gen_comment_item = function(cmt) {
	var s = '';

    var loggedin = cmt.ip != undefined;

    // 登录后可以显示评论者的详细信息
    var info = '';
    if(loggedin) {
        // author, email, url, ip, date
        info = '作者：' + this.h2a(cmt.author)
            + '\n邮箱：' + this.h2a(cmt.email)
            + '\n网址：' + this.h2a(cmt.url)
            + '\n地址：' + cmt.ip
            + '\n日期：' + cmt.date
            ;
    }

	s += '<li style="display: none;" class="comment-li" id="comment-' + cmt.id + '" itemprop="comment">\n';
	s += '<div class="comment-avatar">'
		+ '<img src="' + this.gen_avatar(cmt.avatar, 48) + '" width="48px" height="48px" title="'+info+'"/>'
		+ '</div>\n';
	s += '<div class="comment-meta">\n';

	if(cmt.is_admin) {
		s += '<span class="author">【作者】' + this.h2t(cmt.author) + '</span>\n';
	} else {
		var nickname;
		if(typeof cmt.url == 'string' && cmt.url.length) {
			if(!cmt.url.match(/^https?:\/\//i))
				cmt.url = 'http://' + cmt.url;

			nickname = '<i class="fa fa-home"></i>';
			nickname += '<a rel="nofollow" target="_blank" href="' + this.h2a(cmt.url) + '">' + this.h2t(cmt.author) + '</a>';
		} else {
			nickname = this.h2t(cmt.author);
		}

		s += '<span class="nickname">' + nickname + '</span>\n';
	}

    s += '<time class="date" datetime="' + cmt.date + '">' + this.friendly_date(cmt.date) + '</time>\n</div>\n';
	s += '<div class="comment-content">' + this.normalize_content(cmt.content) + '</div>\n';
    s += '<div class="toolbar no-sel" style="margin-left: 54px;">';
	s += '<a onclick="comment.reply_to('+cmt.id+');return false;">回复</a>';
	if(loggedin) {
		s += '<a onclick="confirm(\'确定要删除？\') && comment.delete_me('+cmt.id+');return false;">删除</a>';
	}
    s += '</div>';
	s += '</li>';

    // console.log(s);

	return s;
};

Comment.prototype.reply_to = function(p){
	$('#comment-form-post-id').val($('#post-id').val());
	$('#comment-form-parent').val(p);

	// 设置已保存的作者/邮箱/网址,其实只需要在页面加载完成后设置一次即可，是嘛？
	if(window.localStorage) {
		$('#comment-form input[name=author]').val(localStorage.getItem('cmt_author'));
		$('#comment-form input[name=email]').val(localStorage.getItem('cmt_email'));
		$('#comment-form input[name=url]').val(localStorage.getItem('cmt_url'));
	}

	$('#comment-form-div').fadeIn();
    $('#comment-content').focus();
};

Comment.prototype.delete_me = function(p) {
    var pid = $('#post-id').val();
	$.ajax({
        url: '/v1/posts/' + pid + '/comments/' + p,
        type: 'DELETE',
        success: function() {
            $('#comment-'+p).remove();
		},
        error: function(){
            alert('删除失败。');
        }
    });
};

// 为上一级评论添加div
Comment.prototype.add_reply_div = function(id){
	if($('#comment-'+id+' .comment-replies').length === 0){
		$('#comment-'+id).append('<div class="comment-replies" id="comment-reply-'+id+'"><ol></ol></div>');
	}
};

Comment.prototype.append_children = function(ch, p) {
	for(var i=0; i<ch.length; i++){
		if(!ch[i]) continue;

		if(ch[i].parent == p) {
			$('#comment-reply-'+p + ' ol:first').append(this.gen_comment_item(ch[i]));
			$('#comment-'+ch[i].id).fadeIn();
            TaoBlog.events.dispatch('comment', 'post', $('#comment-'+ch[i].id));
			this.add_reply_div(ch[i].id);
			this.append_children(ch, ch[i].id);
			delete ch[i];
		}
	}
};


Comment.prototype.save_info = function() {
	if(window.localStorage) {
		localStorage.setItem('cmt_author', $('#comment-form input[name=author]').val());
		localStorage.setItem('cmt_email', $('#comment-form input[name=email]').val());
		localStorage.setItem('cmt_url', $('#comment-form input[name=url]').val());
	} else {
		if(typeof show_tips == 'function') {
			show_tips('抱歉，你的浏览器不支持 localStorage，评论者的相关信息将无法正确地保存以便后续使用。');
		}
	}
};

Comment.prototype.load_comments = function() {
    $('#load-comments').click();
};

var comment = new Comment();
comment.init();

