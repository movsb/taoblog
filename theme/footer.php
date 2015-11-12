		</div><!-- content -->	
	</section>
	<div id="mobile-footer">
		<span>&copy; <?php echo date('Y'),' ',$tbopt->get('author'); ?></span>
	</div>
</div><!-- wrapper -->

<?php apply_hooks('tb_footer'); ?>

<div class="footer-toolbar" id="footer-toolbar">
	<div class="back-to-top" id="back-to-top" title="回到顶端">
		<i class="fa fa-arrow-circle-up"></i>
	</div>
	<div class="reading-mode no-sel" id="reading-mode" title="阅读模式">
		<i class="fa fa-plus-circle"></i>
	</div>
</div>
<div class="img-view" id="img-view"><img /><div class="tip"></div></div>
<div style="display: none;">
	<script src="http://js.users.51.la/17768957.js"></script>
	<script>
		(function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
		(i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
		m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
		})(window,document,'script','//www.google-analytics.com/analytics.js','ga');

		ga('create', 'UA-65174773-1', 'auto');
		ga('send', 'pageview');
	</script>
</div>
<script src="/theme/scripts/footer.js"></script>
</body>
</html>
<!-- 执行时间: <?php echo $execution_time; ?>s -->
<?php

