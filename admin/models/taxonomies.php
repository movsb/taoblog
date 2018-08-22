<?php

class TB_Taxonomies {
    public $error = '';

    public function get_hierarchically() {
        $cats = Invoke('/categories:tree', 'json', null);
        return json_decode($cats);
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

