<?php

require "./client.php";

$c = new TaskClient();
$c->connect("127.0.0.1", 8989);

if (isset($argv[1])) {
    add();
} else {
    usr1();
}

function add() {
    global  $c;
    while (true) {

        $keys = array();
        for ($i = 0; $i < 10; $i++) {
            if ($key = $c->AddJob("ssssss", "ddd")) {
                $keys[] = $key;
            }
        }

        foreach ($keys as $key) {
            if ($data = $c->GetReturn($key,3000)) {
                var_dump("key:" . $key .  ", data:" . $data);
            } else {
                var_dump($c->GetErrMsg());
            }
        }
        sleep(3);
    }
}

function usr1() {
    global  $c;
    while (true) {
        if ($c->Usr1("ssssss")) {
            if ($c->GetJob("ssssss", $key, $data)) {
//                sleep(1);
                var_dump("DO:" . $key);
                if (mt_rand(1,2) == 1) {
                    $data = file_get_contents("./t1.log");
                } else {
                    $data = file_get_contents("./t2.log");
                }

                if ($c->SetReturn($key, $data)) {
                    var_dump("成功:" . $data);
                }
            }
        }
    }
}
