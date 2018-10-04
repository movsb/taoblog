<?php

class TB_DateTime {
    // mysql格式的日期/时间格式 -> GMT
    public function  mysql_datetime_gmt($t=null) {
        return gmdate('Y-m-d H:i:s', $t?$t:time());
    }

    // HTTP GMT时间转本地MySQL时间
    public function http_gmt_to_mysql_datetime_gmt($g) {
        return $this->mysql_datetime_gmt(strtotime($g));

    }

    public function mysql_datetime_to_local($t) {
        return date('Y-m-d H:i:s', strtotime($t.' GMT+0000'));
    }

    public function mysql_local_to_http_gmt($t) {
        return gmdate('D, d M Y H:i:s \G\M\T', strtotime($t.' GMT+0800'));
    }

    public function mysql_local_to_timestamp($t) {
        return strtotime($t.' GMT+0800');
    }

    public function is_valid_mysql_datetime($t) {
        return preg_match('/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/', $t);
    }
}

