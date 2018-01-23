<?php

// 这代码加这命名加这逻辑，我估计你得看醉，反正我是醉了
function list_all_cats() {
	global $tbtax;
    global $tbpost;

	$taxes = $tbtax->get_hierarchically();
    $cat_posts = $tbpost->get_count_of_cats_all();

    $_tax_add = function( &$taxes,&$count_of_func) use($cat_posts,&$_tax_add) {
        $count_of_func = 0;
        $s = '';
        foreach($taxes as $t) {
            $post_count_of_cat = $cat_posts[$t->id] ?? 0;

            $s1 = '<li data-cid="'.$t->id.'" class="folder"><i class="folder-name fa fa-folder-o"></i><span class="folder-name">'.htmlspecialchars($t->name).'(';
            $s2 = ')</span><ul>';
            $s3 = '';

            $child_count_of_func = 0;

            if(isset($t->sons))
                $s3 = $_tax_add($t->sons, $child_count_of_func);

            $s4 = '</ul></li>';

            $s .= $s1.$post_count_of_cat.(isset($t->sons) ? '/'.($post_count_of_cat+$child_count_of_func) : '').$s2.$s3.$s4;

            $count_of_func += $post_count_of_cat + $child_count_of_func;
        }
        return $s;
    };

    echo '<ul class="roots">',$_tax_add($taxes, $count_of_total/*not used*/),'</ul>';
}

function tb_head() {
?>
<style>
.archives-view {
    display: flex;
    height: 100%;
}
.cats {
    flex: 1;
    max-width: 25%;
    border-right: 1px solid #efefef;
    height: 100%;
    overflow: auto;
}
.folder > ul {
    display: none;
}
.framebox {
    flex: 1;
    margin: 1em;
    position:relative;
}
.framebox.fullscreen {
    position: fixed;
    left:0;
    top:0;
    width:100%;
    height:100%;
    margin: 0;
    background:white;
}
#frame {
    width: 100%;
    height: 100%;
}
#full {
    position: absolute;
    right: 1em;
    top:1em;
    opacity: 0.3;
}
#full:hover {
    opacity:1;
}
#content {
    height: 100%;
    overflow: auto;
}

.roots {
    list-style: none;
    padding: 0;
}

</style>
<script>
function resize() {
    var header = $('#header');
    var main = $('#main');
    main.css('height', $(window).height() + 'px');
}

$(window).on('load', resize);
$(window).on('resize', resize);
</script>
<?php }

add_hook('tb_head', 'tb_head');
require('header.php');
?>
<div class="archives-view" >
    <div class="cats">
        <?php list_all_cats(); ?>
    </div>
    <div class="framebox">
        <button id="full">fullscreen</button>
        <iframe id="frame" frameborder="0"></iframe>
        <script>
        $('#full').on('click',function(){
            $('.framebox').toggleClass('fullscreen');
        });
        </script>
    </div>
</div>

<?php

function tb_footer_hook() { ?>
<script>
function gen_entry(p) {
    return $('<li/>')
        .attr('class', 'title')
        .append($('<span/>')
            .attr('class', 'item')
            .attr('data-id', p.id)
            .attr('title', p.title)
            .text(p.title)
        );
}

function get_entries_callback(data, ul) {
    if(data.code == 0) {
        var ps = data.data;
        for(var i=0; i< ps.length; i++) {
            ul.append(gen_entry(ps[i]));
        }
        if(ps.length == 0) {
            ul.append($('<li/>')
                .attr('class', 'none')
                .text('（没有文章）')
            );
        }
    }
    else {
        alert(data.msg);
    }
}

function toggle_loading(ul, on) {
    return on
        ? ul.append($('<li/>')
            .attr('class', 'loading')
            .append($('<i/>')
                .attr('class', 'fa fa-cog fa-spin')
                )
            .append($('<span/>').text(' 正在加载...'))
            )
        : ul.find('li.loading').remove()
        ;
}

$('.cats').on('click',function(e) {
    var t = $(e.target);
    if(t.hasClass('folder-name')) {
        var li = t.parent();
        var ul = li.find('>ul');
        var fa = li.find('>.folder-name.fa');
        ul.toggle();
        fa.toggleClass('fa-folder-open-o');
        fa.toggleClass('fa-folder-o');
        if(li.attr('data-clicked') != '1') {
            li.attr('data-clicked', '1');
            toggle_loading(ul, true);
            var cid = li.attr('data-cid');
            $.get('/api/post/get_cat_posts',
                {
                    cid: cid,
                },
                function(data) {
                    get_entries_callback(data, ul);
                }
            ).always(function() {
                toggle_loading(ul, false);
            });
        }
    }
    else if(t.hasClass('item')) {
        var id = t.attr('data-id');
        $('#frame').attr('src', '/' + id + '/?content_only=1');
    }
    e.stopPropagation();
});
</script>
<?php
}

add_hook('tb_footer', 'tb_footer_hook');
require('footer.php');

