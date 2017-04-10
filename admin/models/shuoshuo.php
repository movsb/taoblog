<?php

class TB_Shuoshuo {
    public function post($post) {
        global $tbdb;
        global $tbdate;

        $date = $post['date'];
        if(!$date) $date = $tbdate->mysql_datetime_local();
        if(!$tbdate->is_valid_mysql_datetime($date)) {
            $this->error = '无效时间。';
            return false;
        }

        $date = $tbdate->mysql_local_to_gmt($date);

        $sql = "INSERT INTO shuoshuo (content,geo_lat,geo_lng,geo_addr,date,source) VALUES(?,?,?,?,?)";
        if($stmt = $tbdb->prepare($sql)) {
            if($stmt->bind_param('sddsss', $post['content'], $post['lat'], $post['lng'], $post['addr'],  $date, $post['source'])) {
                $r = $stmt->execute();
                $stmt->close();

                if($r){
                    $iid = $tbdb->insert_id;
                    return $iid;
                }
            }
        }

        $this->error = $tbdb->error or $stmt->error;
        return false;
    }

    public function update($id, $post) {
        global $tbdb;
        global $tbdate;

        $id = (int)$id;

        $date = $post['date'];
        if(!$tbdate->is_valid_mysql_datetime($date)) {
            $this->error = '无效修改时间。';
            return false;
        }

        $date = $tbdate->mysql_local_to_gmt($date);

        $sql = "UPDATE shuoshuo SET content=?,geo_lat=?,geo_lng=?,geo_addr=?,date=?,source=? WHERE id=$id";
        if($stmt = $tbdb->prepare($sql)) {
            if($stmt->bind_param('sddsss', $post['content'], $post['lat'], $post['lng'], $post['addr'], $date, $post['source'])) {
                $r = $stmt->execute();
                $ars = $stmt->affected_rows; // 貌似无需判断

                return $r;
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
        global $tbdate;

        $id = (int)$id;
        $sql = "SELECT * FROM shuoshuo WHERE id=$id";
        $rows = $tbdb->query($sql);
        if(!$rows) return $shuoshuos;

        $ss = $rows->fetch_object();
        $ss->date = $tbdate->mysql_datetime_to_local($ss->date);
        return $ss;
    }

    public function has($id) {
        global $tbdb;

        $id = (int)$id;
        $sql = "SELECT id FROM shuoshuo WHERE id=$id";
        $rows = $tbdb->query($sql);
        return $rows !== false && $rows->num_rows == 1;
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

class TB_ShuoshuoComments {
    public function post($sid, $author, $content, $date='') {
        global $tbdb;
        global $tbdate;

        $sid = (int)$sid;
        if($sid <=0) return false;
        if(!$content) return false;
        if(!$author) return false;
        if(!$date || !$tbdate->is_valid_mysql_datetime($date)) {
            $date = $tbdate->mysql_datetime_gmt();
        }

        $sql = "INSERT INTO shuoshuo_comments (sid,author,date,content) VALUES (?,?,?,?)";
        if($stmt = $tbdb->prepare($sql)) {
            if($stmt->bind_param('isss', $sid, $author, $date, $content)) {
                $r = $stmt->execute();
                $stmt->close();

                if($r) {
                    $iid = $tbdb->insert_id;
                    return $iid;
                }
            }
        }

        return false;
    }

    public function &get($sid) {
        global $tbdb;
        global $tbdate;

        $comments = [];

        $sid = (int)$sid;
        if($sid <= 0) return $comments;

        $sql = "SELECT * FROM shuoshuo_comments WHERE sid=$sid";
        $rows = $tbdb->query($sql);
        if(!$rows) return $comments;

        while($r = $rows->fetch_object()) {
            $r->date = $tbdate->mysql_datetime_to_local($r->date);
            $comments[] = $r;
        }

        return $comments;
    }
}

