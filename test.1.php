<?php

require "./client.php";

$c = new TaskClient();
$c->connect("127.0.0.1", 8989);

function Addjob()
{
    global $c;
    while (true) {
        $keys  = array();
        var_dump("添加任务");
        for ($i = 0; $i < 1000; $i++) {
            $c->Addjob("test", "ddd", $key);
            $keys[] = $key;
        }

        var_dump("获取数据");
        foreach ($keys as $key) {
            $res = '';
            if ($c->GetReturn($key,$res, 1000)) {
                var_dump($res);
            } else {
                var_dump($key, $c->GetErrMsg());
                exit;
            }
        }
    }
}

Addjob();
