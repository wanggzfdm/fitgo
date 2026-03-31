package traininganalysis

import (
	"testing"

	"coros-fit-mcp/internal/service/coros"
)

type stubCorosService struct {
	accountData         *coros.AccountQueryData
	dashboardData       *coros.DashboardQueryData
	dashboardDetailData *coros.DashboardDetailQueryData
	activityList        map[string]interface{}
	sportSummary        *coros.SportsSummaryResult
}

func (s *stubCorosService) Login() (string, error) { return "token", nil }
func (s *stubCorosService) AccountQuery() (*coros.AccountQueryData, error) {
	return s.accountData, nil
}
func (s *stubCorosService) DashboardQuery() (*coros.DashboardQueryData, error) {
	return s.dashboardData, nil
}
func (s *stubCorosService) DashboardDetailQuery() (*coros.DashboardDetailQueryData, error) {
	return s.dashboardDetailData, nil
}
func (s *stubCorosService) SportsSummary(labelId, sportType string) (*coros.SportsSummaryResult, error) {
	return s.sportSummary, nil
}
func (s *stubCorosService) ActivityList(size, pageNumber, modeList int) (map[string]interface{}, error) {
	return s.activityList, nil
}
func (s *stubCorosService) ActivityListByModeList(size, pageNumber int, modeList string) (map[string]interface{}, error) {
	return s.activityList, nil
}

func TestAnalyzeTrainingStatusBuildsThreeDayPlan(t *testing.T) {
	service := &stubCorosService{
		accountData: &coros.AccountQueryData{
			Weight: 70, Stature: 178, Birthday: 19990202, MaxHr: 195, Rhr: 49,
			ZoneData: coros.AccountZoneData{
				LTHR: 172, LTSP: 307,
				LTHRZone: []coros.ZoneBoundary{{Hr: 138}, {Hr: 155}, {Hr: 163}, {Hr: 175}, {Hr: 182}},
				LTSPZone: []coros.PaceZoneBoundary{{Pace: 439}, {Pace: 372}, {Pace: 334}, {Pace: 307}, {Pace: 301}, {Pace: 271}},
			},
		},
		dashboardData: &coros.DashboardQueryData{SummaryInfo: coros.DashboardSummaryInfo{StaminaLevel: 74.1}},
		dashboardDetailData: &coros.DashboardDetailQueryData{
			SummaryInfo: coros.DashboardDetailSummaryInfo{
				ATI: 67, CTI: 64, TrainingLoadRatio: 1.04, TiredRate: 44, TiredRateNew: 3, TiredRateNewState: 2,
			},
			SportDataList: []coros.RecentSportData{{HappenDay: 20260325, LabelID: "1", Distance: 5273.57, Duration: 1983, AvgPace: 376, AvgHeartRate: 138, TrainingLoad: 46}},
			CurrentWeekRecord: coros.WeekRecord{
				DistanceRecord: coros.AggregateRecord{TotalValue: 18426.2},
				DurationRecord: coros.AggregateRecord{TotalValue: 6396},
				TLRecord:       coros.AggregateRecord{TotalValue: 289},
			},
			DetailList: []coros.TrainingTrendPoint{{HappenDay: 20260325, TrainingLoad: 60, ATI: 78, CTI: 66, TiredRate: 51, TiredRateNew: 12}},
		},
		activityList: map[string]interface{}{
			"data": map[string]interface{}{
				"dataList": []interface{}{map[string]interface{}{
					"labelId": "1", "sportType": 100.0, "name": "晚跑", "startTime": 1774435524.0, "endTime": 1774437507.0,
					"totalTime": 1983.0, "workoutTime": 1983.0, "distance": 5273.57, "adjustedPace": 376.0,
					"avgHr": 138.0, "avgCadence": 174.0, "avgPower": 190.0, "trainingLoad": 46.0,
				}},
			},
		},
		sportSummary: &coros.SportsSummaryResult{Summary: map[string]interface{}{"maxHr": 150.0}},
	}

	resp, err := AnalyzeTrainingStatus(service, "")
	if err != nil {
		t.Fatalf("AnalyzeTrainingStatus returned error: %v", err)
	}
	if resp.Data.Recommendation.Tomorrow == "" {
		t.Fatalf("expected tomorrow recommendation, got %#v", resp.Data.Recommendation)
	}
	if len(resp.Data.NextPlan.Days) != 3 {
		t.Fatalf("expected 3-day plan, got %#v", resp.Data.NextPlan)
	}
	if resp.Data.NextPlan.Days[0].Type == "" {
		t.Fatalf("expected day 1 plan type, got %#v", resp.Data.NextPlan.Days[0])
	}
	if resp.Markdown == "" {
		t.Fatalf("expected markdown output")
	}
}
