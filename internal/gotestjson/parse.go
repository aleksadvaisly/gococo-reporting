package gotestjson

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aleksadvaisly/gococo-reporting/internal/cucumberjson"
)

type event struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

type featureNode struct {
	pkg       string
	key       string
	name      string
	scenarios []*scenarioNode
}

type scenarioNode struct {
	key    string
	name   string
	status string
	steps  []*stepNode
}

type stepNode struct {
	key          string
	keyword      string
	name         string
	status       string
	durationNS   int64
	errorMessage string
}

// ParseFile loads a line-delimited `go test -json` stream and converts gobdd nested tests into cucumber-like features.
func ParseFile(path string) ([]cucumberjson.Feature, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	featuresByKey := map[string]*featureNode{}
	scenariosByKey := map[string]*scenarioNode{}
	stepsByKey := map[string]*stepNode{}
	var featureOrder []*featureNode

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var evt event
		if err := json.Unmarshal(line, &evt); err != nil {
			return nil, fmt.Errorf("parse %s as go test jsonl: %w", path, err)
		}

		parts := splitTestPath(evt.Test)
		if len(parts) < 2 {
			continue
		}

		featurePart := parts[1]
		if !strings.HasPrefix(featurePart, "Feature_") {
			continue
		}

		featureKey := evt.Package + "|" + featurePart
		feature := featuresByKey[featureKey]
		if feature == nil {
			feature = &featureNode{
				pkg:  evt.Package,
				key:  featureKey,
				name: decodeNamedSegment(featurePart, "Feature_"),
			}
			featuresByKey[featureKey] = feature
			featureOrder = append(featureOrder, feature)
		}

		if len(parts) >= 3 && strings.HasPrefix(parts[2], "Scenario_") {
			scenarioKey := featureKey + "|" + parts[2]
			scenario := scenariosByKey[scenarioKey]
			if scenario == nil {
				scenario = &scenarioNode{
					key:  scenarioKey,
					name: decodeNamedSegment(parts[2], "Scenario_"),
				}
				scenariosByKey[scenarioKey] = scenario
				feature.scenarios = append(feature.scenarios, scenario)
			}

			if len(parts) >= 4 {
				stepKey := scenarioKey + "|" + parts[3]
				step := stepsByKey[stepKey]
				if step == nil {
					keyword, text := decodeStep(parts[3])
					step = &stepNode{
						key:     stepKey,
						keyword: keyword,
						name:    text,
						status:  "undefined",
					}
					stepsByKey[stepKey] = step
					scenario.steps = append(scenario.steps, step)
				}

				applyStepEvent(step, evt)
				continue
			}

			applyScenarioEvent(scenario, evt)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}

	features := make([]cucumberjson.Feature, 0, len(featureOrder))
	for _, feature := range featureOrder {
		elements := make([]cucumberjson.Element, 0, len(feature.scenarios))
		for _, scenario := range feature.scenarios {
			steps := make([]cucumberjson.Step, 0, len(scenario.steps))
			for _, step := range scenario.steps {
				steps = append(steps, cucumberjson.Step{
					Keyword: step.keyword,
					Name:    step.name,
					Result: cucumberjson.Result{
						Status:       normalizeAction(step.status),
						Duration:     step.durationNS,
						ErrorMessage: strings.TrimSpace(step.errorMessage),
					},
				})
			}

			elements = append(elements, cucumberjson.Element{
				Name:    scenario.name,
				Type:    "scenario",
				Keyword: "Scenario",
				Steps:   steps,
			})
		}

		features = append(features, cucumberjson.Feature{
			Name:     feature.name,
			URI:      feature.pkg,
			Keyword:  "Feature",
			Elements: elements,
		})
	}

	if len(features) == 0 {
		return nil, fmt.Errorf("%s does not contain gobdd-style go test events", path)
	}

	return features, nil
}

func splitTestPath(testName string) []string {
	testName = strings.TrimSpace(testName)
	if testName == "" {
		return nil
	}
	return strings.Split(testName, "/")
}

func applyScenarioEvent(scenario *scenarioNode, evt event) {
	switch evt.Action {
	case "pass", "fail", "skip":
		scenario.status = evt.Action
	}
}

func applyStepEvent(step *stepNode, evt event) {
	switch evt.Action {
	case "pass", "fail", "skip":
		step.status = evt.Action
		step.durationNS = secondsToNanos(evt.Elapsed)
	case "output":
		line := sanitizeOutput(evt.Output)
		if line == "" {
			return
		}
		if step.errorMessage != "" {
			step.errorMessage += "\n"
		}
		step.errorMessage += line
	}
}

func sanitizeOutput(line string) string {
	line = strings.TrimSpace(line)
	switch {
	case line == "":
		return ""
	case strings.HasPrefix(line, "=== RUN"):
		return ""
	case strings.HasPrefix(line, "--- PASS:"):
		return ""
	case strings.HasPrefix(line, "--- FAIL:"):
		return ""
	case strings.HasPrefix(line, "--- SKIP:"):
		return ""
	default:
		return line
	}
}

func decodeNamedSegment(segment, prefix string) string {
	text := strings.TrimPrefix(segment, prefix)
	text = strings.ReplaceAll(text, "_", " ")
	return strings.TrimSpace(text)
}

func decodeStep(segment string) (string, string) {
	text := strings.ReplaceAll(strings.TrimSpace(segment), "_", " ")
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "Step", "Unnamed step"
	}

	switch fields[0] {
	case "Given", "When", "Then", "And", "But":
		return fields[0], text
	default:
		return "Step", text
	}
}

func normalizeAction(action string) string {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "pass":
		return "passed"
	case "fail":
		return "failed"
	case "skip":
		return "skipped"
	default:
		return "undefined"
	}
}

func secondsToNanos(seconds float64) int64 {
	if seconds <= 0 {
		return 0
	}
	return int64(seconds * 1_000_000_000)
}
