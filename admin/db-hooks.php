<?php

defined('TBPATH') or die("Silence is golden.");

function dbh_on_comment_posted($unused, $POST) {
    global $tbdb;

    $post_id = (int)$POST['post_id'];
    $sql = "UPDATE posts INNER JOIN (SELECT post_id,count(post_id) count FROM comments WHERE post_id=$post_id) x ON posts.id=x.post_id SET posts.comments=x.count WHERE posts.id=$post_id";
    $tbdb->query($sql);
}

add_hook('comment_posted', 'dbh_on_comment_posted');

