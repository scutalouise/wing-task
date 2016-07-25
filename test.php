<?php

require "./client.php";

$c = new TaskClient();
$c->connect("127.0.0.1", 8989);

function Usr1() {
    global $c;
    while (true) {
    if ($c->Usr1("test")) {
        if ($c->GetJob("test", $key, $data)) {
            var_dump("sleep");
//            sleep(1);
            $c->SetReturn($key, json_encode(array("xxxxxxxx")));
        }
    }
    }
}

function Addjob() {
    global $c;
    $c->Addjob("test", "ddd" ,$key);
    echo $key, "\n";
}

function SetReturn($key) {
     global $c;
    if ($c->SetReturn($key, "ddd")) {
        echo $key, "\n";
    }
    
}

function GetReturn($key) {
    global $c;
    if ($c->GetReturn($key, $data, 1000)) {
        var_dump($data);
    } else {
        var_dump($c->getErrMsg());
    }
}

//  Usr1();
//
// echo serialize(array(1,2));

Addjob();
// SetReturn();

//$key = $argv[1];
//GetReturn($key);