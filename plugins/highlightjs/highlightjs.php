<?php

add_hook('tb_head', 'hljs_head');
add_hook('tb_footer', 'hljs_footer');

function hljs_head() { ?>
	<link rel="stylesheet" type="text/css" href="/plugins/highlightjs/monokai_sublime.css" />
<?php }

function hljs_footer() { ?>
	<script type="text/javascript" src="/plugins/highlightjs/highlight.min.js"></script>
	<script type="text/javascript" src="/plugins/highlightjs/hl.js"></script>
<?php }

