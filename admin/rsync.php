<?php
defined('TBPATH') or die('Silence is golden.');

function rsync_posts() {
    global $tbpost;
    global $tbdate;

    $dh = opendir(RSYNC_DIR);
    if($dh) {
        while(($entry = readdir($dh)) !== false) {
            if((int)$entry > 0) {
                $file = RSYNC_DIR . '/' . (int)$entry;
                $st = stat($file);
                $mtime = $tbdate->mysql_datetime_gmt($st['mtime']);
                $content = file_get_contents($file);
                if(strlen($content) > 0) {
                    if($tbpost->rsync_post((int)$entry, $content, $mtime)) {
                        unlink($file);
                    }
                }
            }
        }
        closedir($dh);
    }
}


