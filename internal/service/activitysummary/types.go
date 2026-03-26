package activitysummary

type SummaryResponse struct {
	Summary  ActivitySummary `json:"summary"`
	Markdown string          `json:"markdown"`
}

type DailySummaryResponse struct {
	Date          string            `json:"date"`
	Timezone      string            `json:"timezone"`
	ActivityCount int               `json:"activity_count"`
	DailySummary  SummaryResponse   `json:"daily_summary"`
	Activities    []SummaryResponse `json:"activities"`
}

type ActivitySummary struct {
	Source                    string         `json:"source"`
	SourceName                string         `json:"source_name,omitempty"`
	ActivityID                string         `json:"activity_id,omitempty"`
	Name                      string         `json:"name,omitempty"`
	SportType                 string         `json:"sport_type,omitempty"`
	StartTime                 string         `json:"start_time,omitempty"`
	EndTime                   string         `json:"end_time,omitempty"`
	DurationSeconds           *float64       `json:"duration_seconds,omitempty"`
	MovingSeconds             *float64       `json:"moving_seconds,omitempty"`
	PauseSeconds              *float64       `json:"pause_seconds,omitempty"`
	DistanceMeters            *float64       `json:"distance_meters,omitempty"`
	Calories                  *float64       `json:"calories,omitempty"`
	AscentMeters              *float64       `json:"ascent_meters,omitempty"`
	DescentMeters             *float64       `json:"descent_meters,omitempty"`
	AverageHeartRate          *float64       `json:"average_heart_rate,omitempty"`
	MaxHeartRate              *float64       `json:"max_heart_rate,omitempty"`
	AverageCadence            *float64       `json:"average_cadence,omitempty"`
	MaxCadence                *float64       `json:"max_cadence,omitempty"`
	AveragePower              *float64       `json:"average_power,omitempty"`
	MaxPower                  *float64       `json:"max_power,omitempty"`
	AverageSpeedMPS           *float64       `json:"average_speed_mps,omitempty"`
	MaxSpeedMPS               *float64       `json:"max_speed_mps,omitempty"`
	AveragePaceSecPerKM       *float64       `json:"average_pace_sec_per_km,omitempty"`
	AverageMovingPaceSecPerKM *float64       `json:"average_moving_pace_sec_per_km,omitempty"`
	BestPaceSecPerKM          *float64       `json:"best_pace_sec_per_km,omitempty"`
	TrainingLoad              *float64       `json:"training_load,omitempty"`
	TrainingEffect            *float64       `json:"training_effect,omitempty"`
	AverageTemperature        *float64       `json:"average_temperature,omitempty"`
	AverageStrideLengthMeters *float64       `json:"average_stride_length_meters,omitempty"`
	StepCount                 *float64       `json:"step_count,omitempty"`
	RecordCount               int            `json:"record_count,omitempty"`
	LapCount                  int            `json:"lap_count,omitempty"`
	Device                    string         `json:"device,omitempty"`
	Highlights                []string       `json:"highlights,omitempty"`
	Laps                      []LapSummary   `json:"laps,omitempty"`
	RawLatestActivity         map[string]any `json:"raw_latest_activity,omitempty"`
}

type LapSummary struct {
	Index               int      `json:"index"`
	StartTime           string   `json:"start_time,omitempty"`
	DurationSeconds     *float64 `json:"duration_seconds,omitempty"`
	MovingSeconds       *float64 `json:"moving_seconds,omitempty"`
	DistanceMeters      *float64 `json:"distance_meters,omitempty"`
	AverageHeartRate    *float64 `json:"average_heart_rate,omitempty"`
	MaxHeartRate        *float64 `json:"max_heart_rate,omitempty"`
	AverageCadence      *float64 `json:"average_cadence,omitempty"`
	MaxCadence          *float64 `json:"max_cadence,omitempty"`
	AveragePower        *float64 `json:"average_power,omitempty"`
	MaxPower            *float64 `json:"max_power,omitempty"`
	AverageSpeedMPS     *float64 `json:"average_speed_mps,omitempty"`
	MaxSpeedMPS         *float64 `json:"max_speed_mps,omitempty"`
	AveragePaceSecPerKM *float64 `json:"average_pace_sec_per_km,omitempty"`
}
