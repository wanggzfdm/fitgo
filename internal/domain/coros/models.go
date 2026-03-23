package coros

// Summary represents a Coros activity summary.
type Summary struct {
	ID        string  `json:"id"`
	Date      string  `json:"date"`
	Duration  int     `json:"duration"`
	Distance  float64 `json:"distance"`
	Calories  float64 `json:"calories"`
	SportType string  `json:"sport_type"`
}

// SportsSummary represents detailed activity summary data.
type SportsSummary struct {
	LapList []map[string]interface{} `json:"lapList"`
	Summary map[string]interface{}   `json:"summary"`
}
