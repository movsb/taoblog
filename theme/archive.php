<?php

function list_all_tags() {
    global $tbtag;

    $tags = $tbtag->list_all_tags();
    foreach($tags as &$t) {
        echo '<a href="/tags/',urlencode($t->name),'">',htmlspecialchars($t->name),
            '<span>', $t->size,'</span>','</a>';
    }
    unset($t);
}

require('header.php');
?>
<div class="archives">
<div class="tags no-sel">
<h2>标签</h2>
    <?php list_all_tags(); ?>
</div>
</div>

<?php
require('footer.php');

