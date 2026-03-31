package activitysummary

import (
	"strings"
	"testing"
	"time"

	"coros-fit-mcp/internal/service/coros"
	fitlib "github.com/tormoder/fit"
)

type stubCorosService struct {
	list    map[string]interface{}
	pages   map[int]map[string]interface{}
	detail  *coros.SportsSummaryResult
	details map[string]*coros.SportsSummaryResult
}

func (s *stubCorosService) Login() (string, error) { return "token", nil }
func (s *stubCorosService) AccountQuery() (*coros.AccountQueryData, error) {
	return nil, nil
}
func (s *stubCorosService) DashboardQuery() (*coros.DashboardQueryData, error) {
	return nil, nil
}
func (s *stubCorosService) DashboardDetailQuery() (*coros.DashboardDetailQueryData, error) {
	return nil, nil
}
func (s *stubCorosService) SportsSummary(labelId, sportType string) (*coros.SportsSummaryResult, error) {
	if s.details != nil {
		return s.details[labelIDKey(labelId, sportType)], nil
	}
	return s.detail, nil
}
func (s *stubCorosService) ActivityList(size, pageNumber, modeList int) (map[string]interface{}, error) {
	if s.pages != nil {
		return s.pages[pageNumber], nil
	}
	return s.list, nil
}
func (s *stubCorosService) ActivityListByModeList(size, pageNumber int, modeList string) (map[string]interface{}, error) {
	if s.pages != nil {
		return s.pages[pageNumber], nil
	}
	return s.list, nil
}

func labelIDKey(labelID, sportType string) string {
	return labelID + ":" + sportType
}

func TestSummarizeLatestCorosActivity(t *testing.T) {
	service := &stubCorosService{
		list: map[string]interface{}{
			"data": map[string]interface{}{
				"dataList": []interface{}{
					map[string]interface{}{
						"labelId":      "abc123",
						"sportType":    100.0,
						"name":         "晨跑",
						"startTime":    1773918916.0,
						"endTime":      1773923688.0,
						"totalTime":    4772.0,
						"workoutTime":  4772.0,
						"distance":     12027.47,
						"adjustedPace": 396.0,
						"avgSpeed":     396.78,
						"bestKm":       380.0,
						"avgHr":        143.0,
						"avgCadence":   169.0,
						"avgPower":     185.0,
						"calorie":      951707.0,
						"trainingLoad": 123.0,
						"step":         13658.0,
						"device":       "COROS APEX 4 46mm",
					},
				},
			},
		},
		detail: &coros.SportsSummaryResult{
			Summary: map[string]interface{}{
				"maxHr":      161.0,
				"maxCadence": 179.0,
				"maxPower":   240.0,
				"descent":    16.0,
			},
			LapList: []map[string]interface{}{
				{
					"distance":   1000.0,
					"sportTime":  383.24,
					"avgHr":      145.0,
					"avgCadence": 176.0,
					"avgPower":   240.0,
					"avgPace":    383.24,
				},
			},
		},
	}

	summary, err := SummarizeLatestCorosActivity(service)
	if err != nil {
		t.Fatalf("SummarizeLatestCorosActivity returned error: %v", err)
	}

	if summary.Source != "coros" {
		t.Fatalf("expected source coros, got %q", summary.Source)
	}
	if summary.Name != "晨跑" {
		t.Fatalf("expected activity name, got %q", summary.Name)
	}
	if summary.DistanceMeters == nil || *summary.DistanceMeters != 12027.47 {
		t.Fatalf("expected distance 12027.47, got %#v", summary.DistanceMeters)
	}
	if summary.AveragePaceSecPerKM == nil || *summary.AveragePaceSecPerKM != 396 {
		t.Fatalf("expected pace 396, got %#v", summary.AveragePaceSecPerKM)
	}
	if summary.AverageMovingPaceSecPerKM == nil || *summary.AverageMovingPaceSecPerKM != 396.78 {
		t.Fatalf("expected moving pace 396.78, got %#v", summary.AverageMovingPaceSecPerKM)
	}
	if summary.Calories == nil || *summary.Calories != 951.707 {
		t.Fatalf("expected calories 951.707, got %#v", summary.Calories)
	}
	if summary.StepCount == nil || *summary.StepCount != 13658 {
		t.Fatalf("expected step count 13658, got %#v", summary.StepCount)
	}
	if summary.AverageStrideLengthMeters == nil {
		t.Fatalf("expected stride length")
	}
	if len(summary.Laps) != 1 {
		t.Fatalf("expected 1 lap, got %d", len(summary.Laps))
	}
}

func TestSummarizeFITActivity(t *testing.T) {
	session := &fitlib.SessionMsg{
		StartTime:           time.Date(2026, 3, 21, 7, 0, 0, 0, time.UTC),
		Timestamp:           time.Date(2026, 3, 21, 7, 45, 0, 0, time.UTC),
		Sport:               fitlib.SportRunning,
		TotalElapsedTime:    2700000,
		TotalTimerTime:      2640000,
		TotalDistance:       1000000,
		TotalCalories:       680,
		AvgHeartRate:        152,
		MaxHeartRate:        176,
		AvgCadence:          180,
		AvgPower:            268,
		AvgSpeed:            3780,
		MaxSpeed:            5000,
		TotalAscent:         120,
		TrainingStressScore: 86,
	}

	lap := &fitlib.LapMsg{
		StartTime:        session.StartTime,
		TotalElapsedTime: 540000,
		TotalTimerTime:   535000,
		TotalDistance:    200000,
		AvgHeartRate:     148,
		MaxHeartRate:     160,
		AvgCadence:       178,
		AvgPower:         255,
		AvgSpeed:         3730,
	}

	activity := &fitlib.ActivityFile{
		Sessions: []*fitlib.SessionMsg{session},
		Laps:     []*fitlib.LapMsg{lap},
		Records:  []*fitlib.RecordMsg{{}, {}},
	}

	file := &fitlib.File{}
	file.FileId.Type = fitlib.FileTypeActivity
	file.FileId.ProductName = "COROS PACE"

	summary := summarizeFITActivity(file, activity, "run.fit")
	if summary.Source != "fit" {
		t.Fatalf("expected source fit, got %q", summary.Source)
	}
	if summary.SportType != "Running" {
		t.Fatalf("expected Running, got %q", summary.SportType)
	}
	if summary.DistanceMeters == nil || *summary.DistanceMeters != 10000 {
		t.Fatalf("expected 10000m, got %#v", summary.DistanceMeters)
	}
	if summary.RecordCount != 2 {
		t.Fatalf("expected 2 records, got %d", summary.RecordCount)
	}
	if summary.AveragePaceSecPerKM == nil {
		t.Fatalf("expected average pace")
	}
}

func TestSummarizeCorosDailyActivities(t *testing.T) {
	service := &stubCorosService{
		pages: map[int]map[string]interface{}{
			1: {
				"data": map[string]interface{}{
					"dataList": []interface{}{
						map[string]interface{}{
							"labelId":      "warmup",
							"sportType":    100.0,
							"name":         "热身跑",
							"startTime":    1773916800.0,
							"endTime":      1773918600.0,
							"totalTime":    1800.0,
							"workoutTime":  1800.0,
							"distance":     3000.0,
							"adjustedPace": 360.0,
							"avgSpeed":     362.0,
							"bestKm":       340.0,
							"avgHr":        130.0,
							"avgCadence":   165.0,
							"avgPower":     170.0,
							"calorie":      220000.0,
							"trainingLoad": 30.0,
							"step":         3600.0,
						},
						map[string]interface{}{
							"labelId":      "tempo",
							"sportType":    100.0,
							"name":         "强度跑",
							"startTime":    1773920400.0,
							"endTime":      1773923100.0,
							"totalTime":    2700.0,
							"workoutTime":  2700.0,
							"distance":     6000.0,
							"adjustedPace": 390.0,
							"avgSpeed":     392.0,
							"bestKm":       355.0,
							"avgHr":        150.0,
							"avgCadence":   172.0,
							"avgPower":     205.0,
							"calorie":      430000.0,
							"trainingLoad": 60.0,
							"step":         7200.0,
						},
						map[string]interface{}{
							"labelId":      "prev",
							"sportType":    100.0,
							"name":         "前一天跑步",
							"startTime":    1773830400.0,
							"endTime":      1773832200.0,
							"totalTime":    1800.0,
							"workoutTime":  1800.0,
							"distance":     3000.0,
							"adjustedPace": 360.0,
							"avgSpeed":     362.0,
							"bestKm":       340.0,
						},
					},
				},
			},
			2: {
				"data": map[string]interface{}{
					"dataList": []interface{}{},
				},
			},
		},
		details: map[string]*coros.SportsSummaryResult{
			labelIDKey("warmup", "100"): {
				Summary: map[string]interface{}{
					"maxHr":      142.0,
					"maxCadence": 170.0,
					"maxPower":   200.0,
				},
			},
			labelIDKey("tempo", "100"): {
				Summary: map[string]interface{}{
					"maxHr":      168.0,
					"maxCadence": 178.0,
					"maxPower":   260.0,
				},
			},
		},
	}

	response, err := SummarizeCorosDailyActivities(service, "2026-03-19")
	if err != nil {
		t.Fatalf("SummarizeCorosDailyActivities returned error: %v", err)
	}

	if response.ActivityCount != 2 {
		t.Fatalf("expected 2 activities, got %d", response.ActivityCount)
	}
	if response.Date != "2026-03-19" {
		t.Fatalf("expected date 2026-03-19, got %s", response.Date)
	}
	if response.DailySummary.Summary.DistanceMeters == nil || *response.DailySummary.Summary.DistanceMeters != 9000 {
		t.Fatalf("expected daily distance 9000, got %#v", response.DailySummary.Summary.DistanceMeters)
	}
	if response.DailySummary.Summary.StepCount == nil || *response.DailySummary.Summary.StepCount != 10800 {
		t.Fatalf("expected daily steps 10800, got %#v", response.DailySummary.Summary.StepCount)
	}
	if response.DailySummary.Summary.BestPaceSecPerKM == nil || *response.DailySummary.Summary.BestPaceSecPerKM != 340 {
		t.Fatalf("expected best pace 340, got %#v", response.DailySummary.Summary.BestPaceSecPerKM)
	}
	if response.Activities[0].Summary.Name != "热身跑" || response.Activities[1].Summary.Name != "强度跑" {
		t.Fatalf("expected activities sorted by start time, got %#v", response.Activities)
	}
}

func TestTrailSummaryFormatting(t *testing.T) {
	summary := ActivitySummary{
		Source:                    "coros",
		SourceName:                "coros_latest_activity",
		Name:                      "山径长距离",
		SportType:                 "越野跑",
		StartTime:                 "2026-03-28T10:45:06+08:00",
		DurationSeconds:           ptr(17318.0),
		MovingSeconds:             ptr(17317.0),
		DistanceMeters:            ptr(28084.6),
		AscentMeters:              ptr(2037.0),
		DescentMeters:             ptr(2013.0),
		AverageHeartRate:          ptr(160.0),
		MaxHeartRate:              ptr(178.0),
		AveragePower:              ptr(170.0),
		AverageCadence:            ptr(136.0),
		AverageStrideLengthMeters: ptr(0.74),
		AverageMovingPaceSecPerKM: ptr(616.6),
		BestPaceSecPerKM:          ptr(326.0),
		TrainingLoad:              ptr(896.0),
		Calories:                  ptr(3165.7),
		StepCount:                 ptr(38172.0),
	}
	enrichTerrainMetrics(&summary)

	response := ToResponse(summary)

	if response.Summary.ElevationGainPerKM == nil || *response.Summary.ElevationGainPerKM < 72 || *response.Summary.ElevationGainPerKM > 73 {
		t.Fatalf("expected elevation gain per km around 72.5, got %#v", response.Summary.ElevationGainPerKM)
	}
	if response.Summary.VerticalAscentPerHour == nil || *response.Summary.VerticalAscentPerHour < 423 || *response.Summary.VerticalAscentPerHour > 424 {
		t.Fatalf("expected vertical ascent per hour around 423.5, got %#v", response.Summary.VerticalAscentPerHour)
	}
	if !strings.Contains(response.Markdown, "📍 越野概况") {
		t.Fatalf("expected trail markdown section, got %s", response.Markdown)
	}
	if !strings.Contains(response.Markdown, "- 累计爬升：2037 m") {
		t.Fatalf("expected ascent in markdown, got %s", response.Markdown)
	}
	if !strings.Contains(response.Markdown, "- 爬升效率：424 m/h") {
		t.Fatalf("expected vertical ascent efficiency in markdown, got %s", response.Markdown)
	}
	if len(response.Summary.Highlights) == 0 || !strings.Contains(response.Summary.Highlights[0], "累计爬升 2037 m") {
		t.Fatalf("expected trail highlight to include ascent, got %#v", response.Summary.Highlights)
	}
	if strings.Contains(response.Markdown, "📍 基本信息") {
		t.Fatalf("expected trail summary to avoid road template, got %s", response.Markdown)
	}
}
