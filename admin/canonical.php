<?php

function home() {
    global $tbopt;
    return 'https://'.$tbopt->get('home');
}

function the_link(&$p, $home=true) {
    global $tbpost;
    global $tbopt;

    $home = $home ? home() : '';
    $link = '';

    if($p->type === 'post') {
        $link = $home.'/'.$p->id.'/';
    } else if($p->type === 'page') {
        $link = $home.$tbpost->get_the_parents_string($p->id).'/'.$p->slug;
    } else {
        $link = '/';
    }

    return $link;
}

function the_id_link(&$p, $home=true) {
    global $tbpost;
    global $tbopt;

    $home = $home ? home() : '';

    return $home . '/' . $p->id . '/';
}

function the_edit_link(&$p, $ret_anchor = true, $blank = false) {
    $link = '/admin/post.php?do=edit&amp;id='.(int)$p->id;

    return $ret_anchor
        ? '<a href="'.$link.'"'.($blank?'target="_blank"':'').'>编辑</a>'
        : $link;
}
