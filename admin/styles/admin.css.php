<?php 
header('HTTP/1.1 200 OK');
header('Content-Type: text/css');

$admin_top = '32px';
$admin_left = '120px';

?>

* {
	margin: 0px;
	padding: 0px;
}

body {
	font-family: "Microsoft YaHei";
	padding-top: <?php echo $admin_top; ?>;
	padding-left: <?php echo $admin_left; ?>;
	background-color: #F1F1F1;
}

#admin-top {
	background-color: purple;
	color: white;
	font-size: 1.5em;
	position: fixed;
	left: 0px;
	top: 0px;
	right: 0px;
	height: <?php echo $admin_top; ?>;
	padding: 0px 0.5em 0px 0.5em;
}

#admin-top .logo {
	width: <?php echo $admin_left; ?>;
	padding-top: 0.5em;
}

#admin-top .logout {
	position: absolute;
	right: 0px;
}

#admin-left {
	background-color: purple;
	color: white;
	font-size: 1.5em;
	position: fixed;
	left: 0px;
	top: <?php echo $admin_top; ?>;
	bottom: 0px;
	width: <?php echo $admin_left; ?>;
}

#admin-left ul {
	list-style: none;
}

#admin-left ul li .menu {
	padding: 0.4em;
}

#admin-left ul li:hover {
	background-color: #C34343;
}

#admin-left ul a {
	text-decoration: none;
	color: inherit;
	outline: 0px none;
}

#admin-wrap {
	padding: 1em;
}

