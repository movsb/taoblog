<div id="comments">
	<h3 id="comment-title">
		评论(0/0)
	</h3>

	<ol id="comment-list">

	</ol>
	<div class="comment-func">
		<span id="post-comment">
			<span class="post-comment">发表评论</span>
		</span>
		<span id="load-comments">
			<span class="load">加载评论</span>
			<span class="loading">
				<i class="fa fa-spin fa-spinner"></i> 
				<span> 加载中...</span>
			</span>
			<span class="none">
				<i class="fa fa-info-circle"></i>
				<span> 没有了！</span>
			</span>
			<input type="hidden" id="post_id" name="post_id" value="<?php echo $post->ID; ?>" />
		</span>
	</div>
	<?php if(is_singular()) : ?>
	<div id="comment-form-div">
		<div class="comment-form-div-1">
			<div class="closebtn" title="关闭">
				<i class="fa fa-times"></i>
			</div>
			<div class="maxbtn" title="最大化">
				<i class="fa fa-plus"></i>
			</div>

			<div class="comment-title">
				<span>评论</span>
			</div>

			<form id="comment-form" action="/wp-comments-post.php">
				<div class="fields">
					<div class="field">
						<label>昵称</label>
						<input type="text" name="author" />
						<span class="needed">必填</span>
					</div>
					<div class="field">
						<label>邮箱</label>
						<input type="text" name="email" />
						<span class="needed">必填</span>
					</div>
					<div class="field">
						<label>网址</label>
						<input type="text" name="url" />
					</div>
					<div style="display: none;">
						<input id="comment-form-post-id" type="hidden" name="post_id" value="" />
						<input id="comment-form-parent"  type="hidden" name="parent" value="" />
					</div>
				</div>

				<div class="comment-content">
					<label style="position: absolute;">评论</label>
					<label style="visibility: hidden;">评论</label>
					<textarea id="comment-content" name="content" wrap="off"></textarea>
				</div>

				<div class="comment-submit">
					<span id="comment-submit">发表评论</span>
					<span class="submitting">
						<i class="fa fa-spin fa-spinner"></i>
						<span> 正在提交...</span>
					</span>'
					<span class="succeeded">
						<i class="fa fa-mr fa-info-circle"></i>
						<span></span>
					</span>
					<span class="failed">
						<i class="fa fa-mr fa-info-circle"></i>
						<span></span>
					</span>
				</div>
			</form>
		</div>
		<textarea id="comment-content-2" wrap="off"></textarea>
		<div class="comment-form-div-data" style="display: none;">
			<input type="hidden" name="maximized" value="false" />
		</div>
	</div>
	<?php endif; ?>
</div>

