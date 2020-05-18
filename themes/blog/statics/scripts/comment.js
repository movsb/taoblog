document.write(function(){/*
	<!--评论标题 -->
	<h3 id="comment-title">
        文章评论
        <span class="no-sel count-wrap"> <span class="loaded">0</span>/<span class="total" itemprop="commentCount">0</span></span>
        <span class="post-comment">发表评论</span>
	</h3>

	<!-- 评论列表  -->
	<ol id="comment-list">

	</ol>

	<!-- 评论功能区  -->
	<div class="comment-func">
        <input type="hidden" id="post-id" name="post-id" value="" />
        <div>还没有用户发表评论，<span class="post-comment">发表评论</span></div>
	</div>

	<!-- 评论框 -->
	<div id="comment-form-div" class="normal">
		<div class="comment-form-div-1 no-sel">
			<div class="no-sel" class="nc">
                <div class="ncbtns">
                    <img src="/images/close.svg" width="20" height="20" title="隐藏" class="closebtn"/>
                    <img src="/images/question.svg" width="20" height="20" title="支持插入代码片段，请使用3个反引号括起来，如：

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

			<form id="comment-form">
                <div style="display: none;">
                    <input id="comment-form-post-id" type="hidden" name="post_id" value="" />
                    <input id="comment-form-parent"  type="hidden" name="parent" value="" />
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

	self.convert_commenter();

    $('.post-comment').click(function(){
        self.reply_to(0);
    });

    // Ajax评论提交
    $('#comment-submit').click(function() {
        $(this).attr('disabled', 'disabled');
        $('#comment-submit').val('提交中...');
        $.post(
            '/v2/posts/' + $('#post-id').val() + '/comments',
            $('#comment-form').serialize(),
            function(cmt){
				cmt = self.normalize_comment(cmt);
                var parent = $('#comment-form input[name="parent"]').val();
                if(parent == 0) {
                    $('#comment-list').prepend(self.gen_comment_item(cmt));
                    // 没有父评论，避免二次加载。
                    self._loaded ++;
                } else {
                    self.add_reply_div(parent);
                    $('#comment-reply-'+parent + ' ol:first').append(self.gen_comment_item(cmt));
                }
                $('#comment-'+cmt.id).fadeIn();
                TaoBlog.events.dispatch('comment', 'post', $('#comment-'+cmt.id));
                $('#comment-content').val('');
                $('#comment-form-div').fadeOut();
                self.save_info();
                self._count++;
                self.toggle_post_comment_button();
            },
            'json'
        )
        .fail(
            function(xhr, sta, e){
                alert(JSON.parse(xhr.responseText).message);
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
            self.toggle_post_comment_button(self._count == 0);
        });
    });
};

Comment.prototype.toggle_post_comment_button = function(show) {
    if (typeof show == 'undefined') {
        if (this._count <= 0) {
            show = true
        } else if(this._count == 1) {
            show = false;
        } else {
            return;
        }
    }
    if(show) {
        $('#comment-title .post-comment').fadeOut();
        $('.comment-func').fadeIn();
    } else {
        $('#comment-title .post-comment').fadeIn();
        $('.comment-func').fadeOut();
    }
};

Comment.prototype.load_essential_comments = function() {
    if(this._loaded + this._loaded_ch < this._count
        && window.scrollY + window.innerHeight + 100 >= document.body.scrollHeight) 
    {
        this.load_comments();
    }
};

Comment.prototype.get_count = function(callback) {
    var self = this;
    var pid = $('#post-id').val();
    $.get('/v2/posts/' + pid + '/comments!count',
        function(data) {
            self._count = data;
            $('#comment-title .total').text(self._count);
            callback();
        },
        'json'
    );
};

Comment.prototype.gen_avatar = function(eh, sz) {
	return '/v2/avatar?' + encodeURIComponent(eh + '?d=mm&s=' + sz);
};

Comment.prototype.normalize_content = function(c) {
    var s = this.h2t(c);
    s = s.replace(/```(\s*(\w+)\s*)?\r?\n([\s\S]+?)```/mg, '<pre class="code"><code class="language-$2">$3</code></pre>');
    return s;
};

Comment.prototype.normalize_comment = function(c) {
	c.author = c.author || '';
	c.email = c.email || '';
	c.url = c.url || '';
	c.ip = c.ip || '';
	c.children = c.children || [];
	c.is_admin = c.is_admin || false;
	return c;
}

Comment.prototype.friendly_date = function(d) {
	let date = new Date(d.seconds * 1000);
	return date.toLocaleString();
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

    var loggedin = cmt.ip != '';

    // 登录后可以显示评论者的详细信息
    var info = '';
    if(loggedin) {
        // author, email, url, ip, date
		info = '编号：' + cmt.id
			+ '\n作者：' + this.h2a(cmt.author)
            + '\n邮箱：' + this.h2a(cmt.email)
            + '\n网址：' + this.h2a(cmt.url)
            + '\n地址：' + cmt.ip
            + '\n日期：' + new Date(cmt.date.seconds*1000).toLocaleString()
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
			nickname = '<a rel="nofollow" target="_blank" href="' + this.h2a(cmt.url) + '">' + this.h2t(cmt.author) + '</a>';
		} else {
			nickname = this.h2t(cmt.author);
		}

		s += '<span class="nickname">' + nickname + '</span>\n';
	}

    s += '<time class="date" datetime="' + (new Date(cmt.date.seconds*1000)).toJSON() + '">' + this.friendly_date(cmt.date) + '</time>\n</div>\n';
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
		var commenter = JSON.parse(localStorage.getItem('commenter') || '{}');
		$('#comment-form input[name=author]').val(commenter.name || '');
		$('#comment-form input[name=email]').val(commenter.email || '');
		$('#comment-form input[name=url]').val(commenter.url || '');
	}

	$('#comment-form-div').fadeIn();
    $('#comment-content').focus();
};

Comment.prototype.delete_me = function(p) {
    var self = this;
    var pid = $('#post-id').val();
	$.ajax({
        url: '/v2/posts/' + pid + '/comments/' + p,
        type: 'DELETE',
        success: function() {
            $('#comment-'+p).remove();
            self._count--;
            self.toggle_post_comment_button();
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
			ch[i] = this.normalize_comment(ch[i]);
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
		var commenter = {
			name: $('#comment-form input[name=author]').val(),
			email: $('#comment-form input[name=email]').val(),
			url: $('#comment-form input[name=url]').val(),
		};
		localStorage.setItem('commenter', JSON.stringify(commenter));
	} else {
		if(typeof show_tips == 'function') {
			show_tips('抱歉，你的浏览器不支持 localStorage，评论者的相关信息将无法正确地保存以便后续使用。');
		}
	}
};

Comment.prototype.load_comments = function() {
    if (this.loading) {
        return;
    }
    this.loading = true;

    var self = this;
    var pid = $('#post-id').val();

    $.get('/v2/posts/' + pid + '/comments',
        {
            limit: 10,
            offset: self._loaded,
        },
        function(resp) {
			let cmts = resp.comments;
            var ch_count = 0;
            for(var i=0; i<cmts.length; i++){
				cmts[i] = self.normalize_comment(cmts[i]);
                $('#comment-list').append(self.gen_comment_item(cmts[i]));
                $('#comment-'+cmts[i].id).fadeIn();
                TaoBlog.events.dispatch('comment', 'post', $('#comment-'+cmts[i].id));
                self.add_reply_div(cmts[i].id);
                if(cmts[i].children) {
                    self.append_children(cmts[i].children, cmts[i].id);
                    ch_count += cmts[i].children.length;
                }
            }

            if(typeof(jQuery)=='function' && typeof(jQuery.timeago)=='function')
                jQuery('.comment-meta .date').timeago();

            if(cmts.length != 0) {
                self._loaded += cmts.length;
            }
            self._loaded_ch += ch_count;
            $('#comment-title .loaded').text(self._loaded + self._loaded_ch);
        },
        'json'
    ).fail(function(x) {
        alert(x.responseText);
    }).always(function(){
        self.loading = false;
    });
};

Comment.prototype.convert_commenter = function() {
	if(window.localStorage && !localStorage.getItem('commenter') && localStorage.getItem('cmt_author')) {
		var commenter = {
			name: localStorage.getItem('cmt_author') || '',
			email: localStorage.getItem('cmt_email') || '',
			url: localStorage.getItem('cmt_url') || '',
		};
		localStorage.setItem('commenter', JSON.stringify(commenter));
		localStorage.removeItem('cmt_author');
		localStorage.removeItem('cmt_email');
		localStorage.removeItem('cmt_url');
	}
};

var comment = new Comment();
comment.init();
