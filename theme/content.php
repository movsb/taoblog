<?php
/**
 * Post content template
 */
?>
<article class="post" itemscope itemtype="http://schema.org/Article">
    <div class="title clearfix">
        <h1 itemprop="name"><?php echo htmlspecialchars($the->title); ?></h1>
    </div>

    <?php if ($logged_in) { ?>
    <div class="toolbar"><span><?php echo the_edit_link($the);?></span></div>
    <?php } ?>

    <div class="entry clearfix" itemprop="articleBody">
        <?php
            $is_preview = ($_POST['do'] ?? '') === 'preview';
            echo $is_preview ? $_POST['content'] : $the->content;
        ?>
        <?php if ($the->type == 'post') { ?>
        <div class="meta clearfix">
            <p><span class="item author" itemprop="author" itemscope itemtype="http://schema.org/Person">
                <span itemprop="name"><?php echo htmlspecialchars($tbopt->get('author')); ?></span>
            </span>
            发表于：<?php echo the_meta_date();?>，阅读量：<?php echo $the->page_view; ?><?php
            $s = the_meta_tag();
            if ($s != '') {
                echo '，标签：<span itemprop="keywords">', $s, '</span>';
            }
            ?></p>
            <p>版本声明：若非特别注明，本站所有文章均为作者原创，转载请务必注明原文地址。</p>
        </div>
        <?php } ?>
    </div><!-- end entry -->

    <div class="related clearfix">
        <?php theRelatedPosts(); ?>
    </div><!-- end related -->

    <!-- comments begin -->
    <div id="comments" class="clearfix">
        <script src="/theme/scripts/comment.js"></script>
    </div>
    <!-- comments end -->
</article>
