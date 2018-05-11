<?php

add_hook('tb_footer', 'mathjax_footer');

function mathjax_footer() {
    global $tbquery;
    if($tbquery->is_singular()) {
?>
<script>
function hasMath() {
    var has = false;
    $('code:not([class*="lang"])').each(function(_, e){
        var t = $(e).html();
        if(t.startsWith('$') && t.endsWith('$')) {
            has = true;
            // break;
        }
    })
    return has;
}

if(hasMath()) {
    var $body = $('body');

    var config = function(){/*
        MathJax.Hub.Config({
            jax: ["input/TeX", "output/CommonHTML"],
            extensions: ["tex2jax.js"],
            tex2jax: {
                skipTags: [],
                inlineMath: [['$', '$']],
                displayMath: [['$$', '$$']],
            },
            skipStartupTypeset: true,
            //showMathMenu: false,
            menuSettings: {
                zoom: 'Click',
            }
        });
    */}.toString().slice(14,-3);

    $('<script/>')
        .attr('type', 'text/x-mathjax-config')
        .text(config)
        .appendTo($body);

    // jQuery adds _ parameter to skip cache
    var s = document.createElement('script');
    s.src= 'https://cdnjs.cloudflare.com/ajax/libs/mathjax/2.7.4/MathJax.js';
    // s.src = '/plugins/mathjax/MathJax-2.7.4/unpacked/MathJax.js';
    s.async = true;
    $body[0].appendChild(s);

    s = document.createElement('script');
    s.src = '/plugins/mathjax/mathjax.js';
    s.async = true;
    $body[0].appendChild(s);
}
</script>
<?php
    }
}
