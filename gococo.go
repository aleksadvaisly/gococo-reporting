package gococo

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aleksadvaisly/gococo-reporting/internal/input"
	"github.com/aleksadvaisly/gococo-reporting/internal/render"
	"github.com/aleksadvaisly/gococo-reporting/internal/reportview"
)

// Options configures report generation.
type Options struct {
	Title       string
	GeneratedAt time.Time
}

// GenerateHTMLFromFiles parses cucumber JSON files and returns a single self-contained HTML report.
func GenerateHTMLFromFiles(inputPaths []string, opts Options) ([]byte, error) {
	if len(inputPaths) == 0 {
		return nil, errors.New("at least one input file is required")
	}

	features, err := input.ParseFiles(inputPaths)
	if err != nil {
		return nil, err
	}

	view := reportview.Build(features, inputPaths, opts.Title, opts.GeneratedAt)
	return render.Render(view)
}

// WriteHTML writes a report to any io.Writer.
func WriteHTML(w io.Writer, inputPaths []string, opts Options) error {
	html, err := GenerateHTMLFromFiles(inputPaths, opts)
	if err != nil {
		return err
	}

	_, err = w.Write(html)
	return err
}

// WriteHTMLReport writes a report into a file path, creating parent directories when needed.
func WriteHTMLReport(outputPath string, inputPaths []string, opts Options) error {
	if outputPath == "" {
		return errors.New("output path is required")
	}

	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return WriteHTML(file, inputPaths, opts)
}
