<?php

/**
 * 数据库评论模块
 *
 * 该模块用来管理数据库评论表
 */

class TB_Comments
{
    public $error = ''; // 用来保存操作过程中的错误消息

    /** Gone
     * 获取近期评论（无层次关系）
     *
     * 条数为 10 条，暂时写死了
     *
     * @return 近期评论数组。若失败，返回空数组。
     *
     * @fixme 把默认获取条数写进数据库设置中
     */
    public function &get_recent_comments()
    {
        global $tbdb;

        $cmts = [];

        $sql = "SELECT * FROM comments ORDER BY date DESC LIMIT 10";
        $rows = $tbdb->query($sql);
        if (!$rows) {
            return $cmts;
        }

        while ($r = $rows->fetch_object()) {
            $cmts[] = $r;
        }

        return $cmts;
    }
}