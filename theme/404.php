<?php header('HTTP/1.1 404 Not Found');
$tbquery->push_404();
require('header.php');
$tbquery->pop_404();
?>
	<div class="err-404">		
		<div style="text-align: center; padding: 4em;">
			<span style="font-size: 2em;"><?php
				if($tbquery->is_page()) {
					echo '此页面不存在。';
				} else if($tbquery->is_post()) {
					echo '此文章不存在。';
				} else if($tbquery->is_category()) {
                    echo '此分类下不存在相关文章。';
				} else if($tbquery->is_date()) {
                    echo '此时段不存在相关文章。';
				} else if($tbquery->is_tag()) {
                    echo '此标签下不存在相关文章。';
				}
			?></span>
		</div>
	</div>
<?php
$tbquery->push_404();
require('footer.php');
$tbquery->pop_404();

