<?php

declare(strict_types=1);

$sleep_time = rand(0, 5);
sleep($sleep_time);

$failure = rand(0, 1);
if ($failure) {
    echo "Failure. I slept for $sleep_time seconds";
} else {
    echo "The time is " . date('H:i:s', time()) . ". I slept for $sleep_time seconds";
}

exit($failure);
