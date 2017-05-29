<?php

/*
 * API
 *
 */

class TB_API
{
    private $module = '';
    private $method = '';

    /*
     * 结束一条API请求
     *
     * @param $obj 任意有效的 json 值类型
     *
     * @return nothing
     */
    public function die($obj)
    {
        header('HTTP/1.1 200 OK');
        header('Content-Type: application/json');
        echo json_encode($obj, JSON_UNESCAPED_UNICODE);
        die(0);
    }

    /*
     * 以错误码+错误消息结束请求
     *
     * @param code 错误码
     * @param msg  错误消息
     *
     * @return nothing
     */
    public function err(int $code, string $msg)
    {
        $this->die([
            'code' => $code,
            'msg'  => $msg,
        ]);
    }

    /*
     * 成功结束一条请求
     *
     * @param data 返回数据
     */
    public function done($dat)
    {
        $this->die([
            'code' => 0,
            'data' => $dat,
        ]);
    }

    /*
     * 返回请求的方法不存在错误
     *
     */
    public function bad()
    {
        $this->err(-1, "Unknown method `{$this->method}' of module `{$this->module}'.");
    }

    /*
     * 授权一个请求
     *
     * 如果已登录，不作任何操作
     * 如果未登录，且请求不是登录，则返回登录请求
     * 如果未登录，且请求是登录，则验证登录并返回授权
     *
     */
    public function auth() {
        global $logged_in;

        // 已登录
        if($logged_in) {
            return;
        }

        // 未登录，登录？
        if($this->module != 'login' || $this->method != 'auth') {
            $this->err(-1, 'login please');
        }


        $user   = $_REQUEST['user'] ?? '';
        $passwd = $_REQUEST['passwd'] ?? '';

        $arg = compact('user', 'passwd');

        $ok = login_auth_passwd($arg);

        if($ok) {
            $this->done([
                "login" => login_gen_cookie(),
            ]);
        }
        else {
            $this->err(-1, 'login auth failed.');
        }
    }

    /*
     * 检测请求中是否有某个参数
     *
     * @param $name 参数的名字
     *
     * @return 如果参数存在，返回其值；
     *         如果不存在，结束请求并返回缺少参数错误
     */
    public function expected(string $name)
    {
        if(!isset($_REQUEST[$name])) {
            $this->err(-1, "expect argument `$name'");
        }
        else {
            return $_REQUEST[$name];
        }
    }

    /*
     * 检测请求中是否有某个可选参数
     *
     * @param $name 参数名字
     *
     * @param $def 指定的默认值
     *
     * @return 如果参数存在，返回其值；
     *         如果参数不存在，返回指定的默认值；
     */
     public function optional(string $name, $def)
     {
         return $_REQUEST[$name] ?? $def;
     }

     /*
      * 初始化
      *
      */
    public function init(string $module, string $method)
    {
        $this->module = $module;
        $this->method = $method;
    }
}

