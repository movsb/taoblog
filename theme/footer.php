		</div><!-- content -->	
	</section>
	<div class="footer" id="footer">
		<div class="footer-wrapper">
			<div class="avatar">
				<img class="me" src="/theme/images/me.png" />
				<div class="copyright mb">
					<span>&copy; <?php echo date('Y'),' ',$tbopt->get('nickname'); ?></span>
				</div>
				<div class="declaration mb">
					<span>若非特别注明，本站所有内容均为原创，转载请注明出处。</span>
				</div>
			</div>
			<div class="column about">
				<h3>ABOUT</h3>
				<ul>
					<li><a target="_blank" href="/blog">关于博客</a></li>
					<li><a target="_blank" href="/echo">建议反馈</a></li>
					<li><a target="_blank" href="/about">关于我</a></li>
				</ul>
			</div>
			<div class="column links">
				<h3>LINKS</h3>
				<ul>
					<li><a title="自学去 - 一个免费的自学网站~" target="_blank" href="http://www.zixue7.com">自学去</a></li>
					<li><a title="小谢的博客" target="_blank" href="http://memorycat.com">小写adc</a></li>
					<li><a title="网事如风的博客" target="_blank" href="http://godebug.org">网事如风</a></li>
					<li><a title="道不行-技术、文学、评论、吐槽" target="_blank" href="http://www.daobuxing.com">道不行</a></li>
				</ul>
			</div>
			<div class="column blog">
				<h3>BLOG</h3>
				<ul>
					<li><a>日志归档</a></li>
					<li><a>日志分类</a></li>
					<li><a target="_blank" href="<?php echo $tbopt->get('home'),'/rss'; ?>">订阅博客</a></li>
				</ul>
			</div>
		</div>
	</div>
</div><!-- wrapper -->

<?php apply_hooks('tb_footer'); ?>

<div style="display: none;">
	<script src="http://js.users.51.la/17768957.js"></script>
</div>
</body>
</html>

