<?php

defined('TBPATH') or die('Silence is golden');

function pm_mail($recipient, $nickname, $subject, $body) {
    $url = 'http://127.0.0.1:2564/.v1/send_mail';

    $data = array(
        'author' => $nickname,
        'email' => $recipient,
        'subject' => $subject,
        'body' => $body
    );
    
    // use key 'http' even if you send the request to https://...
    $options = array(
        'http' => array(
            'header'  => "Content-type: application/x-www-form-urlencoded\r\n",
            'method'  => 'POST',
            'content' => http_build_query($data)
        )
    );
    $context  = stream_context_create($options);
    $result = file_get_contents($url, false, $context);
}

function pm_notify_admin(&$arg) {
    global $tbopt;
    global $tbmain;

    $title_encoded   = htmlspecialchars($arg->post_title);
    $content_encoded = htmlspecialchars($arg->content);
    $url_encoded     = htmlspecialchars($arg->url);
    $author_encoded  = htmlspecialchars($arg->author);

    $subject = '[新博文评论] '.$arg->post_title;
    $link = $tbmain->home.'/?p='.$arg->post_id.'#comments';

    $body = "<b>您的博文“{$title_encoded}”有新的评论啦！</b><br><br>";
    $body .= "<b>链接：</b>{$link}<br>";
    $body .= "<b>作者：</b>{$author_encoded}<br>";
    $body .= "<b>邮箱：</b>{$arg->email}<br>";
    $body .= "<b>网址：</b>{$url_encoded}<br>";
    $body .= "<b>时间：</b>{$arg->date}<br>";
    $body .= "<b>内容：</b>{$content_encoded}<br>";

    return pm_mail($tbopt->get('email'), $tbopt->get('nickname'), $subject, $body);
}

function pm_notify_user(&$arg) {
    global $tbopt;
    global $tbmain;

    $title_encoded   = htmlspecialchars($arg->post_title);
    $content_encoded = htmlspecialchars($arg->content);
    $author_encoded  = htmlspecialchars($arg->author);

    $subject = '[回复评论] '.$arg->post_title;
    $link = $tbmain->home.'/?p='.$arg->post_id.'#comments';

    $body = "<b>您在博文“{$title_encoded}”的评论有新的回复啦！</b><br><br>";
    $body .= "<b>链接：</b>{$link}<br>";
    $body .= "<b>作者：</b>{$author_encoded}<br>";
    $body .= "<b>时间：</b>{$arg->date}<br>";
    $body .= "<b>内容：</b>{$content_encoded}<br>";
    $body .= "<br>该邮件为系统自动发出，请勿直接回复该邮件。<br>";

    foreach($arg->parents as &$pc) {
        pm_mail($pc->email, $pc->author, $subject, $body);
    }
}

function pm_comment_posted($unused1, $a) {
    global $tbopt;
    global $tbpost;
    global $tbcmts;

    $admin_email = $tbopt->get('email');

    $arg = (object)$a;

    $arg->post_title = $tbpost->get_title((int)$arg->post_id);

    // 通知站长
    if($arg->email != $admin_email) {
        pm_notify_admin($arg);
    }

    // 通知父评论
    $parent = (int)$arg->parent;
    $arg->parents = [];

    while($parent > 0) {
        $pc = $tbcmts->get_vars('id,author,email', "id=$parent");
        if(is_object($pc) && $pc->email != $admin_email && $pc->email != $arg->email) {
            $arg->parents[] = $pc;
        }

        $parent = $pc->id;
        break; // 暂时不通知父父级评论
    }

    if(count($arg->parents)) {
        pm_notify_user($arg);
    }
}

add_hook('comment_posted', 'pm_comment_posted');

