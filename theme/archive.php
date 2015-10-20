<?php

function list_all_cats() {
	global $tbtax;
	$taxes = $tbtax->get_hierarchically();

    function _tax_add(&$taxes) {
        $s = '';
        foreach($taxes as $t) {
            $has_sons = isset($t->sons) && count($t->sons);
            $s .= '<li class="'.($has_sons ? 'parent ' : 'child').'">'
                .($has_sons ? '<span class="expandable">&gt;</span>&nbsp;' : '<span class="dash">-</span>&nbsp;')
                .'<span style="cursor: pointer;">'.$t->name."</span>\n";
            if($has_sons) {
                $s .= '<ul>';
                $s .= _tax_add($t->sons);
                $s .= "</ul>\n";
            }
            $s .= '</li>'."\n";
        }
        return $s;
    }


	$content = '<ul>'._tax_add($taxes).'</ul>';
    echo $content;
}

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
    <div class="cats no-sel">
        <h2>分类</h2>
        <?php list_all_cats(); ?>
    </div>
</div>

<?php

function tb_footer_hook() { ?>
<script>
    $('.cats').on('click',function(e) {
        var target = $(e.target);
        if(target.hasClass('expandable')) {
            target.parent().find('>ul').toggle();
            target.toggleClass('expanded');
        }
        e.stopPropagation();
    });
</script>
<?php
}

add_hook('tb_footer', 'tb_footer_hook');
require('footer.php');

