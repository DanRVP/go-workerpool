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

$filename = 'infile.json';
$put_res = file_put_contents($filename, json_encode(JOBS));
if ($put_res === false) {
    echo "Unable to write infile $infile";
    return;
}

// Run jobs from infile with max 4 worker threads
exec("./go-workerpool -i $filename -t 4", $stdout);
if (empty($stdout)) {
    echo "No output from job. Consider running in verbose mode\n";
    return;
}

$results = json_decode(end($stdout), true);
foreach ($results as $result) {
    echo  "Result for {$result['identifier']}: ({$result['exit_code']}) ". ($result['result_body'] ?? 'empty') . "\n";
}
