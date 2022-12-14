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

    $('.entry .toc h2, .entry .toc h3').replaceWith('<div style="margin-bottom: -10px;"><div class=h2 style="float: left; margin-right: 20px;">目录</div><span id="hide-toc" class="no-sel" style="float: right; cursor: pointer;"></span><div style="clear: both;"></div></div>');

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

// 图片来源
(function() {
    function replace(img) {
        var figure = document.createElement('figure');

        var newImg = img.cloneNode(false);
        newImg.removeAttribute('data-name');
        newImg.removeAttribute('data-origin');
        figure.appendChild(newImg);

        var caption = document.createElement('figcaption');
        caption.innerText = '图片来源：';

        var a = document.createElement('a');
        a.innerText = img.getAttribute('data-name') || img.getAttribute('alt');
        a.setAttribute('href', img.getAttribute('data-origin'));
        a.setAttribute('target', '_blank');

        caption.appendChild(a);

        figure.appendChild(caption);

        img.parentNode.insertBefore(figure, img);
        img.parentNode.removeChild(img);
    }

    var imgs = document.querySelectorAll('.entry img[data-origin]');
    imgs.forEach(function(img){ replace(img); });
})();

(function() {
	let anchors = document.querySelectorAll('article a');
	anchors.forEach(a => { TaoBlog.fn.externAnchor(a); });
})();

// 数学公式
(function(){
function hasMath() {
	var math = document.querySelector('.math');
	if (math !== null) {
		return true;
	}

    var has = false;

    // old-style, to-be-removed
	let codes = document.querySelectorAll('code:not([class*="lang"])');
	codes.forEach(function(e) {
		var t = e.innerHTML;
		if(t.startsWith('$') && t.endsWith('$')) {
			has = true;
		}
	});
    return has;
}

if(hasMath()) {
    var s = document.createElement('script');
    s.src = '/plugins/mathjax/mathjax.js';
    s.async = true;
    document.body.appendChild(s);
}
})();

// 代码高亮
(function(){
var s = document.createElement('script');
s.src = '/plugins/highlight/prism.js';
s.async = false;
s.setAttribute('data-manual', '');
s.onload = function() {
	$('pre').each(function(_, re, _){
		TaoBlog.fn.highlight(re);
	});
};
document.body.appendChild(s);

TaoBlog.events.add('comment', 'post', function(jItem) {
	let pres = jItem[0].querySelectorAll(':scope > .comment-content pre');
	pres.forEach(function(pre) {
		TaoBlog.fn.highlight(pre);
	});
});
})();
