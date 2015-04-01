<article class="post">	
	<div class="title">
		<h1>
		<div id="closebtn" style="background-color: purple; width: 1.2em; height: 1.2em; cursor: pointer; float: left; margin-right: 0.8em; margin-top: 0.2em;">
		</div>
		<a style="color: #234E40;" href="<?php echo $tbopt->get('home').'/archives/'.$the->id.'.html'; ?>" title="<?php echo $the->title; ?>"><?php echo $the->title; ?></a></h1>
		<script>
			$('#closebtn').click(function(){
				var p = $('#panel');
				if(p.css('width') == '0px') {
					p.css('width', '250px').css('opacity', '1');
					$('body').css('margin-left', '250px');
				} else {
					$('#panel').css('width', '0px').css('opacity', '0');
					$('body').css('margin-left', '0px');
				}
			});
		</script>
	</div>

	<div class="meta">
	</div>

	<div class="entry">
		<?php echo $the->content; ?>
	</div>
</article>

