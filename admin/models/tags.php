<?php

/*
 *  数据库标签管理模块
 *
 *  该模块用来管理数据库中的标签表
 *
 */

class TB_Tags {
    /*
     *  插入一个标签
     *
     * @param string $name 标签的名字
     *
     * @return 若插入成功，则返回该名字标签对应的编号。否则，返回 false
     */  
    public function insert_tag($name) {
        global $tbdb;

        $sql = "INSERT INTO tags (name) values (?);";
        if($stmt = $tbdb->prepare($sql)){
            $stmt->bind_param('s', $name);
            if($stmt->execute())
                return $tbdb->insert_id;
        }

        return false;
    }

    /*
     * 根据标签名字来取得其编号
     *
     * 可以用来判断某个标签是否存在
     *
     * @param string $name 标签名
     *
     * @return 若标签存在，返回其编号。否则返回 false
     */
    public function get_tag_id($name) {
        global $tbdb;

        $sql = "SELECT id FROM tags WHERE name='".$tbdb->real_escape_string($name)."' LIMIT 1";
        if(!($results = $tbdb->query($sql)) || !$results->num_rows) {
            return false;
        }

        return $results->fetch_object()->id;
    }

    /**
     * 判断某个标签是否存在
     *
     * 该函数调用 get_tag_id 来判断某个名为 $name 的标签是否存在
     *
     * @param string $name 标签名
     *
     * @return 返回布尔值代表存在与否
     */
    public function has_tag_name($name) {
        return !!$this->get_tag_id($name);
    }

    /**
     * 取得某篇文章的标签名列表
     *
     * @param int $id 文章编号
     *
     * @return 若成功，返回文章拥有的标签名列表。若失败（如文章不存在），返回空列表。
     */
    public function &get_post_tag_names($id) {
        global $tbdb;

        $id = (int)$id;
        $sql = "SELECT tags.name FROM post_tags,tags WHERE post_tags.post_id=$id AND post_tags.tag_id=tags.id";

        $names = [];

        $results = $tbdb->query($sql);
        if(!$results) return $names;

        while($n = $results->fetch_object()) {
            $names[] = $n->name;
        }

        return $names;
    }

    /**
     * 取得某篇文章的标签编号列表
     *
     * @param int  $id    文章编号
     * @param bool $alias 是否同时返回别名标签
     *
     * @return 若成功，返回文章拥有的标签编号列表。若失败，返回空列表。
     */
    public function &get_post_tag_ids($id, bool $alias)
    {
        global $tbdb;

        $id = (int)$id;
        $sql = "SELECT tag_id FROM post_tags WHERE post_id=$id";

        $ids = [];

        $results = $tbdb->query($sql);
        if (!$results) {
            return $ids;
        }

        while ($n = $results->fetch_object()) {
            $ids[] = $n->tag_id;
        }

        if (!$alias) {
            return $ids;
        }

        return $this->getAliasTagsAll($ids);
    }

    /**
     * 取得包括别名在内的标签编号列表（没有递归）
     * 
     * @param array $ids 文章标签，不包含别名（即别名为0的标签）
     * 
     * @return ?
     */
    public function &getAliasTagsAll(array $ids)
    {
        global $tbdb;

        $sids = join(',', $ids);
        $sql = "SELECT alias FROM tags WHERE id in ($sids)";
        $rids = [];

        $results = $tbdb->query($sql);
        if (!$results) {
            return $ids;
        }

        while ($n = $results->fetch_object()) {
            $rids[] = $n->alias;
        }

        $ids = array_merge($ids, $rids);
        return $ids;
    }

    /*
     * 更新文章的标签名列表
     *
     * 更新时，会自动向数据库中写入新的标签和删除原来有但现在没有的标签（不删除标签本身）
     *
     * @param int $id       文章编号
     *
     * @param string $tags  标签列表（逗号分隔）
     *
     * @return 始终返回true
     */
    public function update_post_tags($id, $tags)
    {
        global $tbdb;

        $tags = is_string($tags) ? str_replace("，", ",", $tags) : "";

        $oldts = $this->get_post_tag_names($id);
        $newts = explode(',', $tags);

        $deleted = []; // 删除的
        $added = []; // 新增加的

        // 计算需要删除的
        foreach ($oldts as $o) {
            if (!in_array($o, $newts)) {
                $deleted[] = $o;
            }
        }

        // 计算需要增加的
        foreach ($newts as $n) {
            $n = trim($n);
            if ($n && !in_array($n, $oldts)) {
                $added[] = $n;
            }
        }

        // 删除需要删除的
        foreach ($deleted as $d) {
            $tid = $this->get_tag_id($d);
            $this->delete_post_tag($id, $tid);
        }

        // 增加需要增加的
        foreach ($added as $a) {
            if (!$this->has_tag_name($a)) {
                $tid = $this->insert_tag($a);
            } else {
                $tid = $this->get_tag_id($a);
            }

            $this->insert_post_tag($id, $tid);
        }

        return true;
    }

    /*
     * 插入一条文章标签记录
     *
     * @param int $post_id  文章编号
     *
     * @param int $tag_id   标签编号
     *
     * @return 插入的文章标签记录编号
     */
    public function insert_post_tag($post_id, $tag_id) {
        global $tbdb;

        $post_id = (int)$post_id;
        $tag_id = (int)$tag_id;

        $sql = "INSERT INTO post_tags (post_id,tag_id) values ($post_id,$tag_id)";
        $results = $tbdb->query($sql);
        if(!$results) return false;

        return $tbdb->insert_id;
    }

    /*
     * 删除一条文章标签记录
     *
     * @param int $post_id  文章编号
     *
     * @param int $tag_id   标签编号
     *
     * @return 若成功，返回 true。若失败，返回 false。
     */
    public function delete_post_tag($post_id, $tag_id) {
        global $tbdb;

        $post_id = (int)$post_id;
        $tag_id = (int)$tag_id;

        $sql = "DELETE FROM post_tags WHERE post_id=$post_id AND tag_id=$tag_id LIMIT 1";
        $results = $tbdb->query($sql);
        if(!$results) return false;

        return true;
    }

    /**
     * 获取所有的标签以及其拥有的文章数
     *
     * 返回时按拥有的文章数量从多到少排列
     *
     * @param int $limit 获取至多 $limit 条记录
     *
     * @return 返回对象数组，对象格式为：
     *      {
     *          "tags.*": "*",
     *          "size": 拥有的文章数,
     *      }
     */
    public function list_all_tags(int $limit)
    {
        global $tbdb;

        $limit = (int)$limit;

        $sql = array();
        $sql['select'] = "t.*,COUNT(pt.id) as size";
        $sql['from'] = "post_tags pt,tags t";
        $sql['where'] = "pt.tag_id=t.id";
        $sql['groupby'] = "t.id";
        $sql['orderby'] = "size DESC";

        if ($limit > 0) {
            $sql['limit'] = $limit;
        }

        $sql = make_query_string($sql);
        $results = $tbdb->query($sql);

        if (!$results) {
            return false;
        }

        $tag_objs = [];

        while ($to = $results->fetch_object()) {
            $tag_objs[] = $to;
        }

        return $tag_objs;
    }

    /**
     * 更新某标签数据
     * 
     * @param int    $id    标签编号
     * @param string $name  新标签名字
     * @param int    $alias 新标签别名
     * 
     * @return boolean
     */
    public function updateTag(int $id, string $name, int $alias)
    {
        global $tbdb;

        $r = false;

        $sql = "UPDATE tags SET name=?,alias=? WHERE id=? LIMIT 1";
        if ($stmt = $tbdb->prepare($sql)) {
            if ($stmt->bind_param('sii', $name, $alias, $id)) {
                $r = $stmt->execute();
            }
            $stmt->close();
        }

        return $r;
    }
}

