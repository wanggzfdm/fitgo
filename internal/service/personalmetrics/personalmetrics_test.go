package personalmetrics

import (
	"testing"

	"coros-fit-mcp/internal/service/coros"
)

type stubCorosService struct {
	accountData         *coros.AccountQueryData
	dashboardData       *coros.DashboardQueryData
	dashboardDetailData *coros.DashboardDetailQueryData
}

func (s *stubCorosService) Login() (string, error) { return "token", nil }
func (s *stubCorosService) SportsSummary(labelId, sportType string) (*coros.SportsSummaryResult, error) {
	return nil, nil
}
func (s *stubCorosService) ActivityList(size, pageNumber, modeList int) (map[string]interface{}, error) {
	return nil, nil
}
func (s *stubCorosService) AccountQuery() (*coros.AccountQueryData, error) {
	return s.accountData, nil
}
func (s *stubCorosService) DashboardQuery() (*coros.DashboardQueryData, error) {
	return s.dashboardData, nil
}
func (s *stubCorosService) DashboardDetailQuery() (*coros.DashboardDetailQueryData, error) {
	return s.dashboardDetailData, nil
}

func TestGetRunnerProfile(t *testing.T) {
	service := &stubCorosService{
		accountData: &coros.AccountQueryData{
			Weight:      70,
			Stature:     178,
			Birthday:    19990202,
			Sex:         0,
			CountryCode: "CN",
			MaxHr:       195,
			Rhr:         49,
			HrZoneType:  3,
			ZoneData: coros.AccountZoneData{
				LTHR:     172,
				LTSP:     307,
				LTHRZone: []coros.ZoneBoundary{{Hr: 138, Ratio: 80}},
				LTSPZone: []coros.PaceZoneBoundary{{Pace: 439, Ratio: 69.9}},
			},
		},
	}

	profile, err := GetRunnerProfile(service)
	if err != nil {
		t.Fatalf("GetRunnerProfile returned error: %v", err)
	}

	if profile.Birthday != "1999-02-02" {
		t.Fatalf("expected formatted birthday, got %q", profile.Birthday)
	}
	if profile.LactateThresholdPace != "5'07\"/km" {
		t.Fatalf("expected threshold pace text, got %q", profile.LactateThresholdPace)
	}
	if len(profile.PaceZonesLTSP) != 1 || profile.PaceZonesLTSP[0].PaceText != "7'19\"/km" {
		t.Fatalf("unexpected pace zones: %#v", profile.PaceZonesLTSP)
	}
}

func TestGetTrainingZonesIncludesGuidance(t *testing.T) {
	service := &stubCorosService{
		accountData: &coros.AccountQueryData{
			HrZoneType: 3,
			ZoneData: coros.AccountZoneData{
				LTHRZone: []coros.ZoneBoundary{
					{Hr: 138, Ratio: 80},
					{Hr: 155, Ratio: 90},
					{Hr: 163, Ratio: 95},
					{Hr: 175, Ratio: 102},
					{Hr: 182, Ratio: 106},
				},
				LTSPZone: []coros.PaceZoneBoundary{
					{Pace: 439, Ratio: 69.9},
					{Pace: 372, Ratio: 82.5},
					{Pace: 334, Ratio: 91.9},
					{Pace: 307, Ratio: 100.0},
					{Pace: 301, Ratio: 101.9},
					{Pace: 271, Ratio: 113.2},
				},
			},
		},
	}

	zones, err := GetTrainingZones(service)
	if err != nil {
		t.Fatalf("GetTrainingZones returned error: %v", err)
	}

	if zones.Guidance.Recovery.HeartRate == "" || zones.Guidance.Recovery.Pace == "" {
		t.Fatalf("expected recovery guidance, got %#v", zones.Guidance.Recovery)
	}
	if zones.Guidance.Interval.Description == "" {
		t.Fatalf("expected interval guidance, got %#v", zones.Guidance.Interval)
	}
}

func TestGetTrainingDashboard(t *testing.T) {
	service := &stubCorosService{
		dashboardData: &coros.DashboardQueryData{
			SummaryInfo: coros.DashboardSummaryInfo{
				StaminaLevel:                  74.1,
				AerobicEnduranceScore:         74.1,
				LactateThresholdCapacityScore: 73,
				AnaerobicEnduranceScore:       73,
				AnaerobicCapacityScore:        72.9,
				StaminaLevelRanking:           45.91,
				LTHR:                          172,
				LTSP:                          307,
				FitnessMaxHr:                  195,
				Rhr:                           49,
				LTHRZone:                      []coros.ZoneBoundary{{Hr: 138, Ratio: 80}},
				LTSPZone:                      []coros.PaceZoneBoundary{{Pace: 439, Ratio: 69.9}},
				RecoveryPct:                   100,
				RecoveryState:                 4,
				FullRecoveryHours:             0,
				SleepHrvData: coros.DashboardHRVData{
					AvgSleepHRV:  61,
					SleepHrvBase: 54,
					SleepHrvSD:   8.3,
					SleepHrvList: []coros.HRVDailyValue{{HappenDay: 20260325, Value: 65}},
				},
				RecordDetailList: []coros.RecordDetail{{Type: 7, Record: 284}, {Type: 5, Record: 1599}},
			},
		},
	}

	dashboard, err := GetTrainingDashboard(service)
	if err != nil {
		t.Fatalf("GetTrainingDashboard returned error: %v", err)
	}

	if dashboard.HRV.RecentDays[0].Date != "2026-03-25" {
		t.Fatalf("expected formatted HRV date, got %#v", dashboard.HRV.RecentDays)
	}
	if dashboard.Records["1km"].Time != "00:04:44" {
		t.Fatalf("expected formatted 1km record, got %#v", dashboard.Records)
	}
	if dashboard.Threshold == nil || dashboard.Threshold.LTSP == "" {
		t.Fatalf("expected threshold data, got %#v", dashboard.Threshold)
	}
}

func TestGetTrainingLoadStatus(t *testing.T) {
	service := &stubCorosService{
		dashboardDetailData: &coros.DashboardDetailQueryData{
			SummaryInfo: coros.DashboardDetailSummaryInfo{
				ATI:                    67,
				CTI:                    64,
				TrainingLoadRatio:      1.04,
				TrainingLoadRatioState: 4,
				TiredRate:              44,
				TiredRateNew:           3,
				TiredRateNewState:      3,
				RecomendTlInDays:       0,
			},
		},
	}

	status, err := GetTrainingLoadStatus(service)
	if err != nil {
		t.Fatalf("GetTrainingLoadStatus returned error: %v", err)
	}

	if status.LoadRatioPercent == nil || *status.LoadRatioPercent != 104 {
		t.Fatalf("expected load ratio percent 104, got %#v", status.LoadRatioPercent)
	}
	if status.RecommendedTrainingLoadNextDays == nil || *status.RecommendedTrainingLoadNextDays != 0 {
		t.Fatalf("expected zero recommended days to be preserved, got %#v", status.RecommendedTrainingLoadNextDays)
	}
}

func TestGetRecentActivities(t *testing.T) {
	service := &stubCorosService{
		dashboardDetailData: &coros.DashboardDetailQueryData{
			SportDataList: []coros.RecentSportData{
				{
					HappenDay:      20260325,
					LabelID:        "476321952934428773",
					SportType:      100,
					Mode:           8,
					SubMode:        1,
					Distance:       5273.57,
					Duration:       1983,
					AvgPace:        376,
					AvgHeartRate:   138,
					AvgPower:       190,
					Step:           5784,
					TrainingLoad:   46,
					TotalElevation: 0,
				},
				{
					HappenDay:    20260323,
					LabelID:      "older",
					Distance:     2363.32,
					Duration:     965,
					AvgPace:      408,
					TrainingLoad: 14,
				},
			},
		},
	}

	activities, err := GetRecentActivities(service, 1)
	if err != nil {
		t.Fatalf("GetRecentActivities returned error: %v", err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}
	if activities[0].Date != "2026-03-25" {
		t.Fatalf("expected formatted date, got %#v", activities[0])
	}
	if activities[0].DistanceKM == nil || *activities[0].DistanceKM != 5.27 {
		t.Fatalf("expected rounded km distance, got %#v", activities[0].DistanceKM)
	}
	if activities[0].AvgPace != "6'16\"/km" {
		t.Fatalf("expected pace text, got %#v", activities[0].AvgPace)
	}
}

func TestGetWeeklySummary(t *testing.T) {
	service := &stubCorosService{
		dashboardDetailData: &coros.DashboardDetailQueryData{
			CurrentWeekRecord: coros.WeekRecord{
				DistanceRecord: coros.AggregateRecord{TotalValue: 18426.2},
				DurationRecord: coros.AggregateRecord{TotalValue: 6396},
				TLRecord:       coros.AggregateRecord{TotalValue: 289},
			},
		},
	}

	summary, err := GetWeeklySummary(service)
	if err != nil {
		t.Fatalf("GetWeeklySummary returned error: %v", err)
	}

	if summary.DistanceKM == nil || *summary.DistanceKM != 18.43 {
		t.Fatalf("expected weekly km 18.43, got %#v", summary.DistanceKM)
	}
	if summary.Duration != "01:46:36" {
		t.Fatalf("expected formatted duration, got %#v", summary.Duration)
	}
}

func TestGetTrainingTrends(t *testing.T) {
	service := &stubCorosService{
		dashboardDetailData: &coros.DashboardDetailQueryData{
			DetailList: []coros.TrainingTrendPoint{
				{HappenDay: 20260323, TrainingLoad: 229, ATI: 95, CTI: 67, VO2Max: 52, StaminaLevel: 74.1, Performance: 4, LTSP: 307},
				{HappenDay: 20260325, TrainingLoad: 60, ATI: 78, CTI: 66, VO2Max: 52, StaminaLevel: 74.1, Performance: 3, LTSP: 307},
				{HappenDay: 20260326, TrainingLoad: 0, ATI: 67, CTI: 64, TiredRate: 44, TiredRateNew: 3},
			},
		},
	}

	trends, err := GetTrainingTrends(service, 2)
	if err != nil {
		t.Fatalf("GetTrainingTrends returned error: %v", err)
	}

	if len(trends) != 2 {
		t.Fatalf("expected 2 trends, got %d", len(trends))
	}
	if trends[0].Date != "2026-03-25" || trends[1].Date != "2026-03-26" {
		t.Fatalf("expected latest 2 trends, got %#v", trends)
	}
	if trends[0].LTSPText != "5'07\"/km" {
		t.Fatalf("expected LTSP pace text, got %#v", trends[0].LTSPText)
	}
	if trends[1].VO2Max != nil {
		t.Fatalf("expected zero VO2Max to be omitted, got %#v", trends[1].VO2Max)
	}
	if trends[1].TrainingLoad == nil || *trends[1].TrainingLoad != 0 {
		t.Fatalf("expected zero training load to be preserved, got %#v", trends[1].TrainingLoad)
	}
	if trends[1].ATI == nil || *trends[1].ATI != 67 {
		t.Fatalf("expected ATI to be preserved, got %#v", trends[1].ATI)
	}
}
