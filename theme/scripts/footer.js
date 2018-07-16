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

/* 点击图片放大 & 拖动浏览*/
/* 写得超级烂，完全没管性能 */
(function() {
    var body = $('body');

    body.append(
        $('<div>')
            .attr('class', 'img-view')
            .attr('id', 'img-view')
            .append($('<img/>'))
            .append($('<div/>')
                .attr('class', 'tool')
                .append(
                    $('<div/>')
                        .attr('class', 'info')
                )
                .append(
                    $('<div/>')
                        .attr('class', 'tip')
                )
            )
        );

    var imgdiv = $('#img-view');
    var img = $('#img-view > img');
    var images = $('.entry img:not(.nz)');
    var image_index = -1;
    var key_handler_added = false;

    // current image's raw width & height
    var curImageRawHeight = 0, curImageRawWidth = 0;

    // show_info timer
    var tmrShowInfo = null;

    function key_handler(e) {
        var handled = true;

        if(e.keyCode == 27 || e.keyCode == 32) {
            // Escape or Space to exit
            view_image(null);
        } else if(e.keyCode == 37) {
            // left to previous
            next_image(-1);
        } else if(e.keyCode ==  39) {
            // right to next
            next_image(1);
        } else if(e.keyCode == 38) {
            // up to rotate anticlockwisely
            rotateImage(false);
        } else if(e.keyCode == 40) {
            // down to rotate clockwisely
            rotateImage(true);
        } else {
            handled = false;
        }

        if(handled) {
            e.preventDefault();
        }
    };

    function set_current_image(ele) {
        for(var i=0; i<images.length; i++) {
            if(ele == images[i]) {
                image_index = i;
                break;
            }
        }
    }

    function show_info(rawWidth, rawHeight, scale) {
        var s = '';
        s += '第 ' + (image_index+1) + '/' + images.length + ' 张';
        s += '，原始尺寸：' + rawWidth + '*' + rawHeight;
        s += '，缩放比例：' +  Math.round(scale*100) + '%';
        var e = $('#img-view .info');
        e.text(s);
        e.show();
        if(tmrShowInfo != null) {
            clearTimeout(tmrShowInfo);
        }
        tmrShowInfo = setTimeout(function() {
            e.hide();
            tmrShowInfo = null;
        }, 3000);
    }

    function view_image(ele) {
        if(ele != null) {
            // tips
            var tip_times_total = 1;
            var tip_times = +img.attr('data-times') || 0;

            if(tip_times == 0) {
                $('#img-view .tip').text('左键拖动，上中下键旋转，滚动缩放，左右切换；双击图片或单击空白区域退出。');
            }

            if(tip_times < tip_times_total) {
                tip_times++;
                img.attr('data-times', tip_times);
            } else if(tip_times == tip_times_total) {
                $('#img-view .tip').hide();
            }

            body.css('max-height', window.innerHeight);
            body.css('overflow', 'hidden');

            img[0].onload = function() {
                curImageRawWidth = img.prop('naturalWidth');
                curImageRawHeight = img.prop('naturalHeight');

                var initScale = 1;
                var initWidth = 0, initHeight = 0;
                {
                    var scaleWidth = imgdiv.width() / curImageRawWidth,
                        scaleHeight = imgdiv.height() / curImageRawHeight;

                    // if smaller than container
                    if(scaleWidth >= 1 && scaleHeight >= 1) {
                        initScale = 1;
                    }
                    // if larger than container, scale to fit
                    else {
                        initScale = Math.min(scaleWidth, scaleHeight);
                     }

                    initWidth = curImageRawWidth * initScale;
                    initHeight = curImageRawHeight * initScale;
                }

                img.css('left', (imgdiv.width()-initWidth)/2 + 'px');
                img.css('top', (imgdiv.height()-initHeight)/2 + 'px');
                img.css('width', initWidth + 'px');
                img.css('height', initHeight + 'px');
                imgdiv.show();

                show_info(curImageRawWidth, curImageRawHeight, initScale);
            };
            img.attr('src', ele.src);
        } else {
            // 以下两行清除因拖动导致的设置
            img.css('left', '0px');
            img.css('top', '0px');
            body.css('max-height', 'none');
            body.css('overflow', 'auto');
            // 清除旋转
            img.css('transform','');

            imgview.dragging = false;
            imgview.degree = 0;
            imgdiv.hide();
        }

        if(ele != null) {
            if(!key_handler_added) {
                window.addEventListener('keydown', key_handler);
                key_handler_added = true;
            }
        } else {
            if(key_handler_added){
                window.removeEventListener('keydown', key_handler);
                key_handler_added = false;
            }
        }
    }

    function next_image(dir) {
        rotateImage();
        image_index += dir > 0 ? 1 : -1;
        if(image_index > images.length-1) image_index = images.length - 1;
        if(image_index < 0) image_index = 0;
        if(image_index < images.length) {
            var img = images[image_index];
            view_image(img);
        }
    }

    $('.entry img:not(.nz)').click(function(e) {
        set_current_image(this);
        view_image(this);
    });

    imgdiv.click(function() {
        view_image(null);
    });

    window.imgview = {};
    imgview.dragging = false;
    imgview.degree = 0;

    img.on('mousedown', function(e) {
        if(img.attr('data-busy') == '1') {
            e.preventDefault();
            return false;
        }

        var target = e.target;
        
        // http://stackoverflow.com/a/2725963
        if(e.which == 1) {  // left button
            imgview.offset_x = e.clientX;
            imgview.offset_y = e.clientY;

            imgview.coord_x = parseInt(target.style.left);
            imgview.coord_y = parseInt(target.style.top);

            imgview.dragging =true;
        } else if(e.which == 2) {    // middle button
            rotateImage(true);
        }

        e.preventDefault();
        return false;
    });

    function rotateImage(clockwisely) {
        if (clockwisely == undefined) {
            img.css('transform', '');
            imgview.degree = 0;
        } else {
            imgview.degree += !!clockwisely ? +90 : -90;
            img.attr('data-busy', '1');
            img.css('transition', 'transform 0.3s linear');
            img.css('transform', 'rotateZ(' + imgview.degree + 'deg)');
        }
    }

    img.on('mousemove', function(e) {
        if(!imgview.dragging) return;

        var target = e.target;
        target.style.left = imgview.coord_x + e.clientX - imgview.offset_x + 'px';
        target.style.top = imgview.coord_y + e.clientY - imgview.offset_y + 'px';

        e.preventDefault();
        return false;
    });

    img.on('mouseup', function(e) {

        imgview.dragging = false;
        e.preventDefault();
        return false;
    });

    imgdiv.on('mousemove', function(e) {
        if(imgview.dragging) {
            img.css('left', imgview.coord_x + e.clientX - imgview.offset_x + 'px');
            img.css('top', imgview.coord_y + e.clientY - imgview.offset_y + 'px');
            return false;
        }

        return true;
    });

    img.on('click', function(e) {
        e.preventDefault();
        return false;
    });

    img.on('dblclick', function(e) {
        view_image(null);
        
        e.preventDefault();
        return false;
    });

    img.on('transitionend', function(){
        img.css('transition', '');
        img.attr('data-busy', '');
    });

    imgdiv.on('wheel', function(e) {
        if(img.attr('data-busy') == '1') {
            e.preventDefault();
            return false;
        }

        var x = e.originalEvent.clientX;
        var y = e.originalEvent.clientY;
        var left = parseInt(img.css('left'));
        var top = parseInt(img.css('top'));
        var width = parseInt(img.css('width'));
        var height = parseInt(img.css('height'));
        var zoomin = e.originalEvent.deltaY < 0;

        var scale = 1.5;

        var new_width = zoomin ? width * scale : width / scale;
        var new_height = zoomin ? height * scale : height / scale;
        var new_left = x >= left && x < left + width
            ? left + (width - new_width) * ((x-left)/width)
            : left + (width - new_width) / 2;
        var new_top = y >= top && y <= top + height
            ? top + (height - new_height) * ((y-top)/height)
            : top + (height - new_height) / 2;

        if(new_width >= 1 && new_height >= 1) {
            img.attr('data-busy', '1');
            img.css('transition', 'all 0.3s linear 0s');
            img.css('left', new_left + 'px');
            img.css('top', new_top + 'px');
            img.css('width', new_width + 'px');
            img.css('height', new_height + 'px');

            show_info(curImageRawWidth, curImageRawHeight, new_width / curImageRawWidth);
        }

        e.preventDefault();
        return false;
    });
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
