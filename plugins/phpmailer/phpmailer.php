<?php

function pm_mail($subject, $body) {
	require 'class.smtp.php';
	require 'class.phpmailer.php';

	$mail = new PHPMailer;

	//$mail->SMTPDebug = 5;									// Enable verbose debug output

	$mail->isSMTP();										// Set mailer to use SMTP
	$mail->CharSet = 'utf-8';

	$mail->Host = 'smtp.qq.com';							// Specify main and backup SMTP servers
	$mail->SMTPAuth = true;									// Enable SMTP authentication
	$mail->Username = 'blog@twofei.com';					// SMTP username
	$mail->Password = '*************';                      // SMTP password
	$mail->SMTPSecure = 'ssl';								// Enable TLS encryption, `ssl` also accepted
	$mail->Port = 465;										// TCP port to connect to

	$mail->From = 'blog@twofei.com';
	$mail->FromName = '博客评论';
	$mail->addAddress('anhbk@qq.com', '女孩不哭');			// Add a recipient
	$mail->isHTML(true);									// Set email format to HTML

	$mail->Subject = '[博客评论] '.$subject;
	$mail->Body    = $body;

	return $mail->send();
}

function pm_comment_posted($unused1, $arg) {
	global $tbpost;

	$subject = $tbpost->get_title((int)$arg['post_id']);

	$body = "<b>您的博文“".$subject."”有新的评论啦！</b><br><br>";
	$body .= "<b>内容：</b>".$arg['content']."<br><br>";
	$body .= "<b>作者：</b>".$arg['author']."<br><br>";

	return pm_mail($subject, $body);
}

add_hook('comment_posted', 'pm_comment_posted');

