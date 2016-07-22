<?php

require "./client.php";

$c = new TaskClient();
$c->connect("127.0.0.1", 8989);

function Usr1() {
    global $c;
    if ($c->Usr1("test")) {
        echo "1";
    }
}

function Addjob() {
    global $c;
    $c->Addjob("test", "ddd" ,$key);
    echo $key, "\n";
}

function SetReturn() {
     global $c;
     $key = '18onq5vecpmml';
    if ($c->SetReturn($key)) {
        echo $key, "\n";
    }
    
}

function GetReturn() {
    global $c;
    $key = '18onq5vecpmml';
    if ($c->GetReturn($key, $data)) {
        var_dump($data);
    }
}

// Usr1();

// echo serialize(array(1,2));

// Addjob();
// SetReturn();
GetReturn();