<?php

/**
 * Task任务中间件包.
 *
 * Class TaskClient
 */
class TaskClient
{
    /**
     * @var string $socket  连接对象.
     * @var string $status  连接对象状态.
     * @var string $address 连接地址.
     * @var string $port    端口.
     * @var string $errno   错误码.
     * @var string $errmsg  错误信息.
     */
    private static $socket  = "";
    private static $status  = false;
    private static $address = "";
    private static $port    = "";
    private static $errno   = "";
    private static $errmsg  = "";

    /**
    * @param string  $address 网络连接地址.
    * @param integer $port    端口.
    */
    public function __construct($address = '', $port = 0) {
        if (!empty($address)) {
            self::$address = $address;
        }

        if (!empty($port)) {
            self::$port = $port;
        }

        if (!empty(self::$address) && !empty(self::$port)) {
            $this->connect(self::$address, self::$port);
        }
    }
    /**
     * 连接网络对象.
     *
     * @param string  $address 连接地址.
     * @param integer $port    端口.
     *
     * @return boolean
     */
    public function connect($address, $port)
    {
        if (!empty($address)) {
            self::$address = $address;
        }
        if (!empty($port)) {
            self::$port = $port;
        }

        if (empty(self::$socket)) {
            self::$socket = $socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
        }

        self::$status = @socket_connect(self::$socket, self::$address, self::$port);
        self::$errno = socket_last_error(self::$socket);


        return self::$status;
    }

    /**
     * 格式化传输数据.
     *
     * @param array $arr 被被格式的数据. []string|integer|float
     *
     * @return string
     */
    private function format($arr)
    {

        $str = sprintf("*%d\n", count($arr));
        foreach ($arr as $item) {
            $str .= sprintf("$%d\n%s\n", strlen($item), $item);
        }

        return $str;
    }

    /**
     * 添加任务到队列.
     *
     * @param string $tube 队列名称.
     * @param string $data 数据.
     * @param string &$key 如果成功,添加到队列后,返回的唯一KEY.
     *
     * @return boolean
     */
    public function AddJob($tube, $data, &$key = '')
    {
        $str = $this->format(array("AddJob", $tube, $data));
        $ok = $this->finish($str, $tmp);
        if ($ok) {
            $key = $tmp[0];

            return true;
        }

        return false;
    }

    /**
     * 获取一个任务.
     *
     * @param string $tube 队列名称.
     * @param string $key  任务唯一标示KEY.
     * @param string $data 数据.
     *
     * @return bool
     */
    public function GetJob($tube, &$key, &$data)
    {
        $str = $this->format(array("GetJob", $tube));
        $ok = $this->finish($str, $tmp);
        if ($ok) {
            $key = $tmp[0];
            $data = $tmp[1];

            return true;
        }

        return false;
    }


    /**
     * 获取任务完成的结果.
     *
     * @param string $key  唯一标示KEY.
     * @param string $data 数据.
     * @param string $timeout 超时.
     *
     * @return boolean
     */
    public function GetReturn($key, &$data, $timeout)
    {
        $str = $this->format(array("GetReturn", $key, $timeout));
        $ok = $this->finish($str, $tmp);
        if ($ok) {
            $data = $tmp[0];

            return true;
        }

        return false;
    }

    /**
     * 设置任务的数据.
     *
     * @param string $key  任务唯一标示KEY.
     * @param string $data 数据.
     *
     * @return boolean
     */
    public function SetReturn($key, $data)
    {
        $str = $this->format(array("SetReturn", $key, $data));
        $ok = $this->finish($str, $tmp);
        if ($ok) {
            return true;
        }

        return false;
    }

    /**
     * 添加队列通知, 如果队列中存在数据或者向队列添加数据,立即返回.
     *
     * @param string $tube 队列名称.
     *
     * @return boolean
     */
    public function Usr1($tube)
    {
        $str = $this->format(array("Usr1", $tube));
        $ok = $this->finish($str, $tmp);
        if ($ok) {
            return true;
        }

        return false;
    }

    /**
     * 完成一个网络请求.
     *
     * @param string $str  数据.
     * @param array  $data $返回数据.
     *
     * @return boolean
     */
    private function finish($str, &$data = array())
    {
        if (empty(self::$socket)) {
            $this->connect();
        }
        if (!self::$status) {
            return false;
        }

        $int = @socket_write(self::$socket, $str);
        self::$errno = socket_last_error(self::$socket);
        if (!empty(self::$errno)) {
            self::$errmsg = socket_strerror(self::$errno);

            return false;
        }
        if ($int != strlen($str)) {
            self::$errno = "-2";
            self::$errmsg = "fail";

            return false;
        }

        $tmp = array();
        try {
            $ok = $this->getOneRequest($d);
        } catch (Exception $e) {
            return false;
        }
        if ($ok) {
            if ($d[0] == 1) {
                $len = count($d);
                for ($i = 2; $i < $len; $i++) {
                    $tmp[] = $d[$i];
                }
                $data = $tmp;

                return true;
            } else {
                self::$errno = $d[0];
                self::$errmsg = $d[1];
            }
        }

        return false;
    }

    /**
     * 获取一个完整的请求结果.
     *
     * @param array $data 返回的数据.
     *
     * @return boolean
     */
    private function getOneRequest(&$data)
    {
        // 查询到头
        $len = 0;
        while ($str = @socket_read(self::$socket, 1024, PHP_NORMAL_READ)) {
            $attr = strpos($str, "*");
            if ($attr !== false) {
                $len = substr($str, $attr + 1);
                break;
            }
        }
        $this->doExt();

        for ($i = 0; $i < $len; $i++) {
            while ($str = @socket_read(self::$socket, 1024, PHP_NORMAL_READ)) {
                $attr = strpos($str, "$");
                if ($attr !== false) {
                    $l = substr($str, $attr + 1);
                    $data[$i] = @socket_read(self::$socket, intval($l), PHP_BINARY_READ);
                    if ($l <= 0) {
                        $data[$i] = '';
                    }
                }
            }
            $this->doExt();
        }

        return true;
    }

    private function  doExt() {
        self::$errno = socket_last_error(self::$socket);
        if (!empty(self::$errno)) {
            self::$errmsg = socket_strerror(self::$errno);
            if (self::$errmsg == 'EOF') {
                self::$status = false;
            }
            throw  new Exception(self::$errmsg, self::$errno);
        }
    }

    /**
     * 获取异常信息.
     *
     * @return string
     */
    public function GetErrMsg()
    {
        return self::$errmsg;
    }

    /**
     * 获取异常号.
     *
     * @return string
     */
    public function GetErrNo()
    {
        return self::$errno;
    }

}
