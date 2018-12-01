<?php

function home() {
    return 'https://'.get_opt('home');
}

function the_link(&$p, $home=true) {
    global $tbpost;

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
