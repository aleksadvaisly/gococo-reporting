package reportview

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aleksadvaisly/gococo-reporting/internal/cucumberjson"
)

// Counts stores status counters.
type Counts struct {
	Passed    int
	Failed    int
	Skipped   int
	Pending   int
	Undefined int
}

// Report is a render-ready view model for the single-file report.
type Report struct {
	Title       string
	GeneratedAt string
	InputFiles  []string
	Summary     Summary
	Failures    []Failure
	Features    []Feature
}

// Summary contains top-level totals.
type Summary struct {
	Features       int
	Scenarios      int
	Steps          int
	Duration       string
	FeatureCounts  Counts
	ScenarioCounts Counts
	StepCounts     Counts
}

// Failure contains one failed step for the failures strip.
type Failure struct {
	FeatureName  string
	ScenarioName string
	StepName     string
	Status       string
	StatusLabel  string
	Message      string
	Location     string
}

// Feature contains a feature card and its nested items.
type Feature struct {
	Name          string
	URI           string
	Description   string
	Tags          []string
	Status        string
	StatusLabel   string
	Duration      string
	ScenarioCount int
	StepCount     int
	SearchText    string
	Scenarios     []Scenario
}

// Scenario is a single scenario or background block.
type Scenario struct {
	Name        string
	Type        string
	Keyword     string
	Description string
	Tags        []string
	Status      string
	StatusLabel string
	Duration    string
	Line        int
	SearchText  string
	Steps       []Step
}

// Step is a single rendered step row.
type Step struct {
	Keyword      string
	Name         string
	Status       string
	StatusLabel  string
	Duration     string
	Location     string
	ErrorMessage string
}

// Build converts parsed cucumber data into a render-ready report.
func Build(features []cucumberjson.Feature, inputFiles []string, title string, generatedAt time.Time) Report {
	if strings.TrimSpace(title) == "" {
		title = "GoCoCo Report"
	}
	if generatedAt.IsZero() {
		generatedAt = time.Now()
	}

	report := Report{
		Title:       title,
		GeneratedAt: generatedAt.Format("2006-01-02 15:04:05 MST"),
		InputFiles:  shortenPaths(inputFiles),
	}

	var totalDuration int64

	for _, sourceFeature := range features {
		featureView, featureCounts, scenarioCounts, stepCounts, featureDuration, failures := buildFeature(sourceFeature)

		report.Features = append(report.Features, featureView)
		report.Failures = append(report.Failures, failures...)
		totalDuration += featureDuration

		report.Summary.Features++
		report.Summary.FeatureCounts = addCounts(report.Summary.FeatureCounts, featureCounts)
		report.Summary.ScenarioCounts = addCounts(report.Summary.ScenarioCounts, scenarioCounts)
		report.Summary.StepCounts = addCounts(report.Summary.StepCounts, stepCounts)
	}

	report.Summary.Scenarios = totalCounts(report.Summary.ScenarioCounts)
	report.Summary.Steps = totalCounts(report.Summary.StepCounts)
	report.Summary.Duration = formatDuration(totalDuration)

	return report
}

func buildFeature(feature cucumberjson.Feature) (Feature, Counts, Counts, Counts, int64, []Failure) {
	var (
		featureDuration int64
		failures        []Failure
		scenarioCounts  Counts
		stepCounts      Counts
		scenarioViews   []Scenario
		searchParts     []string
	)

	for _, element := range feature.Elements {
		scenarioView, scenarioStatus, scenarioDuration, scenarioStepCounts, scenarioFailures := buildScenario(feature, element)
		scenarioViews = append(scenarioViews, scenarioView)
		failures = append(failures, scenarioFailures...)
		featureDuration += scenarioDuration
		stepCounts = addCounts(stepCounts, scenarioStepCounts)
		searchParts = append(searchParts, feature.Name, scenarioView.Name)

		if countsAsScenario(element.Type) {
			scenarioCounts = incrementCounts(scenarioCounts, scenarioStatus)
		}
	}

	featureStatus := highestStatusFromCounts(scenarioCounts)
	if len(feature.Elements) > 0 && featureStatus == "undefined" && len(scenarioViews) > 0 {
		featureStatus = highestStatusFromScenarios(scenarioViews)
	}

	featureCounts := incrementCounts(Counts{}, featureStatus)
	featureView := Feature{
		Name:          fallbackText(feature.Name, "Untitled feature"),
		URI:           feature.URI,
		Description:   feature.Description,
		Tags:          tagsToStrings(feature.Tags),
		Status:        featureStatus,
		StatusLabel:   statusLabel(featureStatus),
		Duration:      formatDuration(featureDuration),
		ScenarioCount: totalCounts(scenarioCounts),
		StepCount:     totalCounts(stepCounts),
		SearchText:    strings.ToLower(strings.Join(searchParts, " ")),
		Scenarios:     scenarioViews,
	}

	return featureView, featureCounts, scenarioCounts, stepCounts, featureDuration, failures
}

func buildScenario(feature cucumberjson.Feature, element cucumberjson.Element) (Scenario, string, int64, Counts, []Failure) {
	var (
		duration    int64
		stepCounts  Counts
		failures    []Failure
		stepViews   []Step
		searchParts []string
	)

	status := "undefined"
	if len(element.Steps) > 0 {
		status = "passed"
	}

	for _, step := range element.Steps {
		stepStatus := normalizeStatus(step.Result.Status)
		duration += step.Result.Duration
		stepCounts = incrementCounts(stepCounts, stepStatus)
		status = higherStatus(status, stepStatus)
		searchParts = append(searchParts, step.Name, step.Match.Location)

		stepViews = append(stepViews, Step{
			Keyword:      fallbackText(step.Keyword, "Step"),
			Name:         fallbackText(step.Name, "Unnamed step"),
			Status:       stepStatus,
			StatusLabel:  statusLabel(stepStatus),
			Duration:     formatDuration(step.Result.Duration),
			Location:     step.Match.Location,
			ErrorMessage: step.Result.ErrorMessage,
		})

		if stepStatus == "failed" {
			failures = append(failures, Failure{
				FeatureName:  fallbackText(feature.Name, "Untitled feature"),
				ScenarioName: fallbackText(element.Name, "Unnamed scenario"),
				StepName:     fallbackText(step.Name, "Unnamed step"),
				Status:       stepStatus,
				StatusLabel:  statusLabel(stepStatus),
				Message:      fallbackText(step.Result.ErrorMessage, "Step failed without an error message."),
				Location:     step.Match.Location,
			})
		}
	}

	scenario := Scenario{
		Name:        fallbackText(element.Name, "Unnamed scenario"),
		Type:        fallbackText(element.Type, "scenario"),
		Keyword:     fallbackText(element.Keyword, "Scenario"),
		Description: element.Description,
		Tags:        tagsToStrings(element.Tags),
		Status:      status,
		StatusLabel: statusLabel(status),
		Duration:    formatDuration(duration),
		Line:        0,
		SearchText:  strings.ToLower(strings.Join(searchParts, " ")),
		Steps:       stepViews,
	}

	return scenario, status, duration, stepCounts, failures
}

func countsAsScenario(kind string) bool {
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "", "scenario", "scenario_outline":
		return true
	default:
		return false
	}
}

func incrementCounts(counts Counts, status string) Counts {
	switch normalizeStatus(status) {
	case "passed":
		counts.Passed++
	case "failed":
		counts.Failed++
	case "skipped":
		counts.Skipped++
	case "pending":
		counts.Pending++
	default:
		counts.Undefined++
	}
	return counts
}

func addCounts(left, right Counts) Counts {
	return Counts{
		Passed:    left.Passed + right.Passed,
		Failed:    left.Failed + right.Failed,
		Skipped:   left.Skipped + right.Skipped,
		Pending:   left.Pending + right.Pending,
		Undefined: left.Undefined + right.Undefined,
	}
}

func totalCounts(counts Counts) int {
	return counts.Passed + counts.Failed + counts.Skipped + counts.Pending + counts.Undefined
}

func normalizeStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "passed":
		return "passed"
	case "failed", "ambiguous":
		return "failed"
	case "skipped":
		return "skipped"
	case "pending":
		return "pending"
	default:
		return "undefined"
	}
}

func statusWeight(status string) int {
	switch normalizeStatus(status) {
	case "failed":
		return 5
	case "undefined":
		return 4
	case "pending":
		return 3
	case "skipped":
		return 2
	case "passed":
		return 1
	default:
		return 0
	}
}

func higherStatus(current, candidate string) string {
	if statusWeight(candidate) > statusWeight(current) {
		return normalizeStatus(candidate)
	}
	return normalizeStatus(current)
}

func highestStatusFromCounts(counts Counts) string {
	switch {
	case counts.Failed > 0:
		return "failed"
	case counts.Undefined > 0:
		return "undefined"
	case counts.Pending > 0:
		return "pending"
	case counts.Skipped > 0:
		return "skipped"
	case counts.Passed > 0:
		return "passed"
	default:
		return "undefined"
	}
}

func highestStatusFromScenarios(scenarios []Scenario) string {
	status := "undefined"
	for _, scenario := range scenarios {
		status = higherStatus(status, scenario.Status)
	}
	return status
}

func statusLabel(status string) string {
	switch normalizeStatus(status) {
	case "passed":
		return "Passed"
	case "failed":
		return "Failed"
	case "skipped":
		return "Skipped"
	case "pending":
		return "Pending"
	default:
		return "Undefined"
	}
}

func formatDuration(ns int64) string {
	if ns <= 0 {
		return "0s"
	}

	d := time.Duration(ns)
	switch {
	case d < time.Millisecond:
		return d.Round(time.Microsecond).String()
	case d < time.Second:
		return d.Round(time.Millisecond).String()
	case d < time.Minute:
		return d.Round(10 * time.Millisecond).String()
	default:
		minutes := d / time.Minute
		seconds := (d % time.Minute).Round(time.Second)
		return fmt.Sprintf("%dm %02ds", minutes, seconds/time.Second)
	}
}

func tagsToStrings(tags []cucumberjson.Tag) []string {
	if len(tags) == 0 {
		return nil
	}

	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		name := strings.TrimSpace(tag.Name)
		if name != "" {
			result = append(result, name)
		}
	}
	return result
}

func fallbackText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func shortenPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}

	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		base := filepath.Base(path)
		if base == "." || base == string(filepath.Separator) {
			result = append(result, path)
			continue
		}
		result = append(result, base)
	}
	return result
}
