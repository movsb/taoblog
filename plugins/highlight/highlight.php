<?php

add_hook('tb_footer', 'hljs_footer');

function hljs_footer() { ?>
	<script type="text/javascript" src="/plugins/highlight/hl.js"></script>
<?php }

