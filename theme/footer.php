		</div><!-- content -->	
	</section>
</div><!-- wrapper -->

<?php apply_hooks('tb_footer'); ?>

<div style="display: none;">
	<script src="http://js.users.51.la/17768957.js"></script>
</div>
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
			$('html').animate({
				scrollTop: 0
			}, 300);
		});
	</script>
</div>
</body>
</html>

