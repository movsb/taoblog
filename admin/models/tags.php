<?php

/**
 *  数据库标签管理模块
 *
 *  该模块用来管理数据库中的标签表
 *
 */
class TB_Tags
{
    public $error = '';

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

        $rids = [];
        $sids = join(',', $ids);

        $sql1 = "SELECT alias FROM tags WHERE id in ($sids)";
        $sql2 = "SELECT id FROM tags WHERE alias in ($sids)";

        $results1 = $tbdb->query($sql1);
        $results2 = $tbdb->query($sql2);

        if (!$results1 || !$results2) {
            return $ids;
        }

        while ($n = $results1->fetch_object()) {
            if ($n->alias > 0) {
                $rids[] = $n->alias;
            }
        }

        while ($n = $results2->fetch_object()) {
            $rids[] = $n->id;
        }

        $ids = array_merge($ids, $rids);
        return $ids;
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
    public function list_all_tags(int $limit, bool $mergeAliases)
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
        $tag_alias_objs = [];

        while ($to = $results->fetch_object()) {
            if ($to->alias == 0) {
                $tag_objs[] = $to;
            } else {
                $tag_alias_objs[] = $to;
            }
        }

        if (!$mergeAliases) {
            return array_merge($tag_objs, $tag_alias_objs);
        } else {
            // 写得好麻烦
            foreach ($tag_objs as $to) {
                foreach ($tag_alias_objs as $ta) {
                    if ($ta->alias == $to->id) {
                        $to->size += $ta->size;
                    }
                }
            }
            return $tag_objs;
        }
    }
}
