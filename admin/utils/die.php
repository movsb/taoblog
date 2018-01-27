<?php

function tb_die($code, $str){
    $status = [
        200 => 'HTTP/1.1 200 OK',
        403 => 'HTTP/1.1 403 Forbidden',
        400 => 'HTTP/1.1 400 Bad Request',
        503 => 'HTTP/1.1 503 Service Unavailable',
    ];
    header($status[$code]);
?>
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <title>TaoBlog</title>
</head>
<body>
    <?php echo $str; ?>

</body>
</html>

<?php 
    die(-1);
}

