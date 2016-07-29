<?php

require "./client.php";

$c = new TaskClient();
$c->connect("127.0.0.1", 8989);


$c->GetReturn($argv[1], $data, 3000);

var_dump($data);

