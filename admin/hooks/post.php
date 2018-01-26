<?php
/**
 * Hooks of posts
 */

namespace admin\hooks\post;

defined('TBPATH') or die("Silence is golden.");

/**
 * Updates last post time
 * 
 * @return nothing
 */
function updateLastPostTime($id, $post)
{
    global $tbopt;

    $last = $tbopt->get('last_post_time');
    $pdate = $post['date'];

    if (!$last || $pdate >= $last) {
        $tbopt->set('last_post_time', $pdate);
    }
}

add_hook('post_posted', __NAMESPACE__ . '\\updateLastPostTime');

/**
 * Updates post count
 * 
 * @return nothing
 */
function updatePostCount()
{
    global $tbopt;
    global $tbpost;

    $post_count = $tbpost->get_count_of_type('post');
    $page_count = $tbpost->get_count_of_type('page');

    $tbopt->set('post_count', $post_count);
    $tbopt->set('page_count', $page_count);
}

add_hook('post_posted', __NAMESPACE__ . '\\updatePostCount');

/**
 * Filters posts query
 * 
 * @return the filtered SQL string
 */
function beforeQueryPosts($_, $sql)
{
    global $logged_in;

    if ($logged_in) {

    } else {
        $sql['where'][] = "status='public'";
    }

    return $sql;
}

add_hook('before_query_posts', __NAMESPACE__ . '\\beforeQueryPosts');
