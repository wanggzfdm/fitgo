package tcx

import "fmt"

// Parse parses TCX content into a Summary.
func Parse(content []byte) (*Summary, error) {
	// TODO: implement TCX XML parsing.
	if len(content) == 0 {
		return nil, fmt.Errorf("empty content")
	}

	return &Summary{
		ID:        "example-id",
		Filename:  "example.tcx",
		Duration:  3600,
		Distance:  10000.0,
		Calories:  500.0,
		SportType: "Running",
		AverageHR: 140,
		MaxHR:     180,
	}, nil
}

// Validate performs a basic validation on TCX content.
func Validate(content []byte) error {
	if len(content) == 0 {
		return fmt.Errorf("empty content")
	}
	return nil
}
