<?php

add_hook('tb_footer', 'timeago_footer');

function timeago_footer() { 
	global $tbquery;
	if($tbquery->is_singular()) {?>
        <script type="text/javascript" src="/plugins/timeago/timeago.js"></script>
        <script type="text/javascript">
            $('.comment-meta .date').timeago();
            TaoBlog.events.add('comment', 'post', function(item) {
                $(item).find('.date').timeago();
            });
        </script>
    <?php }
}

