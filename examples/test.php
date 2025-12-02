<?php

declare(strict_types=1);

sleep(rand(0, 2));
$failure = rand(0, 1);
if ($failure) {
    echo "Failure";
} else {
    echo "The time is " . time();
}

exit($failure);
