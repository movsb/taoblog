<?php

/**
 * 数据库评论模块
 *
 * 该模块用来管理数据库评论表
 */

class TB_Comments
{
    public $error = ''; // 用来保存操作过程中的错误消息

    /**
     * 向表中插入一条评论内容
     *
     * @param array $arg 评论内容数组
     *
     * @return 如果插入成功，则返回数值型的键ID值。
     *         否则返回 false，$this->error 可以取得错误原因。
     */
    public function insert(&$arg)
    {
        global $tbdb;
        global $tbpost;
        global $tbdate;
        global $tbopt;
        global $logged_in;

        $def = [
            'post_id'   => false,
            'author'    => '',
            'email'     => '',
            'url'       => '',
            'ip'        => $_SERVER['REMOTE_ADDR'],
            'date'      => $tbdate->mysql_datetime_local(),
            'content'   => '',
            'status'    => 'public',
            'parent'    => false,
        ];

        $arg = tb_parse_args($def, $arg);

        /* < 表单数据验证与过滤 */

        foreach (['author', 'email', 'url', 'content'] as &$f) {
            $arg[$f] = trim($arg[$f]);
        }

        if (!$arg['post_id']) {
            $this->error = '文章编号不能少！';
            return false;
        }

        if (!$arg['author']) {
            $this->error = '昵称不能为空！';
            return false;
        }

        if (filter_var($arg['email'], FILTER_VALIDATE_EMAIL) === false) {
            $this->error = '邮箱地址不规范！';
            return false;
        }

        if ($arg['url'] && filter_var($arg['url'], FILTER_VALIDATE_URL) === false) {
            $this->error = '网址不规范！';
            return false;
        }

        if (!$arg['content']) {
            $this->error = '评论内容不能为空!';
            return false;
        }

        // 未登录时不能随意填写昵称或邮箱地址（以防可能的冒充作者
        // 这两个变量的值保存在设置表中，细节请参考 /doc/设置字段
        if (!$logged_in) {
            $not_allowed_authors = explode(',', $tbopt->get('not_allowed_authors').','.$tbopt->get('nickname'));
            foreach ($not_allowed_authors as &$author) {
                if (strncmp($author, '/', 1)==0 && preg_match($author.'i', $arg['author']) || strcasecmp($author, $arg['author'])==0) {
                    $this->error = '您不应该使用此昵称。';
                    return false;
                }
            }

            $not_allowed_emails = explode(',', $tbopt->get('not_allowed_emails').','.$tbopt->get('email'));
            foreach ($not_allowed_emails as &$email) {
                if (strncmp($email, '/', 1)==0 && preg_match($email.'i', $arg['email']) || strcasecmp($email, $arg['email'])==0) {
                    $this->error = '您不应该使用此邮箱地址。';
                    return false;
                }
            }
        }

        if (!$tbpost->have((int)$arg['post_id'])) {
            $this->error = '此评论对应的文章不存在！';
            return false;
        }

        // 评论的祖先ancestor字段只能系统设置
        if (($arg['ancestor'] = $this->get_ancestor((int)$arg['parent'], true)) === false) {
            return false;
        }

        $arg['date_gmt'] = $tbdate->mysql_local_to_gmt($arg['date']);

        /* > 表单数据验证与过滤 */

        // 向数据库中写入评论
        $sql = "INSERT INTO comments (
            post_id,author,email,url,ip,date,content,status,parent,ancestor)
            VALUES (?,?,?,?,?,?,?,?,?,?)";
        if ($stmt = $tbdb->prepare($sql)) {
            if ($stmt->bind_param(
                'isssssssii',
                $arg['post_id'],    $arg['author'], $arg['email'],
                $arg['url'],        $arg['ip'],     $arg['date_gmt'],
                $arg['content'],    $arg['status'],
                $arg['parent'],     $arg['ancestor']
            )
            ) {
                $r = $stmt->execute();
                $this->error = $stmt->error;
                $stmt->close();

                if ($r) {
                    $id = $tbdb->insert_id;
                    return $id;
                }
            }
        }

        $this->error = $tbdb->error;
        return false;
    }

    /**
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

    /**
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
    public function &get(&$arg=[], bool $pub)
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
            if($pub) {
                $sql .= " AND status='public'";
            }
        } else {
            $sql = "SELECT * FROM comments WHERE parent=0 ";
            if($pub) {
                $sql .= " AND status='public'";
            }
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

    /**
     * 获取某条评论的祖先评论的编号
     *
     * @param (int)  $id                     评论编号
     * @param (bool) $return_this_id_if_zero 如果祖先不存在是返回 $id 还是 0。
     *
     * @return 成功返回编号，失败返回 false
     */
    public function get_ancestor($id, $return_this_id_if_zero=false)
    {
        global $tbdb;

        if ((int)$id == 0) {
            return 0;
        }

        $sql = "SELECT ancestor FROM comments WHERE id=".(int)$id;
        $sql .= " LIMIT 1";
        $rows = $tbdb->query($sql);

        if (!$rows) {
            $this->error = $tbdb->error;
            return false;
        }

        if (!$rows->num_rows) {
            $this->error = '待查询祖先的评论的ID不存在！';
            return false;
        }

        $ancestor = $rows->fetch_object()->ancestor;
        return $ancestor
            ? $ancestor
            : ($return_this_id_if_zero
                ? $id
                : 0
                )
            ;
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

    /**
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

    /**
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
