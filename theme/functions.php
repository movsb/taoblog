<?php

function the_meta_date() {
	global $the;

	$dd = preg_split('/-/', preg_split('/ /', $the->date)[0]);;
    $tt = sprintf('%d年%d月%d日', $dd[0], $dd[1], $dd[2]);

	$DD = preg_split('/-/', preg_split('/ /', $the->modified)[0]);;
    $TT = sprintf('%d年%d月%d日', $DD[0], $DD[1], $DD[2]);

	return '<span class="value" title="发表时间：'.$tt."\n".'修改时间：'.$TT.'">'.$tt.'</span>';
}

function the_meta_tag() {
	global $the;

	$tags = &$the->tag_names;
	$as = [];

	foreach($tags as &$t) {
		$as[] = '<a href="/tags/'.htmlspecialchars(urlencode($t)).'">'.htmlspecialchars($t).'</a>';
	}

    $ts = join(' · ', $as);

    return $ts ? '<span class="value">'.$ts.'</span>' : '';
}

/**
 * Outputs the related posts
 *
 * @return the html content of related posts list
 */
function theRelatedPosts()
{
    global $the, $tbquery;

    if ($the->type == 'post'
        && isset($tbquery->related_posts)
        && is_array($tbquery->related_posts)
        && count($tbquery->related_posts)
    ) {
        echo '<h3>相关文章</h3>', PHP_EOL;
        echo '<ol>',PHP_EOL;

        $ps = &$tbquery->related_posts;
        foreach ($ps as $p) {
            echo sprintf("<li><a href=\"/%d/\">%s</a></li>\n", $p->id, htmlspecialchars($p->title));
        }

        echo '</ol>',PHP_EOL;
    }
}
