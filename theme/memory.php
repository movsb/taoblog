<?php

function memory_head()
{ ?>
<style>
h1 {
    text-align: center;
    margin: 1em;
}
#list {
    margin-bottom: 3em;
    padding: 0px;
    list-style: none;
}

#list > .item {
    border: 1px solid gray;
    margin-bottom: 2em;
}

#list > .item .meta {
    padding: 0.5em;
    background-color: rgba(0,0,0,0.15);
}

#list > .item .content {
    padding: 1em;
}
</style>
<?php }

add_hook('tb_head', 'memory_head');

require('header.php');


echo '<h1>说说</h1>';

$sss = $tbshuoshuo->get_latest(50);
if(!is_array($sss) || count($sss) == 0) die(-1);

?><ul id="list">
<?php
foreach($sss as &$ss) {
    echo '<li class="item">';
    echo '<div class="meta">', substr($ss->date,0,16), ' @ ', $ss->geo_addr, '</div>';
    echo '<div class="content">', $ss->content, '</div>';
    echo '</li>';
}
?>
</ul>
<?php
require('footer.php');

