<?php

$pd = dirname(__FILE__).'/../plugins';

$dirs = dir($pd);
while($p = $dirs->read()){
    if($p[0] === '.') continue;

    require  "$pd/$p/$p.php";
}

