package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// ResolverStatus holds the Alive boolean flag and the Name of the DNS resolver.
type ResolverStatus struct {
	Alive bool
	Name  string
}

// ResolverCheck holds the DNS resolver name, network protocol, and timeout duration.
type ResolverCheck struct {
	Resolver string
	Protocol string
	Timeout  time.Duration
}

// List of test hosts.
var testHosts = []string{"google.com", "cloudflare.com", "amazon.com"}

func main() {
	// Parsing command-line flags.
	list, output, protocol, workers, timeoutSec, silent := ParseFlags()

	// Loading DNS resolvers from a provided list.
	resolverChecks, err := LoadResolvers(list, protocol, time.Duration(timeoutSec)*time.Second)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Performing DNS checks.
	statuses := CheckResolvers(resolverChecks, workers, silent)

	// If an output file is provided, writing DNS resolver statuses to the output file.
	if output != "" {
		writeFile(output, statuses)
	}
}

// CheckResolvers performs DNS checks with requested workers.
func CheckResolvers(resolverChecks []ResolverCheck, workers int, silent bool) []ResolverStatus {
	// Creating channels to manage tasks and results.
	tasks := make(chan ResolverCheck, len(resolverChecks))
	wg := sync.WaitGroup{} // WaitGroup to synchronize worker goroutines.
	results := make(chan ResolverStatus, len(resolverChecks))

	// Starting worker goroutines.
	for i := 0; i < workers; i++ {
		go worker(tasks, &wg, results, silent)
	}

	// Feeding tasks to the worker goroutines.
	for _, check := range resolverChecks {
		wg.Add(1)
		tasks <- check
	}
	close(tasks)

	// Waiting for worker goroutines to finish.
	wg.Wait()

	// Closing the result channel after writing all results.
	close(results)

	// Preparing a list of DNS resolver statuses from the results channel.
	statuses := make([]ResolverStatus, 0, len(resolverChecks))
	for status := range results {
		statuses = append(statuses, status)
	}
	return statuses
}

// A worker goroutine that performs DNS checks for the provided tasks.
func worker(tasks <-chan ResolverCheck, wg *sync.WaitGroup, results chan<- ResolverStatus, silent bool) {
	for task := range tasks {
		// Performing a DNS check.
		results <- checkResolverStatus(task, silent)
		wg.Done()
	}
}

// CheckResolverStatus performs a DNS check and returns a corresponding DNS resolver status.
func checkResolverStatus(check ResolverCheck, silent bool) ResolverStatus {
	isAlive := isAlive(check.Resolver, check.Protocol, check.Timeout)
	if isAlive && !silent {
		fmt.Println(check.Resolver)
	}
	return ResolverStatus{Alive: isAlive, Name: check.Resolver}
}

// IsAlive checks if a DNS resolver is alive by performing a DNS lookup.
func isAlive(resolverHost, protocol string, timeout time.Duration) bool {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.Dial(protocol, resolverHost+":53")
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for _, host := range testHosts {
		_, err := r.LookupHost(ctx, host)
		if err != nil {
			return false
		}
	}
	return true
}

// LoadResolvers loads DNS resolvers for the provided list and from Stdin if available.
func LoadResolvers(filename, protocol string, timeout time.Duration) ([]ResolverCheck, error) {
	// Preparing ResolverCheck objects
	var checks []ResolverCheck

	// Check if there is input from Stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			checks = append(checks, ResolverCheck{Resolver: scanner.Text(), Protocol: protocol, Timeout: timeout})
		}
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
	}

	// If filename is provided, read from the file.
	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			checks = append(checks, ResolverCheck{Resolver: scanner.Text(), Protocol: protocol, Timeout: timeout})
		}
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
	}

	return checks, nil
}

// WriteFile writes DNS resolver statuses to a file.
func writeFile(fileName string, statuses []ResolverStatus) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	for _, resolver := range statuses {
		if resolver.Alive {
			_, err = fmt.Fprintln(file, resolver.Name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}

// ParseFlags parses command-line flags.
func ParseFlags() (string, string, string, int, int64, bool) {
	list := flag.String("list", "", "List of DNS resolvers")
	output := flag.String("output", "", "Output file")
	protocol := flag.String("protocol", "udp", "Network protocol")
	workers := flag.Int("workers", 10, "Number of workers")
	timeoutSec := flag.Int64("timeout", 1, "Timeout in seconds")
	silent := flag.Bool("silent", false, "Silent mode")

	flag.Parse()

	return *list, *output, *protocol, *workers, *timeoutSec, *silent
}
