<?php


// Client.
function test1() {
    $task = new TaskClient();
    $task->connect("127.0.0.1", "9090");
    // key 是服务端生成的唯一值
    if ($task->AddJob("test1", "data(添加数据)", $key)) {

        // 获取任务被执行后的值, 异常，超时
        if($task->GetReturn($key, $data)) {
            var_dump($data);
        } else {
            if ($task->GetErrNo() == 408) {
                var_dump("超时");
            } else {
                var_dump($task->GetErrMsg());
            }
        }
    }
}

// Worker.
function test2() {
    $task = new TaskClient();
    $task->connect("127.0.0.1", "9090");
    if ($task->Usr1("test1")) {
        if ($task->GetJob("test1", $key, $data)) {
            // TODO 处理数据.
            $task->SetReturn($key, "data(设置任务完成后结果数据)");
        }
    }
}