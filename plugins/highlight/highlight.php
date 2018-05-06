<?php

add_hook('tb_footer', 'hljs_footer');

function hljs_footer() { 
    global $tbquery;
    if($tbquery->is_singular()) {?>
    <script type="text/javascript" src="/plugins/highlight/prism.js" data-manual></script>
<script>
$('pre').each(function(_, re, _){
    var e = $(re);
    var lang = e.attr('lang');
    // https://stackoverflow.com/a/1318091/3628322
    var hasLang = typeof lang !== typeof undefined && lang !== false;
    var hasCode = e.find('>code').length > 0;
    // console.log(re, hasLang, hasCode);
    if(hasLang && !hasCode) {
        var code = $('<code/>').html(e.html());
        code.addClass("language-" + lang);
        e.removeAttr('lang');
        e.html('');
        e.append(code);
        hasCode = true;
    }
    if(hasCode) {
        e.removeClass('code');
        // TODO
        // e.addClass('line-numbers');
        Prism.highlightAllUnder(re);
    }
});
</script>
<?php }
}

add_hook('tb_head', 'hljs_header');

function hljs_header() { 
    global $tbquery;
    if($tbquery->is_singular()) {?>
    <link rel="stylesheet" href="/plugins/highlight/prism.css" />
<?php }
}

