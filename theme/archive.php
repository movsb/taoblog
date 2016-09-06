<?php

function list_all_dates() {
    global $tbpost;
    $dd = $tbpost->get_date_archives();

    echo '<ul class="roots">';
    foreach(array_reverse($dd, true) as $yy => &$ya) {
        foreach(array_reverse($ya, true) as $mm => $n) {
            echo '<li class="year-month" data-yy="',$yy,'" data-mm="',$mm,'">';
            echo    '<i class="datetime fa fa-clock-o"></i>';
            echo    '<span class="datetime">',$yy,'年',($mm<10?'0':''),$mm,'月(',$n,')</span>';
            echo    '<ul></ul>';
            echo '</li>';
        }
    }
    echo '</ul>';
}

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

function list_all_tags() {
    global $tbtag;
    $tags = $tbtag->list_all_tags(30);

    echo '<ul class="roots">';
    foreach($tags as &$t) {
        echo '<li class="tag" data-name="', htmlspecialchars($t->name),'">';
        echo    '<i class="tag-name fa fa-tag"></i>';
        echo    '<span class="tag-name">',htmlspecialchars($t->name),'(',$t->size,')</span>';
        echo    '<ul></ul>';
        echo '</li>';
    }
    echo '</ul>';
}

function tb_head_hook() { ?>
<style>
    .archives a {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        display: inline-block;
        max-width: 100%;
    }
</style>
<?php }
add_hook('tb_head', 'tb_head_hook');

require('header.php');
?>
<div class="archives">
    <div class="tags"><h2>标签</h2><?php list_all_tags(); ?></div>
    <div class="cats"><h2>分类</h2><?php list_all_cats(); ?></div>
    <div class="date"><h2>日期</h2><?php list_all_dates(); ?></div>
</div>

<?php

function tb_footer_hook() { ?>
<script>
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
                ul.append('<li class="loading"><i class="fa fa-cog fa-spin"></i>&nbsp;正在加载...</li>');
                var cid = li.attr('data-cid');
                $.get('/admin/ajax.php',
                    'do=get_cat_posts&cid=' + cid,
                    function(data) {
                        if(data.errno == 'ok') {
                            var ps = data.posts;
                            for(var i=0; i< ps.length; i++) {
                                var p = ps[i];
                                var s = '<li class="title"><a target="_blank" href="/' + p.id + '/" title="'+p.title+'">'+ p.title + '</a></li>';
                                ul.append(s);
                            }
                            if(ps.length == 0)
                                ul.append('<li class="none">（没有文章）</li>');
                        }
                        else {
                            alert(data.error);
                        }
                    }
                ).always(function() {
                    ul.find('li.loading').remove();
                });
            }
        }
        e.stopPropagation();
    });
    $('.date').on('click',function(e) {
        var t = $(e.target);
        if(t.hasClass('datetime')) {
            var li = t.parent();
            var ul = li.find('>ul');
            ul.toggle();
            if(li.attr('data-clicked') != '1') {
                li.attr('data-clicked', '1');
                ul.append('<li class="loading"><i class="fa fa-cog fa-spin"></i>&nbsp;正在加载...</li>');
                var yy = li.attr('data-yy');
                var mm = li.attr('data-mm');
                $.get('/admin/ajax.php',
                    'do=get_date_posts&yy=' + yy + '&mm=' + mm,
                    function(data) {
                        if(data.errno == 'ok') {
                            var ps = data.posts;
                            for(var i=0; i< ps.length; i++) {
                                var p = ps[i];
                                var s = '<li class="title"><a target="_blank" href="/' + p.id + '/" title="'+p.title+'">'+ p.title + '</a></li>';
                                ul.append(s);
                            }
                            if(ps.length == 0)
                                ul.append('<li class="none">（没有文章）</li>');
                        }
                        else {
                            alert(data.error);
                        }
                    }
                ).always(function() {
                    ul.find('li.loading').remove();
                });
            }
        }
        e.stopPropagation();
    });
    $('.tags').on('click',function(e) {
        var t = $(e.target);
        if(t.hasClass('tag-name')) {
            var li = t.parent();
            var ul = li.find('>ul');
            ul.toggle();
            if(li.attr('data-clicked') != '1') {
                li.attr('data-clicked', '1');
                ul.append('<li class="loading"><i class="fa fa-cog fa-spin"></i>&nbsp;正在加载...</li>');
                var name = li.attr('data-name');
                $.get('/admin/ajax.php',
                    'do=get_tag_posts&tag=' + name,
                    function(data) {
                        if(data.errno == 'ok') {
                            var ps = data.posts;
                            for(var i=0; i< ps.length; i++) {
                                var p = ps[i];
                                var s = '<li class="title"><a target="_blank" href="/' + p.id + '/" title="'+p.title+'">'+ p.title + '</a></li>';
                                ul.append(s);
                            }
                            if(ps.length == 0)
                                ul.append('<li class="none">（没有文章）</li>');
                        }
                        else {
                            alert(data.error);
                        }
                    }
                ).always(function() {
                    ul.find('li.loading').remove();
                });
            }
        }
        e.stopPropagation();
    });
</script>
<?php
}

add_hook('tb_footer', 'tb_footer_hook');
require('footer.php');

