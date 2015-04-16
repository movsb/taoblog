		</div><!-- content -->	
	</section>
</div><!-- wrapper -->

<?php if($tbquery->is_singular()) {
		echo $snjs->tax->footer,$snjs->post->footer;
	}

	apply_hooks('tb_footer'); ?>

</body>
</html>

