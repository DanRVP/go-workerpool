# go-workerpool

Simple CLI program to invoke multiple processes in parallel using goroutines.

## Why?
I needed to solve a problem. I had to run existing PHP code in parallel and wait for the results, then continue with those results in PHP. As PHP is not a natively multithreaded language and the technique of using `exec` in a non-blocking manner to run tasks in parallel would complicate waiting for and fetching results if gave myself the following brief:

```
The overarching approach will be to use a language which supports parallelisation (in our case Go) and use that to manage the PHP processes which will build, make and process HTTP requests.

The process manager written in Go should have the following characteristics:

1. Be reusable: The processor should be job agnostic. A properly configured PHP shell should be able to be used by the shell regardless of the job being done.
2. Be configurable: The total number of parallel processes should be able to be controlled.
3. Be invokable via PHP: We need to be able to run this manager from within PHP code.
4. Be able to listen for and collate results from each of the child processes.
```

The result of this is this repo

## Compiling
- Install golang
- Run `go build go-workerpool.go` to generate an executable binary

## Running
### Displaying help
To get help you can run `./go-workerpool -h`

### Providing tasks
The program expects an array of the `Task` struct defined in `go-workerpool.go`. For example:
```json
[
    {
        "identifier": "job 1",
        "command": "php",
        "args": [
            "worker.php",
            "an_argument"
        ]
    },
    {
        "identifier": "job 2",
        "command": "php",
        "args": [
            "worker.php",
            "a_different_argument"
        ]
    }
]
```

You can provide this JSON as either the first argument when invoking the program or by providing a path to a JSON file using the `--infile` option. If you provide both the argument and the option, the option will take precedent.

You can see example usage of how to go about providing tasks and running the program in `examples/executor.php`

### Specifying a number of worker threads
By default the program will run using 2 threads. If you wish to allow more threads to be used then include the ``--max_threads`` option when invoking the program. This option accepts an integer. If the number of threads provided exceeds the number of tasks, then the number of tasks will be used as the thread count.

### Handling output
By default the results of the task are directed to stdout. Results are structured as an array of the `Result` struct defined in `go-workerpool.go`. For example:
```json
[
    {
        "identifier": "job 1",
        "exit_code": 1,
        "result_body": "Failure. I slept for 3 seconds"
    },
    {
        "identifier": "job 2",
        "exit_code": 0,
        "result_body": "The time is 13:45:01. I slept for 3 seconds"
    }
]
```

If the `--verbose` flag is provided then additional messaging will be sent to stdout. The results JSON should always be the final output before the program exits.

If you prefer you can send the results JSON to a file using the `--outfile` option. This accepts a path to a file. If specified, then the JSON results are not sent to stdout.

## Writing tasks
The tasks you write should be invokable by your CLI. Any results you want to be captured by the workerpool should be sent to stdout.

You can see example usage of how to go about writing a task in `examples/worker.php`
