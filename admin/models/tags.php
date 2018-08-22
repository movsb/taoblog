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
