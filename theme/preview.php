<?php

defined('TBPATH') or die('Silence is golden.');

?><!doctype html>
<html>
<head>
	<meta charset="UTF-8" />
	<title><?php echo $_POST['title']; ?></title>
	<link rel="stylesheet" type="text/css" href="/theme/style.css" />
	<link rel="stylesheet" type="text/css" href="/theme/font-awesome-4.3.0/css/font-awesome.min.css" />
	<script type="text/javascript" src="/admin/scripts/jquery-2.1.3.min.js"></script>
</head>
<body>
<div id="wrapper">
	<section id="main">
		<div id="content">
<?php echo $_POST['content']; ?>
		</div>
	</section>
</div>
</body>
</html>

