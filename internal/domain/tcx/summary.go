package tcx

// Summary represents a parsed TCX activity summary.
type Summary struct {
	ID          string  `json:"id"`
	Filename    string  `json:"filename"`
	Duration    int     `json:"duration"`
	Distance    float64 `json:"distance"`
	Calories    float64 `json:"calories"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	SportType   string  `json:"sport_type"`
	AverageHR   int     `json:"average_hr"`
	MaxHR       int     `json:"max_hr"`
	TotalAscent float64 `json:"total_ascent"`
	CreatedAt   string  `json:"created_at"`
}
