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
     * 取得属于某个祖先下的所有子评论
     *
     * @param (int) $p 祖先评论的编号
     *
     * @return 成功返回子评论一维数组（即：无层次关系），失败返回 false
     *
     * @fixme 修正：内部失败时可能返回 false
     */
    public function &get_children($p)
    {
        global $tbdb;
        global $tbdate;

        $sql = "SELECT * FROM comments WHERE ancestor=".(int)$p;
        $result = $tbdb->query($sql);
        if (!$result) {
            return false;
        }

        $children = [];
        while ($obj = $result->fetch_object()) {
            $obj->date = $tbdate->mysql_datetime_to_local($obj->date);
            $children[] = $obj;
        }

        $result->free();

        return $children;
    }

    /** Gone
     * 获取某条特定的评论或某篇文章的评论
     *
     * 指定评论编号时，以评论编号为准。否则以文章编号为准。两者不同时使用
     * 会同时获取每条评论的所有子评论，无层次关系，需要自行处理层次关系
     *
     * @param array $arg 字段集
     *                   可能使用的字段：
     *                   id      某个特定编号的评论
     *                   offset  指定获取评论的偏移
     *                   count   指定要获取多少个一级评论
     *                   post_id 指定要获取哪篇文章的评论
     *                   order   结果编号排序：asc 为 升序，其它为降序
     *
     * @return 成功返回获取到的评论，失败返回 false
     *
     * @fixme 根据函数返回值类型，不能返回 false
     */
    public function &get(&$arg=[])
    {
        global $tbdb;
        global $tbdate;

        $defs = [
            'id'        => 0,
            'offset'    => -1,
            'count'     => -1,
            'post_id'   => 0,
            'order'     => 'asc',
            ];

        $arg = tb_parse_args($defs, $arg);

        $id = (int)$arg['id'];
        if ($id > 0) {
            $sql = "SELECT * FROM comments WHERE id=$id";
        } else {
            $sql = "SELECT * FROM comments WHERE parent=0 ";
            if ((int)$arg['post_id'] > 0) {
                $sql .= " AND post_id=".(int)$arg['post_id'];
            }

            $sql .= " ORDER BY id ". (strtolower($arg['order'])==='asc' ? 'ASC' : 'DESC');

            $count = (int)$arg['count'];
            $offset = (int)$arg['offset'];
            if ($count > 0) {
                if ($offset >= 0) {
                    $sql .= " LIMIT $offset,$count";
                } else {
                    $sql .= " LIMIT $count";
                }
            }
        }

        $result = $tbdb->query($sql);
        if (!$result) {
            return false;
        }

        $cmts = [];
        while ($obj = $result->fetch_object()) {
            $obj->date = $tbdate->mysql_datetime_to_local($obj->date);
            $cmts[] = $obj;
        }

        if (!($id > 0)) {
            for ($i=0; $i<count($cmts); $i++) {
                $cmts[$i]->children = $this->get_children($cmts[$i]->id);
            }
        }

        $result->free();

        return $cmts;
    }

    /* 内部使用，暂不说明 */
    public function &get_vars($fields, $where)
    {
        global $tbdb;

        $sql = "SELECT $fields FROM comments WHERE $where LIMIT 1";
        $rows = $tbdb->query($sql);
        if (!$rows) {
            $this->error = $tbdb->error;
            return false;
        }

        if(!$rows->num_rows) return null;

        $r = $rows->fetch_object();
        return $r;
    }

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

    /** Gone
     * 获取所有文章（包括未公开发表的）的评论总数
     *
     * @return 若成功，返回评论总数。若失败，返回 0。
     */
    public function get_count_of_comments()
    {
        global $tbdb;

        $sql = "SELECT count(*) as size FROM comments";
        $rows = $tbdb->query($sql);
        return $rows ? $rows->fetch_object()->size : 0;
    }
}