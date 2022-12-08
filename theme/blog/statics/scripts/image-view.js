function ImageView() {

}

ImageView.prototype.init =function() {
    this._init();
}

ImageView.prototype._init = function() {
    this._initView();

    this._$imgView = $('#img-view');
    this._$img = $('#img-view > img');
    this._$info = $('#img-view .info');
    this._$tip = $('#img-view .tip');
	this._imageIndex = -1;
	this._scale = 1;
	this._transform = {};

	this._countDown = 3;
	setInterval(function() {
		if(--this._countDown == 0) {
			this._$info.hide();
		}
	}.bind(this), 1000);

    this._keyHandlerAdded = false;

    this._boundEvents = {};

    this._initBindings();

    this._images = $('.entry img');
	this._images.click(this._onPostImageClick.bind(this));

	this._retina = false;
};

ImageView.prototype._onPostImageClick = function(e) {
    this._locateIndex(e.target);
    this.viewImage(e.target);
};

ImageView.prototype._keyHandler = function(e) {
    var handled = true;

    if(e.keyCode == 27 || e.keyCode == 32) {
        // Escape or Space to exit
        this.viewImage(null);
    } else if(e.keyCode == 37) {
        // left to previous
        this.nextImage(-1);
    } else if(e.keyCode ==  39) {
        // right to next
        this.nextImage(1);
    } else if(e.keyCode == 38) {
        // up to rotate anticlockwisely
        this.rotateImage(false);
    } else if(e.keyCode == 40) {
        // down to rotate clockwisely
        this.rotateImage(true);
    } else {
        handled = false;
    }

    if(handled) {
        e.preventDefault();
    }
};

ImageView.prototype._locateIndex = function(img) {
    for(var i=0; i < this._images.length; i++) {
        if(img == this._images[i]) {
            this._imageIndex = i;
            break;
        }
    }
};

ImageView.prototype._showInfo = function(rawWidth, rawHeight, scale) {
    var s = '';
    s += '第 ' + (this._imageIndex+1) + '/' + this._images.length + ' 张';
    s += '，原始尺寸：' + rawWidth + '*' + rawHeight;
    s += '，缩放比例：' +  Math.round(scale*100) + '%';

    this._$info.text(s);
	this._$info.show();
	this._countDown = 3;
};

ImageView.prototype.nextImage = function(dir) {
    this.rotateImage(undefined); // clear rotation first
    this._imageIndex += dir > 0 ? 1 : -1;
    if(this._imageIndex > this._images.length-1) {
        this._imageIndex = this._images.length - 1;
    }
    if(this._imageIndex < 0) {
        this._imageIndex = 0;
    }
    if(this._imageIndex < this._images.length) {
        this.viewImage(this._images[this._imageIndex]);
    }
};

ImageView.prototype.transforms = function() {
	var s = '';
	for (var p in this._transform) {
		s += p + '(' + this._transform[p] + ') ';
	}
	return s;
};

ImageView.prototype.rotateImage = function(fwd) {
    if (fwd == undefined) {
		this._transform = {};
        this._$img.css('transform', this.transforms());
		this._degree = 0;
		this._scale = 1;
    } else {
        this._degree += !!fwd ? +90 : -90;
        this._$img.attr('data-busy', '1');
		this._$img.css('transition', 'transform 0.3s linear');
		this._transform.rotateZ = this._degree + 'deg';
        this._$img.css('transform', this.transforms());
    }
};

ImageView.prototype.viewImage = function(img) {
    if(img != null) {
        // tips
        var tip_times_total = 1;
        var tip_times = + this._$img.attr('data-times') || 0;

        if(tip_times == 0) {
            this._$tip.text('左键拖动，上中下键旋转，滚动缩放，左右切换；双击图片或单击空白区域退出。');
        }

        if(tip_times < tip_times_total) {
            tip_times++;
            this._$img.attr('data-times', tip_times);
        } else if(tip_times == tip_times_total) {
            this._$tip.hide();
        }

        this._$img[0].onload = function() {
            this._rawWidth =  this._$img.prop('naturalWidth');
			this._rawHeight = this._$img.prop('naturalHeight');

            var initScale = this._retina ? 0.5 : 1;
			var initWidth = this._rawWidth * initScale;
			var initHeight = this._rawHeight * initScale;

            {
                var scaleWidth =  this._$imgView.width() / initWidth,
                    scaleHeight = this._$imgView.height() / initHeight;

                // if smaller than container
                if(scaleWidth >= 1 && scaleHeight >= 1) {
					initScale *= 1;
                }
                // if larger than container, scale to fit
                else {
                    initScale *= Math.min(scaleWidth, scaleHeight);
                }

                initWidth = this._rawWidth * initScale;
                initHeight = this._rawHeight * initScale;
            }

            this._$img.css('left', (this._$imgView.width()-initWidth)/2 + 'px');
            this._$img.css('top', (this._$imgView.height()-initHeight)/2 + 'px');
            this._$img.css('width', initWidth + 'px');
            this._$img.css('height', initHeight + 'px');
            this._$imgView.show();

            $('body').css('max-height', window.innerHeight);
            $('body').css('overflow', 'hidden');

            this._showInfo(this._rawWidth, this._rawHeight, initScale);
        }.bind(this);
        this._$img[0].onerror = function() {
            this._showInfo(0, 0, 0);
		}.bind(this);
		var src = img.src;
		this._retina = src.indexOf('@2x.') != -1;
        this._$img.attr('src', src);
    } else {
        // 以下两行清除因拖动导致的设置
        this._$img.css('left', '0px');
        this._$img.css('top', '0px');
        $('body').css('max-height', 'none');
		$('body').css('overflow', 'auto');

		// 清除旋转
		this._transform = {};
        this._$img.css('transform',this.transforms());

		this._dragging = false;
		this._scale = 1;
        this._degree = 0;
        this._$imgView.hide();
    }

    if(img != null) {
        if(!this._keyHandlerAdded) {
            window.addEventListener('keydown', this._boundEvents.keyHandler);
            this._keyHandlerAdded = true;
        }
    } else {
        if(this._keyHandlerAdded){
            window.removeEventListener('keydown', this._boundEvents.keyHandler);
            this._keyHandlerAdded = false;
        }
    }
}

ImageView.prototype._initView = function() {
    var $body = $('body');
    var $imgView = $('<div>')
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
    ;

    $body.append($imgView);
};

ImageView.prototype._initBindings = function() {
    this._boundEvents.keyHandler = this._keyHandler.bind(this);

    this._$img.on('mousedown', this._onImgMouseDown.bind(this));
    this._$img.on('mousemove', this._onImgMouseMove.bind(this));
    this._$img.on('mouseup', this._onImgMouseUp.bind(this));
    this._$img.on('click', this._onImgClick.bind(this));
    this._$img.on('dblclick', this._onImgDblClick.bind(this));
    this._$img.on('transitionend', this._onTransitionEnd.bind(this));

    this._$imgView.on('wheel', this._onDivMouseWheel.bind(this));
    this._$imgView.on('mousemove', this._onDivMouseMove.bind(this));
    this._$imgView.on('click', this._onDivClick.bind(this));
};

ImageView.prototype._onImgMouseDown = function(e) {
    if(this._$img.attr('data-busy') == '1') {
        e.preventDefault();
        return false;
    }

    var target = e.target;

    // http://stackoverflow.com/a/2725963
    if(e.which == 1) {  // left button
        this._offsetX = e.clientX;
        this._offsetY = e.clientY;

        this._coordX = parseInt(target.style.left);
        this._corrdY = parseInt(target.style.top);

        this._dragging = true;
    } else if(e.which == 2) {    // middle button
        this.rotateImage(true);
    }

    e.preventDefault();
    return false;
};

ImageView.prototype._onImgMouseMove = function(e) {
    if(!this._dragging) {
        return false;
    }

    var left = this._coordX + e.clientX - this._offsetX + 'px';
    var top = this._corrdY + e.clientY - this._offsetY + 'px';

    var target = e.target;
    target.style.left = left;
    target.style.top = top;

    e.preventDefault();
    return false;
};

ImageView.prototype._onImgMouseUp = function(e) {
    this._dragging = false;
    e.preventDefault();
    return false;
}

ImageView.prototype._onImgClick =function(e) {
	var smallMove = false;
	{
		var horz = Math.abs(e.clientX - this._offsetX);
		var vert = Math.abs(e.clientY - this._offsetY);

		if(horz <= 1 && vert <= 1) {
			smallMove = true;
		}
	}

	if(smallMove) {
		this.viewImage(null);
	}

    e.preventDefault();
    return false;
};

ImageView.prototype._onImgDblClick = function(e) {
    this.viewImage(null);
    e.preventDefault();
    return false;
};

ImageView.prototype._onTransitionEnd = function(e) {
    this._$img.css('transition', '');
    this._$img.attr('data-busy', '');
};

ImageView.prototype._onDivMouseWheel = function(e) {
    if(this._$img.attr('data-busy') == '1') {
		e.preventDefault();
        return false;
    }

	// https://developer.mozilla.org/en-US/docs/Web/API/Element/wheel_event
	this._scale += e.originalEvent.deltaY * -0.01;
	this._scale = Math.min(Math.max(.125, this._scale), 4);
	this._transform.scale = this._scale;
	this._$img.css('transform', this.transforms());

	this._showInfo(this._rawWidth, this._rawHeight, this._scale);

    e.preventDefault();
    return false;
};

ImageView.prototype._onDivMouseMove = function(e) {
    if(this._dragging) {
        var left = this._corrdX + e.clientX - this._offsetX + 'px';
        var top = this._corrdY + e.clientY - this._offsetY + 'px';
        this._$img.css('left', left);
        this._$img.css('top', top);
        return false;
    }

    return true;
};

ImageView.prototype._onDivClick = function(e) {
    this.viewImage(null);
};

var imgview = new ImageView();
imgview.init();
