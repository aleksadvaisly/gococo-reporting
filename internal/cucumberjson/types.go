package cucumberjson

// Feature mirrors the subset of cucumber JSON needed by the report generator.
type Feature struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	URI         string    `json:"uri"`
	Description string    `json:"description"`
	Keyword     string    `json:"keyword"`
	Elements    []Element `json:"elements"`
	Tags        []Tag     `json:"tags"`
}

// Element is typically a scenario or background in cucumber JSON.
type Element struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Keyword        string `json:"keyword"`
	Description    string `json:"description"`
	StartTimestamp string `json:"start_timestamp"`
	Steps          []Step `json:"steps"`
	Tags           []Tag  `json:"tags"`
}

// Step contains an executed step and its result.
type Step struct {
	Name    string `json:"name"`
	Keyword string `json:"keyword"`
	Line    int    `json:"line"`
	Match   Match  `json:"match"`
	Result  Result `json:"result"`
}

// Match stores the step location.
type Match struct {
	Location string `json:"location"`
}

// Result stores execution result details for a step.
type Result struct {
	Status       string `json:"status"`
	Duration     int64  `json:"duration"`
	ErrorMessage string `json:"error_message"`
}

// Tag represents a feature or scenario tag.
type Tag struct {
	Name string `json:"name"`
}
