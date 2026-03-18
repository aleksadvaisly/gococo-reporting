package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aleksadvaisly/gococo-reporting"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gococo-reporting", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var inputCSV string
	var outputPath string
	var title string

	fs.StringVar(&inputCSV, "input", "", "comma-separated list of cucumber JSON files")
	fs.StringVar(&outputPath, "output", "report.html", "output HTML file")
	fs.StringVar(&title, "title", "GoCoCo Report", "report title")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	inputs := parseInputs(inputCSV, fs.Args())
	if len(inputs) == 0 {
		fmt.Fprintln(stderr, "error: no input files provided")
		fmt.Fprintln(stderr, "usage: gococo-reporting -input report.json[,report-2.json] -output report.html")
		return 2
	}

	err := gococo.WriteHTMLReport(outputPath, inputs, gococo.Options{
		Title:       title,
		GeneratedAt: time.Now(),
	})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "wrote %s\n", outputPath)
	return 0
}

func parseInputs(csv string, positional []string) []string {
	seen := map[string]struct{}{}
	var inputs []string

	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		inputs = append(inputs, value)
	}

	for _, part := range strings.Split(csv, ",") {
		add(part)
	}
	for _, arg := range positional {
		add(arg)
	}

	return inputs
}
