<?php

if($_SERVER['REQUEST_METHOD'] === 'GET') :

require_once('admin.php');

function memory_admin_head()
{?>
<script src="scripts/marked.js"></script>
<script>
    var renderer = new marked.Renderer();
    // renderer.code = function(code, lang) {
    //     var beg = '<pre class="code" lang="' + (lang === undefined ? '' : lang) + '">\n';
    //     var text = code.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    //     var end = '\n</pre>';
    //     return beg + text + end;
    // }
    renderer.hr = function() {
        return '<hr/>';
    }
    renderer.br = function() {
        return '<br/>';
    }
    marked.setOptions({renderer: renderer});
</script>
<?php
    apply_hooks('admin:memory:head');
}

add_hook('admin_head', 'memory_admin_head');

function memory_admin_foot() {
    apply_hooks('admin:memory:foot');
}

add_hook('admin_footer', 'memory_admin_foot');

admin_header();

$id = (int)($_GET['id'] ?? 0);
$content = '';
$source = '';
$geo_lat = 0;
$geo_lng = 0;
$geo_addr= '';
$date    = '';

if($id > 0) {
    $item = $tbshuoshuo->get($id);
    $content = $item->content;
    $source  = $item->source;
    $geo_lat = $item->geo_lat;
    $geo_lng = $item->geo_lng;
    $geo_addr= $item->geo_addr;
    $date    = $item->date;
}
?>
<form id="form" method="post" style="margin-bottom: 2em;">
<h2>发表说说</h2>
<textarea id="source" name="source" style="display: block; min-width: 300px; min-height: 150px;"><?php echo htmlspecialchars($source ? $source : $content);?></textarea>
<input type="hidden" name="content" value="" />
<p>时间：<input type="text" name="date" value="<?php echo $date; ?>" /></p>
<p>坐标： <span class="position"><?php echo $geo_addr, '（', $geo_lat,',',$geo_lng, '）'; ?></span></p>
<p>位置：
    <select name="addr">
        <option value="<?php echo htmlspecialchars($geo_addr);?>"><?php echo  htmlspecialchars($geo_addr);?></option>
    </select>
</p>
    <input type="hidden" name="lat" value="<?php echo $geo_lat;?>" />
    <input type="hidden" name="lng" value="<?php echo $geo_lng;?>" />
    <script>
        function update_address_baidu(data) {
            var addr = $('#form select[name=addr]');
            addr.append('<option disabled value="null">---百度地图---</option>');
            update_address(data);
        }

        function update_address_tencent(data) {
            var addr = $('#form select[name=addr]');
            addr.append('<option disabled value="null">---腾讯地图---</option>');
            update_address(data);
        }

        function update_address_gaode(data) {
            var addr = $('#form select[name=addr]');
            addr.append('<option disabled value="null">---高德地图---</option>');
            update_address(data);
        }

        function update_address(data) {
            var coords = $('#form .position');
            var addr = $('#form select[name=addr]');

            if(data.status === 0 || data.status === '1') {
                var r = data.result || data.regeocode;
                (r.pois || []).forEach(function(poi) {
                    var all = poi.name || poi.title;
                    var opt = $('<option/>');
                    opt.attr({'value': all}).text(all);
                    addr.append(opt);
                });
            }
        }

        function update_address_gaode_coord(data) {
            if(data.status === '1') {
                var s = data.locations;
                var g = document.createElement('script');
                g.src = '//restapi.amap.com/v3/geocode/regeo?key=93ebbbb772a24c6625def52e779dc38b&location='+s+'&extensions=all&callback=update_address_gaode'
                $('#form').append(g);
            }
        }

        function get_detail_location(lat,lng) {
            var coords = $('#form .position');
            var b = document.createElement('script');
            b.src = '//api.map.baidu.com/geocoder/v2/?ak=uVQwCq4gtGAyQWmpDwLD7g9bGXRa6UFm&output=json&location='+lat+','+lng+'&coordtype=wgs84ll&pois=1&callback=update_address_baidu';
            $('#form').append(b);
            var q = document.createElement('script');
            q.src = '//apis.map.qq.com/ws/geocoder/v1/?key=3V3BZ-J3KRD-XLI4M-P4VGS-KITNH-7IBO2&location='+lat+','+lng+'&coord_type=1&get_poi=1&output=jsonp&callback=update_address_tencent';
            $('#form').append(q);
            var g = document.createElement('script');
            g.src = '//restapi.amap.com/v3/assistant/coordinate/convert?key=93ebbbb772a24c6625def52e779dc38b&locations='+lng+','+lat+'&coordsys=gps&output=json&callback=update_address_gaode_coord'
            $('#form').append(g);
        }

        <?php if ($geo_lat == 0 || $geo_lng == 0) : ?>
        if(navigator.geolocation) {
            $('#form .position').text('获取中……');
            navigator.geolocation.getCurrentPosition(function(position) {
                $('#form input[name=lat]').val(position.coords.latitude);
                $('#form input[name=lng]').val(position.coords.longitude);

                var latlng = position.coords.latitude + ',' + position.coords.longitude;
                $('#form .position').text('【已更新】' + latlng);
                get_detail_location(position.coords.latitude, position.coords.longitude);
            },
            function(err) {
                $('#form .position').text('未能获取到地址。');
            },
            {
                enableHighAccuracy: true,
            }
            );
        }
        else {
            $('#form .position').text('浏览器不支持位置服务。');
        }
        <?php else : ?>
            get_detail_location(<?php echo $geo_lat,',',$geo_lng; ?>);
        <?php endif; ?>
    </script>
<script>
$('#form').submit(function() {
    $('#form input[name="content"]').val(marked($('#form textarea[name="source"]').val()));
    return true;
});
</script>

<input type="hidden" name="do" value="<?php echo $id > 0 ? 'update' : 'new';?>" />
<input type="hidden" name="id" value="<?php echo $id;?>" />
<p><input type="submit" value="发表" /></p>
</form>

<h2>近期说说</h2>
<?php
    $sss = $tbshuoshuo->get_latest(10);
    if(count($sss) == 0) return false;

    echo '<ul id="shuoshuos" style="list-style: none;">';
    foreach($sss as &$ss) {
        echo '<li data-id="',$ss->id,'">',$ss->content,'<span>',$ss->date,'</span>';
        echo '<button class="edit">编辑</button><button class="delete">删除</button></li>';
    }
    echo '</ul>';
?>
<script>
    $('#shuoshuos').on('click', function(e) {
        var cls = e.target.classList;
        if(cls.contains('edit')) {
            var id = $(e.target.parentNode).attr('data-id');
            location.href = '/admin/shuoshuo.php?id=' + id;
            e.preventDefault();
            e.stopPropagation();
            return false;
        }
        else if(cls.contains('delete')) {
            var id = $(e.target.parentNode).attr('data-id');
            if(!confirm('确定删除吗？'))
                return false;

            $.post('/admin/shuoshuo.php',
                {
                    do: 'delete',
                    id: id,
                },
                function(data) {
                    if(data.errno == 'ok') {
                        $('#shuoshuos > li[data-id=' + id + ']').remove();
                    }
                    else {
                        alert(data.error);
                    }
                },
                'json'
            );
        }
    });
</script>

<?php
admin_footer();

die(0);

else : // POST

function shuoshuo_die_json($arg) {
    header('HTTP/1.1 200 OK');
    header('Content-Type: application/json');

    echo json_encode($arg, JSON_UNESCAPED_UNICODE);
    die(0);
}

require_once('login-auth.php');

function auth() {
    if(!login_auth()) {
        shuoshuo_die_json([
            'errno' => 'unauthorized',
            'error' => '需要登录后才能进行该操作！',
            ]);
    }
}

require_once('load.php');


$do = $_POST['do'] ?? '';

if($do == 'new') {
    auth();
    $r = $tbshuoshuo->post($_POST);
    if($r === false)
        shuoshuo_die_json([
            'errno' => 'failed',
            'error' => $tbshuoshuo->error,
            ]);

    header('HTTP/1.1 302 Found');
    header('Location: /admin/shuoshuo.php');
    die(0);
}
else if($do == 'update') {
    auth();
    $r = $tbshuoshuo->update((int)$_POST['id'], $_POST);
    if($r === false)
        shuoshuo_die_json([
            'errno' => 'failed',
            'error' => $tbshuoshuo->error,
            ]);

    header('HTTP/1.1 302 Found');
    header('Location: /admin/shuoshuo.php');
    die(0);
}
else if($do == 'delete') {
    auth();
    $r = $tbshuoshuo->del((int)$_POST['id']);
    if($r === false)
        shuoshuo_die_json([
            'errno' => 'failed',
            'error' => $tbshuoshuo->error,
            ]);

    else
        shuoshuo_die_json([
            'errno' => 'ok',
            ]);
    die(0);
}
else if($do == 'post-comment') {
    $r = $tbsscmt->post($_POST['sid'], $_POST['author'], $_POST['content']);
    if($r === false) {
        shuoshuo_die_json([
            'errno' => 'failed',
            'error' => $tbsscmt->error,
        ]);
    }
    else {
        // 当前不作任何处理，直接返回
        shuoshuo_die_json([
            'errno' => 'ok',
            'author' => htmlspecialchars($_POST['author']),
            'content' => htmlspecialchars($_POST['content']),
        ]);
    }
}

die(0);

endif;

