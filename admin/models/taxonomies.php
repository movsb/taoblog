<?php

class TB_Taxonomies {
    public $error = '';

    public function get_sons_by_id($id){
        global $tbdb;

        $sql = "SELECT id FROM taxonomies WHERE parent=".intval($id);
        $rows = $tbdb->query($sql);
        if(!$rows) return false;

        $sons = [];
        while($r = $rows->fetch_object())
            $sons[] = $r->id;

        return $sons;
    }

    public function get_parents_ids($id) {
        global $tbdb;

        if(!$this->has($id)) return false;

        $parents = [];

        while($id) {
            $t = $this->get_parent_id($id);
            if(!$t) break;
            $parents[] = $t;
            $id = $t;
        }

        return array_reverse($parents);
    }

    public function add(&$arg){
        global $tbdb;

        if(!$arg['name'] || preg_match('# |	|\'|"|;|/|\\\\#', $arg['name'])) {
            $this->error = '分类名不符合规范！';
            return false;
        }

        if(!$arg['slug'] || preg_match('# |	|\'|"|;|/|\\\\#', $arg['slug'])) {
            $this->error = '分类别名不符合规范！';
            return false;
        }

        if((int)$arg['parent']!=0 && !$this->has($arg['parent'])) {
            $this->error = '分类父ID不存在！';
            return false;
        }

        if($this->has_child((int)$arg['parent'], $arg['slug'])) {
            $this->error = '此父分类下已有别名为 `'.$arg['slug'].'\' 的子分类。';
            return false;
        }

        $arg['ancestor'] = $this->get_ancestor($arg['parent'], true);

        $sql = "INSERT INTO taxonomies (name,slug,parent,ancestor) VALUES (?,?,?,?)";
        if(($stmt = $tbdb->prepare($sql))
            && $stmt->bind_param(
                'ssii',
                $arg['name'],
                $arg['slug'],
                $arg['parent'],
                $arg['ancestor']
                )
            && $stmt->execute())
        {
            $stmt->close();
            $id = $tbdb->insert_id;
            return $id;
        }

        $this->error = $stmt->error;
        return false;
    }

    public function update(&$arg){
        global $tbdb;

        if(!$this->has($arg['id'])) {
            $this->error = '此分类ID不存在！';
            return false;
        }

        $sql = "UPDATE taxonomies SET name=?,slug=? WHERE id=? LIMIT 1";
        if(($stmt = $tbdb->prepare($sql))
            && $stmt->bind_param('ssi',
                $arg['name'],
                $arg['slug'],
                $arg['id'])
            && $stmt->execute())
        {
            $stmt->close();
            return true;
        }

        $this->error = $tbdb->error;
        return false;
    }

    public function has($id){
        global $tbdb;
        $id = (int)$id;
        $sql = "SELECT name FROM taxonomies WHERE id=$id LIMIT 1";
        return ($rs = $tbdb->query($sql)) && $rs->num_rows > 0;
    }

    /*
     * 判断某个父分类下是否存在某个 slug 的子分类
     *
     * 同一个父分类（包括ID为0的根分类）下不能存在slug 相同的子分类，需加以判断。
     * 这里只判断 slug 是否存在，不判断 name。slug 不同，name 一般就不同了。索引分类
     * 文章时是按照 slug 来搜索的，所以就算 name 相同也无关紧要。
     */
    public function has_child($pid, $slug) {
        global $tbdb;

        $pid = (int)$pid;
        $slug = $tbdb->real_escape_string($slug);

        $sql = "SELECT id FROM taxonomies WHERE parent=$pid and slug='$slug' LIMIT 1";
        return ($rs = $tbdb->query($sql)) && $rs->num_rows > 0;
    }

    /* 按ID删除某个分类
        需要删除的有：
            *. 递归删除以此ID为Parent的分类
            *. 删除以此ID为ID的分类
    */
    public function del($id){
        global $tbdb;

        if(!$this->has($id)) {
            $this->error = '待删除的分类ID不存在！';
            return false;
        }

        // 先取得孩子，全部递归删除
        $sons = $this->get_sons_by_id($id);
        foreach($sons as $son){
            $this->del($son);
        }

        // 最后删除自己
        $sql = "DELETE FROM taxonomies WHERE id=$id LIMIT 1";
        return $tbdb->query($sql);
    }

    // TODO 删除
    public function tree_from_id($id) {
        global $tbdb;

        if(!$this->has($id)) return false;

        $slug = [];
        $name = [];
        while($id) {
            $t = $this->get($id)[0];
            $slug[] = $t->slug;
            $name[] = $t->name;

            $id = $t->parent;
        }

        $slug = array_reverse($slug);
        $name = array_reverse($name);

        return compact('slug', 'name');
    }

    public function id_from_tree($tree) {
        global $tbdb;

        $ts = preg_split('~/~', $tree, -1, PREG_SPLIT_NO_EMPTY);
        if(count($ts)<1) return false;

        $sql = "SELECT id FROM taxonomies WHERE slug='".$tbdb->real_escape_string($ts[count($ts)-1])."'";
        if(count($ts) == 1) {
            $sql .= " AND parent=0 LIMIT 1";
        } else {
            $sql .= " AND parent IN (";
            for($i=count($ts)-2; $i>0; --$i) {
                $sql .= "SELECT id FROM `taxonomies` WHERE slug='".$tbdb->real_escape_string($ts[$i])."' AND parent IN (";
            }
            $sql .= "SELECT id FROM `taxonomies` WHERE slug='".$tbdb->real_escape_string($ts[0])."'";
            for($i=count($ts)-1; $i>0; --$i) {
                $sql .= ")";
            }
        }

        $rows = $tbdb->query($sql);
        if(!$rows || !$rows->num_rows) return false;

        return $rows->fetch_object()->id;
    }
}

