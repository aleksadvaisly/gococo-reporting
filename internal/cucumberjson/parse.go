package cucumberjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ParseFiles loads and merges cucumber JSON arrays from multiple files.
func ParseFiles(paths []string) ([]Feature, error) {
	if len(paths) == 0 {
		return nil, errors.New("no cucumber JSON files provided")
	}

	var all []Feature

	for _, path := range paths {
		features, err := ParseFile(path)
		if err != nil {
			return nil, err
		}
		all = append(all, features...)
	}

	if len(all) == 0 {
		return nil, errors.New("input files did not contain any features")
	}

	return all, nil
}

// ParseFile reads a single godog/cucumber JSON report file.
func ParseFile(path string) ([]Feature, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("%s is empty", path)
	}

	var features []Feature
	if err := json.Unmarshal(trimmed, &features); err != nil {
		return nil, fmt.Errorf("parse %s as cucumber JSON array: %w", path, err)
	}

	for index := range features {
		normalizeFeature(&features[index])
	}

	return features, nil
}

func normalizeFeature(feature *Feature) {
	feature.Name = strings.TrimSpace(feature.Name)
	feature.Description = strings.TrimSpace(feature.Description)
	feature.Keyword = strings.TrimSpace(feature.Keyword)

	for i := range feature.Tags {
		feature.Tags[i].Name = strings.TrimSpace(feature.Tags[i].Name)
	}

	for i := range feature.Elements {
		normalizeElement(&feature.Elements[i])
	}
}

func normalizeElement(element *Element) {
	element.Name = strings.TrimSpace(element.Name)
	element.Type = strings.TrimSpace(strings.ToLower(element.Type))
	element.Keyword = strings.TrimSpace(element.Keyword)
	element.Description = strings.TrimSpace(element.Description)

	for i := range element.Tags {
		element.Tags[i].Name = strings.TrimSpace(element.Tags[i].Name)
	}

	for i := range element.Steps {
		normalizeStep(&element.Steps[i])
	}
}

func normalizeStep(step *Step) {
	step.Name = strings.TrimSpace(step.Name)
	step.Keyword = strings.TrimSpace(step.Keyword)
	step.Match.Location = strings.TrimSpace(step.Match.Location)
	step.Result.Status = strings.TrimSpace(strings.ToLower(step.Result.Status))
	step.Result.ErrorMessage = strings.TrimSpace(step.Result.ErrorMessage)
}
