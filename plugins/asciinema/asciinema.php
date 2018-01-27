<?php

add_hook('tb_head', 'asciinema_head');

function asciinema_head() {
    global $tbquery;

    if($tbquery->is_singular()) {
        global $the;
        if(preg_match('~<pre[^>]+\basciinema\b[^>]+>~', $the->content)) {?>
            <link rel="stylesheet" href="//blog-10005538.file.myqcloud.com/asciinema-player.css" />
            <style>
                .asciinema-player .start-prompt,
                .asciinema-player .control-bar
                {
                    display: none;
                }
                .asciinema-player-wrapper {
                    text-align: left;
                }

                pre.asciinema-terminal {
                    box-sizing: content-box !important;
                    font-family: Consolas, Menlo, 'Bitstream Vera Sans Mono', monospace, 'Powerline Symbols' !important;
                }
            </style>
        <?php }
    }
}

add_hook('tb_footer', 'asciinema_footer');

function asciinema_footer() { 
    global $tbquery;
    if($tbquery->is_singular()) {
        global $the;
        if(preg_match('~<pre[^>]+\basciinema\b[^>]+>~', $the->content)) { ?>
            <script src="//blog-10005538.file.myqcloud.com/asciinema-player.js"></script>
            <script>
                $('.asciinema').each(function(i, e) {
                    var e = $(e);
                    var file = e.attr('data-file');
                    var id = 'asciinema: ' + file;
                    
                    $.getJSON(file, function(data) {
                        $('<div/>').attr('id', id).insertAfter(e);
                        asciinema.player.js.CreatePlayer(id, '', {
                                width: data.width, height: data.height,
                                poster: "data:text/plain," + data.stdout,
                            });
                        e.remove();
                    });
                });
            </script>
        <?php }
    }
}

