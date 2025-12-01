package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "Run",
		Usage:   "Run a set of tasks in parallel and await the results",
		Action:  run,
		Suggest: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "infile",
				Usage:    "The path to the file which contains the task definitions",
				Required: true,
				Aliases:  []string{"i"},
			},
			&cli.StringFlag{
				Name:    "outfile",
				Value:   "",
				Usage:   "The path to the file which should receive the results",
				Aliases: []string{"o"},
			},
			&cli.IntFlag{
				Name:    "max_threads",
				Value:   2,
				Usage:   "Set the maximum of parallel goroutines which can run tasks",
				Aliases: []string{"t"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Value:   false,
				Usage:   "Whether to enable verbose mode or not",
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

	infilePath := cmd.String("infile")
	verbose := cmd.Bool("verbose")

	out(fmt.Sprintf("Getting tasks from %s", infilePath), verbose)
	fileContents, err := os.ReadFile(infilePath)
	if err != nil {
		message := fmt.Sprintf("Unable to read infile. Error was: %v", err)
		return cli.Exit(message, 1)
	}

	var tasks_todo []Task
	err = json.Unmarshal(fileContents, &tasks_todo)
	if err != nil {
		message := fmt.Sprintf("Unable to unmarshal infile. Error was: %v", err)
		return cli.Exit(message, 1)
	}

	maxThreads := cmd.Int("max_threads")
	out(fmt.Sprintf("max_threads set to %d\n", maxThreads), verbose)

	length := len(tasks_todo)
	out(fmt.Sprintf("%d tasks to process\n", length), verbose)
	if length < 1 {
		return nil
	}

	workerCount := min(maxThreads, length)
	tasks := make(chan *Task, length)
	results := make(chan *Result)
	var wg sync.WaitGroup

	// Start workers
	for range workerCount {
		wg.Add(1)
		go worker(&wg, tasks, results, verbose)
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
		fmt.Printf("%s\n", json)
		return nil
	}

	os.WriteFile(outfile, json, 0644)
	return nil
}

// Run a task in its own thread and push to results
func worker(wg *sync.WaitGroup, tasks <-chan *Task, results chan<- *Result, verbose bool) {
	defer wg.Done()

	for task := range tasks {
		fullCommand := strings.TrimSpace(task.Command + " " + strings.Join(task.Args, " "))
		cmd := exec.Command(task.Command, task.Args...)
		out(fmt.Sprintf("Running %s\n", fullCommand), verbose)

		exitCode := 0
		resultBody := ""
		if err := cmd.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				status := exitError.Sys().(syscall.WaitStatus)
				exitCode = status.ExitStatus()
			}

			out(fmt.Sprintf("Unable to run \"%s\". Error was %v\n", fullCommand, err), verbose)
		} else {
			resultBody = fullCommand
		}

		results <- &Result{
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
	Identifier string
	Command    string
	Args       []string
}

type Result struct {
	Identifier string `json:"identifier"`
	ExitCode   int    `json:"exit_code"`
	ResultBody string `json:"result_body"`
}
