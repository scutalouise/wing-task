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
    public function __construct($address = '', $port = 0)
    {
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
     *
     * @return boolean
     */
    public function AddJob($tube, $data)
    {
        $str = $this->format(array("AddJob", $tube, $data));
        $ok = $this->finish($str);
        if ($ok !== false) {

            return $ok[0];
        }

        return false;
    }

    /**
     * 获取一个任务.
     *
     * @param string $tube 队列名称.
     * @param string &$_key  任务唯一标示KEY.
     * @param string &$_tmp 数据.
     *
     * @return bool
     */
    public function GetJob($tube, &$_key, &$_tmp)
    {
        $str = $this->format(array("GetJob", $tube));
        $ok = $this->finish($str);
        if ($ok !== false) {
            $_key = $ok[0];
            $_tmp = $ok[1];

            return true;
        }

        return false;
    }


    /**
     * 获取任务完成的结果.
     *
     * @param string $key     唯一标示KEY.
     * @param string $timeout 超时.
     *
     * @return boolean
     */
    public function GetReturn($key, $timeout)
    {
        $str = $this->format(array("GetReturn", $key, $timeout));
        $ok = $this->finish($str);
        if ($ok !== false) {

            return $ok[0];
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
        $ok = $this->finish($str);
        if ($ok !== false) {
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
        $ok = $this->finish($str);
        if ($ok !== false) {
            return true;
        }

        return false;
    }

    /**
     * 完成一个网络请求.
     *
     * @param string $str 数据.
     *
     * @return boolean
     */
    private function finish($str)
    {
        if (!self::$status) {
            $this->connect();
        }
        if (!self::$status) {
            return false;
        }

        try {
            $count = 0;
            $len = strlen($str);
            while ($count < $len) {
                $n = @socket_write(self::$socket, substr($str, $count));
                $count += $n;
                $this->doExt();
            }

            $ok = $this->getOneRequest();
        } catch (Exception $e) {
            return false;
        }

        $tmp = array();
        if ($ok !== false) {
            if ($ok[0] == 1) {
                $len = count($ok);
                for ($i = 2; $i < $len; $i++) {
                    $tmp[] = $ok[$i];
                }

                return $tmp;
            } else {
                self::$errno = $ok[0];
                self::$errmsg = $ok[1];
            }
        }

        return false;
    }

    /**
     * 获取一个完整的请求结果.
     *
     * @return boolean
     */
    private function getOneRequest()
    {
        $buf = "";
        $data = array();
        $this->readString($buf, "*");
        $len = $this->readString($buf, "\n");
        $len = intval(trim($len));
        for ($i = 0; $i < $len; $i++) {
            $ok = $this->readString($buf, "$");
            if ($ok === false) {
                return false;
            }
            $ok = $dataSize = $this->readString($buf, "\n");
            if ($ok === false) {
                return false;
            }
            $ok = $dataSize = intval(trim($dataSize));
            if ($ok === false) {
                return false;
            }
            $data[] = $this->readSize($buf, $dataSize);
        }

        return $data;
    }

    private function readString(&$buf, $delim = "\n")
    {
        while (true) {
            $attr =strpos($buf, $delim);
            if ($attr === false) {
                $buf = socket_read(self::$socket, 2048, PHP_BINARY_READ);
                $this->doExt();
                continue;
            }
            break;
        }
        
        $str = substr($buf, 0, $attr + 1);
        $buf = substr($buf, $attr + 1);
        
        return $str;
    }

    private function readSize(&$buf, $size) {

        while (true) {
            if (strlen($buf) < $size) {
                $buf .= socket_read(self::$socket, $size, PHP_BINARY_READ);
                $this->doExt();
            } else {
                break;
            }
        }

        $str = substr($buf, 0, $size);
        $buf = substr($buf, $size);

        return $str;
    }

    private function doExt()
    {
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
