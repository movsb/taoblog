		</div><!-- content -->	
	</section>
	<div id="mobile-footer">
		<span>&copy; <?php echo date('Y'),' ',$tbopt->get('author'); ?></span>
	</div>
</div><!-- wrapper -->

<?php apply_hooks('tb_footer'); ?>

<div class="tips" id="tips">
	<span></span>
	<script>
		var tmr = null;

		function show_tips(ht) {
			$('#tips span').html(ht);
			$('#tips').fadeIn(300);

			if(tmr) clearTimeout(tmr);

			tmr = setTimeout(function() {
					$('#tips').fadeOut(300);
					tmr = null;
				},
				3000
				);
		}
	</script>
</div>

<div class="footer-toolbar" id="footer-toolbar">
	<div class="back-to-top" id="back-to-top" title="回到顶端">
		<i class="fa fa-arrow-circle-up"></i>
		<script>
			window.onscroll = function() {
				if(window.scrollY > 160) {
					$("#back-to-top").fadeIn(500);
				} else {
					$("#back-to-top").fadeOut(500);
				}
			};
			$('#back-to-top').click(function(){
				$('html,body').animate({
					scrollTop: 0
				}, 300);
			});
		</script>
	</div>
	<div class="reading-mode no-sel" id="reading-mode" title="阅读模式">
		<i class="fa fa-plus-circle"></i>
		<script>
			var in_reading_mode = false;

			function toggle_reading_mode() {
				var header = $('#header');
				var main = $('#main');
				var icon = $('#reading-mode i');

				if(!in_reading_mode) {
					header.css('left', '-300px');
					main.css('margin-left', '0px');
					icon.css('color', '#f66');
				} else {
					header.css('left', '0px');
					main.css('margin-left', '300px');
					icon.css('color', 'inherit');
				}

				in_reading_mode = !in_reading_mode;

				show_tips(in_reading_mode 
					? '<b>已进入阅读模式。</b><br/>您可以点击右下角的“+”号退出阅读模式。'
					: '已退出阅读模式。');
			}

			$('#reading-mode').click(function() {
				toggle_reading_mode();
			});
		</script>
	</div>
</div>
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
</body>
</html>
<!-- 执行时间: <?php echo $execution_time; ?>s -->

