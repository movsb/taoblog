<?php

class TB_Shuoshuo {
    public function post($content, $date='') {
        global $tbdb;
        global $tbdate;

        if(!$content) return false;
        if(!$date) $date = $tbdate->mysql_datetime_gmt();

        $sql = "INSERT INTO shuoshuo (content,date) VALUES(?,?)";
        if($stmt = $tbdb->prepare($sql)) {
            if($stmt->bind_param('ss', $content, $date)) {
                $r = $stmt->execute();
                $stmt->close();

                if($r){
                    $iid = $tbdb->insert_id;
                    return $iid;
                }
            }
        }

        $this->error = $stmt->error;
        return false;
    }

    public function update($id, $content) {
        global $tbdb;

        $id = (int)$id;
        $sql = "UPDATE shuoshuo SET content=? WHERE id=$id";
        if($stmt = $tbdb->prepare($sql)) {
            if($stmt->bind_param('s', $content)) {
                $r = $stmt->execute();
                $ars = $stmt->affected_rows;

                return $r && $ars == 1;
            }
        }

        return false;
    }

    public function del($id) {
        global $tbdb;

        $id = (int)$id;
        $sql = "DELETE FROM shuoshuo WHERE id=$id LIMIT 1";
        $r = $tbdb->query($sql);
        return $r;
    }

    public function get($id) {
        global $tbdb;

        $id = (int)$id;
        $sql = "SELECT content FROM shuoshuo WHERE id=$id";
        $rows = $tbdb->query($sql);
        if(!$rows) return $shuoshuos;
        return $rows->fetch_object()->content;
    }

    public function &get_latest($count) {
        global $tbdb;
        global $tbdate;

        $shuoshuos = [];

        $count = (int)$count;
        if($count <= 0) return $shuoshuos;

        $sql = "SELECT * FROM shuoshuo WHERE 1 ORDER BY date DESC LIMIT $count";
        $rows = $tbdb->query($sql);
        if(!$rows) return $shuoshuos;

        while($r = $rows->fetch_object()) {
            $r->date = $tbdate->mysql_datetime_to_local($r->date);
            $shuoshuos[] = $r;
        }

        return $shuoshuos;
    }
}

