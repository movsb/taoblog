<?php

defined('TBPATH') or die("Silence is golden.");

add_hook('admin_left', 'admin_home_page');

function admin_home_page() {
?><li>
	<a href="/">
	<div class="menu">
		首页
	</div>
	</a>
</li>
<?php 
}

add_hook('admin_left', 'shuoshuo_admin_left');

function shuoshuo_admin_left() {
?><li>
    <a href="shuoshuo.php">
    <div class="menu">
        发表说说
    </div>
    </a>
</li>
<?php
}

add_hook('admin_left', 'post_admin_left');

function post_admin_left() {
?><li>
	<a href="post.php">
	<div class="menu">
		发表文章
	</div>
	</a>
</li>
<?php 
}

add_hook('admin_left', 'page_admin_left');

function page_admin_left() {
?><li>
	<a href="post.php?type=page">
	<div class="menu">
		发表页面
	</div>
	</a>
</li>
<?php 
}

add_hook('admin_left', 'tax_admin_left');

function tax_admin_left() {
?><li>
	<a href="taxonomy.php">
	<div class="menu">
		分类管理
	</div>
	</a>
</li>
<?php 
}

