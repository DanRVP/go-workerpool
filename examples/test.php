<?php

declare(strict_types=1);

// Simulate some work on thread
sleep(rand(0, 2));

// "Randomise" exit code
exit(rand(0, 1));
