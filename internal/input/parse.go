package input

import (
	"fmt"

	"github.com/aleksadvaisly/gococo-reporting/internal/cucumberjson"
	"github.com/aleksadvaisly/gococo-reporting/internal/gotestjson"
)

// ParseFiles loads supported report inputs and merges them into a single cucumber-like slice.
func ParseFiles(paths []string) ([]cucumberjson.Feature, error) {
	var all []cucumberjson.Feature

	for _, path := range paths {
		features, err := cucumberjson.ParseFile(path)
		if err != nil {
			features, err = gotestjson.ParseFile(path)
			if err != nil {
				return nil, fmt.Errorf("unsupported input %s: %w", path, err)
			}
		}

		all = append(all, features...)
	}

	if len(all) == 0 {
		return nil, fmt.Errorf("input files did not contain any supported report data")
	}

	return all, nil
}
