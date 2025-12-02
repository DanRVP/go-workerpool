package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/urfave/cli/v3"
)

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}

	cmd := &cli.Command{
		Name:    "go-workerpool",
		Usage:   "Run a set of tasks in parallel and await the results",
		Action:  run,
		Suggest: true,
		Version: "v0.0.1",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "tasks",
				UsageText: "(Optional) JSON encoded string of tasks to execute. If omitted then the --infile option must be used",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "infile",
				Usage:   "(Optional) The path to the file which contains the task definitions. Will override the tasks argument if included",
				Aliases: []string{"i"},
			},
			&cli.StringFlag{
				Name:    "outfile",
				Usage:   "(Optional) The path to the file which should receive the results",
				Aliases: []string{"o"},
			},
			&cli.IntFlag{
				Name:    "max_threads",
				Value:   2,
				Usage:   "(Optional) Set the maximum of parallel goroutines which can run tasks",
				Aliases: []string{"t"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Value:   false,
				Usage:   "(Optional) Whether to enable verbose mode or not. When disabled only the result JSON is output",
				Aliases: []string{"v"},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

// Run a set of tasks in parallel and await the results
func run(ctx context.Context, cmd *cli.Command) error {
	var err error
	var tasks_todo []Task

	verbose := cmd.Bool("verbose")
	taskArg := cmd.StringArg("tasks")
	infilePath := cmd.String("infile")

	var taskBytes []byte
	if taskArg == "" && infilePath == "" {
		message := "You must provide a JSON list of tasks in the first argument or define a filepath using --infile"
		return cli.Exit(message, 1)
	} else if taskArg != "" {
		taskBytes = []byte(taskArg)
	} else {
		out(fmt.Sprintf("Getting tasks from %s", infilePath), verbose)
		taskBytes, err = os.ReadFile(infilePath)
		if err != nil {
			message := fmt.Sprintf("Unable to read infile. Error was: %v", err)
			return cli.Exit(message, 1)
		}
	}

	err = json.Unmarshal(taskBytes, &tasks_todo)
	if err != nil {
		message := fmt.Sprintf("Unable to unmarshal infile. Error was: %v", err)
		return cli.Exit(message, 1)
	}

	maxThreads := cmd.Int("max_threads")
	out(fmt.Sprintf("max_threads set to %d", maxThreads), verbose)

	length := len(tasks_todo)
	out(fmt.Sprintf("%d tasks to process", length), verbose)
	if length < 1 {
		return nil
	}

	workerCount := min(maxThreads, length)
	tasks := make(chan *Task, length)
	results := make(chan *Result)
	var wg sync.WaitGroup

	// Start workers
	for i := range workerCount {
		wg.Add(1)
		go worker(i+1, &wg, tasks, results, verbose)
	}

	// Send jobs
	go func() {
		for i := range tasks_todo {
			tasks <- &tasks_todo[i]
		}
		close(tasks)
	}()

	// Wait for all tasks to finish.
	// This must be in its own thread or it will cause a deadlock with the blocking channel read below
	go func() {
		wg.Wait()
		close(results)
	}()

	var result_set []*Result
	for result := range results {
		result_set = append(result_set, result)
	}

	json, err := json.Marshal(result_set)
	if err != nil {
		message := fmt.Sprintf("Unable to marshal result to JSON. Error was: %v", err)
		return cli.Exit(message, 1)
	}

	outfile := cmd.String("outfile")
	if outfile == "" {
		fmt.Println(string(json))
		return nil
	}

	os.WriteFile(outfile, json, 0644)
	return nil
}

// Run a task in its own thread and push to results
func worker(workerId int, wg *sync.WaitGroup, tasks <-chan *Task, results chan<- *Result, verbose bool) {
	defer wg.Done()
	prefix := fmt.Sprintf("Worker %d", workerId)

	for task := range tasks {
		cmd := exec.Command(task.Command, task.Args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			message := fmt.Sprintf("Unable to open stdout pipe for %s. Error was %v", task.Identifier, err)
			out(message, verbose)
			results <- &Result{
				Identifier: task.Identifier,
				ExitCode:   1,
				ResultBody: message,
			}

			continue
		}

		out(fmt.Sprintf("%s: Starting %s", prefix, task.Identifier), verbose)
		exitCode := 0
		resultBody := ""
		if err := cmd.Start(); err != nil {
			out(fmt.Sprintf("%s: Unable to start \"%s\". Error was %v", prefix, task.Identifier, err), verbose)
			continue
		}

		resultBytes, err := io.ReadAll(stdout)
		if err != nil {
			out(fmt.Sprintf("%s: Unable to read result body bytes for %s. Error was: %v", prefix, task.Identifier, err), verbose)
		} else {
			resultBody = string(resultBytes)
		}

		err = cmd.Wait()
		if err != nil {
			out(fmt.Sprintf("%s: Error when running %s. Error was: %v", prefix, task.Identifier, err), verbose)
			if exitError, ok := err.(*exec.ExitError); ok {
				status := exitError.Sys().(syscall.WaitStatus)
				exitCode = status.ExitStatus()
			}
		}

		out(fmt.Sprintf("%s: Finished %s", prefix, task.Identifier), verbose)
		results <- &Result{
			Identifier: task.Identifier,
			ExitCode:   exitCode,
			ResultBody: resultBody,
		}
	}
}

// Convenience function for logging to stdout
func out(message string, verbose bool) {
	if verbose {
		fmt.Println(message)
	}
}

type Task struct {
	Identifier string   `json:"identifier"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
}

type Result struct {
	Identifier string `json:"identifier"`
	ExitCode   int    `json:"exit_code"`
	ResultBody string `json:"result_body"`
}
