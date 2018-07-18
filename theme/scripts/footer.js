/* 回到顶端 */
(function(){
    var header_height = $('#header').height();
    var topIsShowing = false;
    var $topElement = $('#back-to-top');
    var fadeDuration = 500;

    window.addEventListener('scroll', function() {
        if (this.window.scrollY > header_height) {
            if (!topIsShowing) {
                topIsShowing = true;
                $topElement.fadeIn(fadeDuration);  // no wait on animation
                console.log("back-top-top: fadeIn");
            }
        } else {
            if (topIsShowing) {
                $topElement.fadeOut(fadeDuration);
                topIsShowing = false;
                console.log("back-to-top: fadeOut");
            }
        }
    });

    $topElement.click(function(){
        $('html,body').animate({
            scrollTop: 0
        }, 300);
    })
})();

/* 目录展开与隐藏 */
(function() {
    if($('.entry .toc').length == 0) return;

    function hide_toc(hide) {
        var hide_toc = $('#hide-toc');
        var toc_ul = $('.entry .toc > ul');

        if(hide) {
            hide_toc.text('[显示]');
            toc_ul.hide();
        } else {
            hide_toc.text('[隐藏]');
            toc_ul.show();
        }
    }

    $('.entry .toc h2, .entry .toc h3').replaceWith('<div style="margin-bottom: -10px;"><h2 style="float: left; margin-right: 20px;">目录</h2><span id="hide-toc" class="no-sel" style="float: right; cursor: pointer;"></span><div style="clear: both;"></div></div>');

    hide_toc(true);

    $('#hide-toc').click(function() {
        var hidden = $('.entry .toc > ul').css('display') == 'none';
        hide_toc(!hidden);
    });

    window.addEventListener('keyup', function(e) {
        if(e.keyCode == 27) {
            hide_toc(true);
        }
    });

  // copy from https://codemirror.net/doc/activebookmark.js
  if (!window.addEventListener) return;
  var pending = false, prevVal = null;

  function updateSoon() {
    if (!pending) {
      pending = true;
      setTimeout(update, 250);
    }
  }

  function update() {
    pending = false;
    var marks = document.getElementsByClassName("toc")[0].getElementsByTagName("a"), found;
    for (var i = 0; i < marks.length; ++i) {
      var mark = marks[i], m;
      if (mark.getAttribute("data-default")) {
        if (found == null) found = i;
      } else if (m = mark.href.match(/#(.*)/)) {
        var ref = document.querySelector('a[name="' + m[1] + '"]');
        if (ref && ref.getBoundingClientRect().top < 50)
          found = i;
      }
    }

    if (found != null && found != prevVal) {
      prevVal = found;
      var lis = document.getElementsByClassName("toc")[0].getElementsByTagName("li");
      for (var i = 0; i < lis.length; ++i) lis[i].className = "";
      for (var i = 0; i < marks.length; ++i) {
        if (found == i) {
          marks[i].className = "active";
          for (var n = marks[i]; n; n = n.parentNode)
            if (n.nodeName == "LI") n.className = "active";
        } else {
          marks[i].className = "";
        }
      }
    }
  }

  window.addEventListener("scroll", updateSoon);
  window.addEventListener("load", updateSoon);
  window.addEventListener("hashchange", function() {
    setTimeout(function() {
      var hash = document.location.hash, found = null, m;
      var marks = document.getElementsByClassName("toc")[0].getElementsByTagName("a");
      for (var i = 0; i < marks.length; i++)
        if ((m = marks[i].href.match(/(#.*)/)) && m[1] == hash) { found = i; break; }
      if (found != null) for (var i = 0; i < marks.length; i++)
        marks[i].className = i == found ? "active" : "";
    }, 300);
  });
})();

/* pre 的双击全选与全窗口 */
(function() {
    $('.entry pre, .entry code').on('dblclick', function(e) {
        var t = e.target.tagName;
        if(t == 'PRE' || t == 'CODE') {
            var selection = window.getSelection();
            var range = document.createRange();
            range.selectNodeContents(e.target);
            selection.removeAllRanges();
            selection.addRange(range);
            e.preventDefault();
            return;
        }
    });
})();

/* 服务器运行时间 */
// http://www.htmlgoodies.com/html5/javascript/calculating-the-difference-between-two-dates-in-javascript.html
(function() {
    function date_between(date1, date2) {
        // get one day in millisecond
        var one_day = 24*60*60*1000;

        // convert to millisecond
        var date1_ms = date1.getTime();
        var date2_ms = date2.getTime();

        // calc diff
        var diff_ms = date2_ms - date1_ms;

        return Math.round(diff_ms / one_day);
    }

    var start = new Date(2014, 12-1, 24);
    var now = new Date();

    var days = date_between(start, now);

    $('#server-run-time').text(days);
})();

/* 表情图片 */
(function() {
    $('.entry i.smiley').each(function(i, e) {
        var that = $(this);
        var link = that.attr('data-link');
        if(!link) return;

        var spec = link.match(/(\d+)\.(.+)/);
        if(!spec || spec.length < 3) return;

        function __make_ele(spec) {
            if(spec[1] == 1) { // QQ
                var s = '<img class="nz smiley" src="/smileys/qq/' + spec[2] + '.gif"';
                s += ' alt="' + $('<div/>').text(that.text()).html() + '" />';
                return s;
            }

            return "";
        }

        var the_ele = __make_ele(spec);
        if(the_ele)
            that.replaceWith(the_ele);
    });
})();

(function(){
    var $raws = $('i[data-aes2htm]');
    $raws.each(function(_, e){
        file = $(e).attr('data-aes2htm');
        $code = $(e).parent().prev().find('code');
        $.get(file, function(data){
            $.post('/v1/tools/aes2htm',
                {
                    source: data,
                },
                function(data) {
                    $code.html(data);
                }
            ).fail(function(x) {
                console.error(x);
            });
        });
    });
})();
