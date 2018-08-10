<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<title><?php echo '标签 “',htmlspecialchars($tbquery->tags),'” 下的归档'; ?></title>
</head>
<body>
<ul>
<?php while($tbquery->has()) {
    $the = $tbquery->the();
    echo '<li><a target="_blank" href="', the_link($the, false), '">', htmlspecialchars($the->title), '</a></li>';
} ?>
</ul>
</body>
</html>
