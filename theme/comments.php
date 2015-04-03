<div id="comments">
	<!--评论标题 -->
	<h3 id="comment-title">
		评论(<span class="loaded">0</span>/<span class="total">0</span>)
	</h3>

	<!-- 评论列表  -->
	<ol id="comment-list">

	</ol>

	<!-- 评论功能区  -->
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
			<input type="hidden" id="post-id" name="post-id" value="<?php echo $the->id; ?>" />
		</span>
	</div>

	<!-- 评论框 -->
	<?php if(1>0) : ?>
	<div id="comment-form-div" class="normal">
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

			<form id="comment-form" action="/admin/comment.php">
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
						<input type="hidden" name="do" value="post-cmt" />
					</div>
				</div>

				<div class="comment-content">
					<label style="position: absolute;">评论</label>
					<label style="visibility: hidden;">评论</label> <!-- 别问我，我想静静 -->
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
		<div class="comment-form-div-2">
			<div class="toolbar" style="color: #C8C8C8; font-size: 24px;">
				<span>评论: </span>
				<div style="position: relative; text-align: right; right: 0px; top: -40px;">
					<span>字体: </span>
					<span class="font-dec" title="减小字号"><i class="fa fa-minus"></i></span>
					<span class="font-inc" title="增大字号"><i class="fa fa-plus"></i></span>&nbsp;
					<span class="close" title="还原"><i class="fa fa-times"></i></span>
				</div>
			</div>
			<div class="dummy-textarea">
				<textarea id="comment-content-2" wrap="off"></textarea>
			</div>
		</div>
		<!--div class="comment-form-div-data" style="display: none;">
			<input type="hidden" name="maximized" value="false" />
		</div-->
	</div>
	<script src="/theme/scripts/comment.js"></script>
	<?php endif; ?>
</div>

