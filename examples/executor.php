<?php

declare(strict_types=1);

const JOBS = [
    [
        'identifier' => 'job 1',
        'command' => 'php',
        'args' => [
            './examples/worker.php'
        ]
    ],
    [
        'identifier' => 'job 2',
        'command' => 'php',
        'args' => [
            './examples/worker.php'
        ]
    ],
    [
        'identifier' => 'job 3',
        'command' => 'php',
        'args' => [
            './examples/worker.php'
        ]
    ],
    [
        'identifier' => 'job 4',
        'command' => 'php',
        'args' => [
            './examples/worker.php'
        ]
    ],
    [
        'identifier' => 'job 5',
        'command' => 'php',
        'args' => [
            './examples/worker.php'
        ]
    ],
    [
        'identifier' => 'job 6',
        'command' => 'php',
        'args' => [
            './examples/worker.php'
        ]
    ],
    [
        'identifier' => 'job 7',
        'command' => 'php',
        'args' => [
            './examples/worker.php'
        ]
    ],
    [
        'identifier' => 'job 8',
        'command' => 'php',
        'args' => [
            './examples/worker.php'
        ]
    ]
];


// Run jobs from arguments with max 4 worker threads
$jobs = json_encode(JOBS);
exec("./go-workerpool -t 4 '$jobs'", $stdout);
if (empty($stdout)) {
    echo "No output from job. Consider running in verbose mode\n";
    return;
}

$results = json_decode(end($stdout), true);
foreach ($results as $result) {
    echo  "Result for {$result['identifier']}: ({$result['exit_code']}) ". ($result['result_body'] ?? 'empty') . "\n";
}
