package personalmetrics

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"coros-fit-mcp/internal/service/coros"
)

type RunnerProfile struct {
	WeightKG                     *float64    `json:"weight_kg,omitempty"`
	HeightCM                     *float64    `json:"height_cm,omitempty"`
	Birthday                     string      `json:"birthday,omitempty"`
	Sex                          *int        `json:"sex,omitempty"`
	CountryCode                  string      `json:"country_code,omitempty"`
	MaxHR                        *float64    `json:"max_hr,omitempty"`
	RestingHR                    *float64    `json:"resting_hr,omitempty"`
	LactateThresholdHR           *float64    `json:"lactate_threshold_hr,omitempty"`
	LactateThresholdPaceSecPerKM *float64    `json:"lactate_threshold_pace_sec_per_km,omitempty"`
	LactateThresholdPace         string      `json:"lactate_threshold_pace,omitempty"`
	HRZoneType                   *int        `json:"hr_zone_type,omitempty"`
	HRZonesLTHR                  []HeartZone `json:"hr_zones_lthr,omitempty"`
	PaceZonesLTSP                []PaceZone  `json:"pace_zones_ltsp,omitempty"`
}

type TrainingZones struct {
	HRZoneType    *int                `json:"hr_zone_type,omitempty"`
	HRZonesLTHR   []HeartZone         `json:"hr_zones_lthr,omitempty"`
	PaceZonesLTSP []PaceZone          `json:"pace_zones_ltsp,omitempty"`
	Guidance      TrainingRunGuidance `json:"guidance"`
}

type TrainingRunGuidance struct {
	Recovery  RunZoneGuidance `json:"recovery_run"`
	Aerobic   RunZoneGuidance `json:"aerobic_run"`
	Threshold RunZoneGuidance `json:"threshold_run"`
	Interval  RunZoneGuidance `json:"interval_run"`
}

type RunZoneGuidance struct {
	Description string `json:"description,omitempty"`
	HeartRate   string `json:"heart_rate,omitempty"`
	Pace        string `json:"pace,omitempty"`
}

type HeartZone struct {
	Index int      `json:"index"`
	HR    *float64 `json:"hr,omitempty"`
	Ratio *float64 `json:"ratio,omitempty"`
}

type PaceZone struct {
	Index    int      `json:"index"`
	Pace     *float64 `json:"pace_sec_per_km,omitempty"`
	PaceText string   `json:"pace,omitempty"`
	Ratio    *float64 `json:"ratio,omitempty"`
}

type TrainingDashboard struct {
	RunningAbility RunningAbility      `json:"running_ability"`
	Threshold      *DashboardThreshold `json:"threshold,omitempty"`
	Recovery       RecoveryStatus      `json:"recovery"`
	HRV            HRVStatus           `json:"hrv"`
	Records        map[string]Record   `json:"records,omitempty"`
}

type DashboardThreshold struct {
	LTHR         *float64    `json:"lthr,omitempty"`
	LTSPSecPerKM *float64    `json:"ltsp_sec_per_km,omitempty"`
	LTSP         string      `json:"ltsp,omitempty"`
	MaxHR        *float64    `json:"max_hr,omitempty"`
	RestingHR    *float64    `json:"resting_hr,omitempty"`
	LTHRZones    []HeartZone `json:"lthr_zones,omitempty"`
	LTSPZones    []PaceZone  `json:"ltsp_zones,omitempty"`
}

type RunningAbility struct {
	StaminaLevel                  *float64 `json:"stamina_level,omitempty"`
	AerobicEnduranceScore         *float64 `json:"aerobic_endurance_score,omitempty"`
	LactateThresholdCapacityScore *float64 `json:"lactate_threshold_capacity_score,omitempty"`
	AnaerobicEnduranceScore       *float64 `json:"anaerobic_endurance_score,omitempty"`
	AnaerobicCapacityScore        *float64 `json:"anaerobic_capacity_score,omitempty"`
	RankingPercent                *float64 `json:"ranking_percent,omitempty"`
}

type RecoveryStatus struct {
	RecoveryPct       *float64 `json:"recovery_pct,omitempty"`
	RecoveryState     *int     `json:"recovery_state,omitempty"`
	FullRecoveryHours *float64 `json:"full_recovery_hours,omitempty"`
}

type HRVStatus struct {
	AvgSleepHRV *float64     `json:"avg_sleep_hrv,omitempty"`
	HRVBase     *float64     `json:"hrv_base,omitempty"`
	HRVSD       *float64     `json:"hrv_sd,omitempty"`
	RecentDays  []HRVDayStat `json:"recent_days,omitempty"`
}

type HRVDayStat struct {
	Date  string   `json:"date"`
	Value *float64 `json:"value,omitempty"`
}

type Record struct {
	Seconds *float64 `json:"seconds,omitempty"`
	Time    string   `json:"time,omitempty"`
}

type TrainingLoadStatus struct {
	AcuteTrainingLoad               *float64 `json:"acute_training_load,omitempty"`
	ChronicTrainingLoad             *float64 `json:"chronic_training_load,omitempty"`
	LoadRatio                       *float64 `json:"load_ratio,omitempty"`
	LoadRatioPercent                *float64 `json:"load_ratio_percent,omitempty"`
	LoadRatioState                  *int     `json:"load_ratio_state,omitempty"`
	Fatigue                         *float64 `json:"fatigue,omitempty"`
	FatigueNew                      *float64 `json:"fatigue_new,omitempty"`
	FatigueState                    *int     `json:"fatigue_state,omitempty"`
	RecommendedTrainingLoadNextDays *float64 `json:"recommended_training_load_next_days,omitempty"`
}

type RecentActivity struct {
	Date            string   `json:"date,omitempty"`
	LabelID         string   `json:"label_id,omitempty"`
	SportType       *int     `json:"sport_type,omitempty"`
	Mode            *int     `json:"mode,omitempty"`
	SubMode         *int     `json:"sub_mode,omitempty"`
	DistanceM       *float64 `json:"distance_m,omitempty"`
	DistanceKM      *float64 `json:"distance_km,omitempty"`
	DurationS       *float64 `json:"duration_s,omitempty"`
	Duration        string   `json:"duration,omitempty"`
	AvgPaceSecPerKM *float64 `json:"avg_pace_sec_per_km,omitempty"`
	AvgPace         string   `json:"avg_pace,omitempty"`
	AvgHR           *float64 `json:"avg_hr,omitempty"`
	AvgPower        *float64 `json:"avg_power,omitempty"`
	Steps           *float64 `json:"steps,omitempty"`
	TrainingLoad    *float64 `json:"training_load,omitempty"`
	ElevationM      *float64 `json:"elevation_m,omitempty"`
}

type WeeklySummary struct {
	DistanceM    *float64 `json:"distance_m,omitempty"`
	DistanceKM   *float64 `json:"distance_km,omitempty"`
	DurationS    *float64 `json:"duration_s,omitempty"`
	Duration     string   `json:"duration,omitempty"`
	TrainingLoad *float64 `json:"training_load,omitempty"`
}

type TrainingTrend struct {
	Date         string   `json:"date,omitempty"`
	TrainingLoad *float64 `json:"training_load,omitempty"`
	ATI          *float64 `json:"ati,omitempty"`
	CTI          *float64 `json:"cti,omitempty"`
	VO2Max       *float64 `json:"vo2max,omitempty"`
	StaminaLevel *float64 `json:"stamina_level,omitempty"`
	Performance  *float64 `json:"performance,omitempty"`
	LTHR         *float64 `json:"lthr,omitempty"`
	LTSP         *float64 `json:"ltsp_sec_per_km,omitempty"`
	LTSPText     string   `json:"ltsp,omitempty"`
	Fatigue      *float64 `json:"fatigue,omitempty"`
	FatigueNew   *float64 `json:"fatigue_new,omitempty"`
}

type Response struct {
	Markdown string      `json:"markdown"`
	Data     interface{} `json:"data"`
}

func GetRunnerProfile(service coros.CorosService) (*RunnerProfile, error) {
	data, err := service.AccountQuery()
	if err != nil {
		return nil, err
	}

	sex := intPtr(data.Sex)
	hrZoneType := intPtr(data.HrZoneType)

	profile := &RunnerProfile{
		WeightKG:                     finite(data.Weight),
		HeightCM:                     finite(data.Stature),
		Birthday:                     formatYMDInt(data.Birthday),
		Sex:                          sex,
		CountryCode:                  data.CountryCode,
		MaxHR:                        finite(data.MaxHr),
		RestingHR:                    finite(data.Rhr),
		LactateThresholdHR:           finite(data.ZoneData.LTHR),
		LactateThresholdPaceSecPerKM: finite(data.ZoneData.LTSP),
		LactateThresholdPace:         paceText(finite(data.ZoneData.LTSP)),
		HRZoneType:                   hrZoneType,
		HRZonesLTHR:                  buildHeartZones(data.ZoneData.LTHRZone),
		PaceZonesLTSP:                buildPaceZones(data.ZoneData.LTSPZone),
	}

	return profile, nil
}

func GetTrainingZones(service coros.CorosService) (*TrainingZones, error) {
	profile, err := GetRunnerProfile(service)
	if err != nil {
		return nil, err
	}

	return &TrainingZones{
		HRZoneType:    profile.HRZoneType,
		HRZonesLTHR:   profile.HRZonesLTHR,
		PaceZonesLTSP: profile.PaceZonesLTSP,
		Guidance:      buildTrainingRunGuidance(profile.HRZonesLTHR, profile.PaceZonesLTSP),
	}, nil
}

func GetTrainingDashboard(service coros.CorosService) (*TrainingDashboard, error) {
	data, err := service.DashboardQuery()
	if err != nil {
		return nil, err
	}

	info := data.SummaryInfo
	recoveryState := intPtr(info.RecoveryState)

	dashboard := &TrainingDashboard{
		RunningAbility: RunningAbility{
			StaminaLevel:                  finite(info.StaminaLevel),
			AerobicEnduranceScore:         finite(info.AerobicEnduranceScore),
			LactateThresholdCapacityScore: finite(info.LactateThresholdCapacityScore),
			AnaerobicEnduranceScore:       finite(info.AnaerobicEnduranceScore),
			AnaerobicCapacityScore:        finite(info.AnaerobicCapacityScore),
			RankingPercent:                finite(info.StaminaLevelRanking),
		},
		Threshold: &DashboardThreshold{
			LTHR:         finite(info.LTHR),
			LTSPSecPerKM: finite(info.LTSP),
			LTSP:         paceText(finite(info.LTSP)),
			MaxHR:        finite(info.FitnessMaxHr),
			RestingHR:    finite(info.Rhr),
			LTHRZones:    buildHeartZones(info.LTHRZone),
			LTSPZones:    buildPaceZones(info.LTSPZone),
		},
		Recovery: RecoveryStatus{
			RecoveryPct:       finite(info.RecoveryPct),
			RecoveryState:     recoveryState,
			FullRecoveryHours: finite(info.FullRecoveryHours),
		},
		HRV: HRVStatus{
			AvgSleepHRV: finite(info.SleepHrvData.AvgSleepHRV),
			HRVBase:     finite(info.SleepHrvData.SleepHrvBase),
			HRVSD:       finite(info.SleepHrvData.SleepHrvSD),
			RecentDays:  buildHRVDays(info.SleepHrvData.SleepHrvList),
		},
		Records: buildRecords(info.RecordDetailList),
	}

	return dashboard, nil
}

func GetTrainingLoadStatus(service coros.CorosService) (*TrainingLoadStatus, error) {
	data, err := service.DashboardDetailQuery()
	if err != nil {
		return nil, err
	}

	info := data.SummaryInfo
	loadRatioState := intPtr(info.TrainingLoadRatioState)
	fatigueState := intPtr(info.TiredRateNewState)

	status := &TrainingLoadStatus{
		AcuteTrainingLoad:               finiteOrZero(info.ATI),
		ChronicTrainingLoad:             finiteOrZero(info.CTI),
		LoadRatio:                       finiteOrZero(info.TrainingLoadRatio),
		LoadRatioPercent:                ratioPercent(finite(info.TrainingLoadRatio)),
		LoadRatioState:                  loadRatioState,
		Fatigue:                         finiteOrZero(info.TiredRate),
		FatigueNew:                      finiteOrZero(info.TiredRateNew),
		FatigueState:                    fatigueState,
		RecommendedTrainingLoadNextDays: finiteOrZero(info.RecomendTlInDays),
	}

	return status, nil
}

func GetRecentActivities(service coros.CorosService, limit int) ([]RecentActivity, error) {
	data, err := service.DashboardDetailQuery()
	if err != nil {
		return nil, err
	}

	activities := make([]RecentActivity, 0, len(data.SportDataList))
	for _, item := range data.SportDataList {
		activities = append(activities, RecentActivity{
			Date:            formatYMDInt(item.HappenDay),
			LabelID:         item.LabelID,
			SportType:       intPtr(item.SportType),
			Mode:            intPtr(item.Mode),
			SubMode:         intPtr(item.SubMode),
			DistanceM:       finiteOrZero(item.Distance),
			DistanceKM:      kmValue(finiteOrZero(item.Distance)),
			DurationS:       finiteOrZero(item.Duration),
			Duration:        durationText(finiteOrZero(item.Duration)),
			AvgPaceSecPerKM: finite(item.AvgPace),
			AvgPace:         paceText(finite(item.AvgPace)),
			AvgHR:           finiteOrZero(item.AvgHeartRate),
			AvgPower:        finiteOrZero(item.AvgPower),
			Steps:           finiteOrZero(item.Step),
			TrainingLoad:    finiteOrZero(item.TrainingLoad),
			ElevationM:      finiteOrZero(item.TotalElevation),
		})
	}

	if limit > 0 && len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

func GetWeeklySummary(service coros.CorosService) (*WeeklySummary, error) {
	data, err := service.DashboardDetailQuery()
	if err != nil {
		return nil, err
	}

	week := data.CurrentWeekRecord
	summary := &WeeklySummary{
		DistanceM:    finiteOrZero(week.DistanceRecord.TotalValue),
		DistanceKM:   kmValue(finiteOrZero(week.DistanceRecord.TotalValue)),
		DurationS:    finiteOrZero(week.DurationRecord.TotalValue),
		Duration:     durationText(finiteOrZero(week.DurationRecord.TotalValue)),
		TrainingLoad: finiteOrZero(week.TLRecord.TotalValue),
	}

	return summary, nil
}

func GetTrainingTrends(service coros.CorosService, days int) ([]TrainingTrend, error) {
	data, err := service.DashboardDetailQuery()
	if err != nil {
		return nil, err
	}

	trends := make([]TrainingTrend, 0, len(data.DetailList))
	for _, item := range data.DetailList {
		trends = append(trends, TrainingTrend{
			Date:         formatYMDInt(item.HappenDay),
			TrainingLoad: finiteOrZero(item.TrainingLoad),
			ATI:          finiteOrZero(item.ATI),
			CTI:          finiteOrZero(item.CTI),
			VO2Max:       finite(item.VO2Max),
			StaminaLevel: finite(item.StaminaLevel),
			Performance:  finite(item.Performance),
			LTHR:         finite(item.LTHR),
			LTSP:         finite(item.LTSP),
			LTSPText:     paceText(finite(item.LTSP)),
			Fatigue:      finiteOrZero(item.TiredRate),
			FatigueNew:   finiteOrZero(item.TiredRateNew),
		})
	}

	sort.Slice(trends, func(i, j int) bool { return trends[i].Date < trends[j].Date })
	if days > 0 && len(trends) > days {
		trends = trends[len(trends)-days:]
	}

	return trends, nil
}

func ResponseText(title string, data interface{}) (string, error) {
	response := Response{
		Markdown: markdownFor(title, data),
		Data:     data,
	}

	body, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func markdownFor(title string, data interface{}) string {
	return fmt.Sprintf("# %s\n\n返回结构化训练指标数据。", title)
}

func buildHeartZones(zones []coros.ZoneBoundary) []HeartZone {
	result := make([]HeartZone, 0, len(zones))
	for i, zone := range zones {
		result = append(result, HeartZone{
			Index: i,
			HR:    finite(zone.Hr),
			Ratio: finite(zone.Ratio),
		})
	}
	return result
}

func buildPaceZones(zones []coros.PaceZoneBoundary) []PaceZone {
	result := make([]PaceZone, 0, len(zones))
	for i, zone := range zones {
		pace := finite(zone.Pace)
		result = append(result, PaceZone{
			Index:    i,
			Pace:     pace,
			PaceText: paceText(pace),
			Ratio:    finite(zone.Ratio),
		})
	}
	return result
}

func buildHRVDays(values []coros.HRVDailyValue) []HRVDayStat {
	result := make([]HRVDayStat, 0, len(values))
	for _, item := range values {
		result = append(result, HRVDayStat{
			Date:  formatYMDInt(item.HappenDay),
			Value: finite(item.Value),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Date < result[j].Date })
	return result
}

func buildRecords(items []coros.RecordDetail) map[string]Record {
	result := make(map[string]Record)
	for _, item := range items {
		name, ok := recordTypeName(item.Type)
		if !ok {
			continue
		}
		seconds := finite(item.Record)
		if existing, ok := result[name]; ok {
			if compareRecordSeconds(seconds, existing.Seconds) >= 0 {
				continue
			}
		}
		result[name] = Record{
			Seconds: seconds,
			Time:    durationText(seconds),
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func buildTrainingRunGuidance(hrZones []HeartZone, paceZones []PaceZone) TrainingRunGuidance {
	return TrainingRunGuidance{
		Recovery: RunZoneGuidance{
			Description: "用于恢复日、放松跑，强度应明显低于阈值。",
			HeartRate:   hrRangeText(hrZones, 0, 1),
			Pace:        paceRangeText(paceZones, 0, 1),
		},
		Aerobic: RunZoneGuidance{
			Description: "用于日常有氧跑和稳态耐力跑，强度低到中等。",
			HeartRate:   hrRangeText(hrZones, 1, 2),
			Pace:        paceRangeText(paceZones, 1, 3),
		},
		Threshold: RunZoneGuidance{
			Description: "用于阈值跑或节奏跑，接近乳酸阈值附近。",
			HeartRate:   hrRangeText(hrZones, 2, 3),
			Pace:        paceRangeText(paceZones, 3, 4),
		},
		Interval: RunZoneGuidance{
			Description: "用于间歇跑或高强度重复跑，强度高于阈值。",
			HeartRate:   hrRangeText(hrZones, 3, 4),
			Pace:        paceRangeText(paceZones, 4, 5),
		},
	}
}

func hrRangeText(zones []HeartZone, startIdx, endIdx int) string {
	if startIdx < 0 || endIdx < 0 || startIdx >= len(zones) {
		return ""
	}
	if endIdx >= len(zones) {
		endIdx = len(zones) - 1
	}
	start := zones[startIdx].HR
	end := zones[endIdx].HR
	if start == nil && end == nil {
		return ""
	}
	if start != nil && end != nil {
		return fmt.Sprintf("约 %.0f-%.0f bpm", *start, *end)
	}
	if start != nil {
		return fmt.Sprintf("约 >= %.0f bpm", *start)
	}
	return fmt.Sprintf("约 <= %.0f bpm", *end)
}

func paceRangeText(zones []PaceZone, slowerIdx, fasterIdx int) string {
	if slowerIdx < 0 || fasterIdx < 0 || slowerIdx >= len(zones) {
		return ""
	}
	if fasterIdx >= len(zones) {
		fasterIdx = len(zones) - 1
	}
	slower := zones[slowerIdx].PaceText
	faster := zones[fasterIdx].PaceText
	if slower == "" && faster == "" {
		return ""
	}
	if slower != "" && faster != "" {
		return fmt.Sprintf("约 %s - %s", slower, faster)
	}
	if slower != "" {
		return fmt.Sprintf("约慢于 %s", slower)
	}
	return fmt.Sprintf("约快于 %s", faster)
}

func compareRecordSeconds(a, b *float64) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1
	}
	if b == nil {
		return -1
	}
	switch {
	case *a < *b:
		return -1
	case *a > *b:
		return 1
	default:
		return 0
	}
}

func recordTypeName(t int) (string, bool) {
	switch t {
	case 4:
		return "10km", true
	case 5:
		return "5km", true
	case 6:
		return "3km", true
	case 7:
		return "1km", true
	case 8:
		return "1mile", true
	default:
		return "", false
	}
}

func finite(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v == 0 {
		return nil
	}
	value := v
	return &value
}

func finiteOrZero(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	value := v
	return &value
}

func intPtr(v int) *int {
	value := v
	return &value
}

func formatYMDInt(v int) string {
	if v <= 0 {
		return ""
	}
	s := fmt.Sprintf("%08d", v)
	return fmt.Sprintf("%s-%s-%s", s[:4], s[4:6], s[6:8])
}

func paceText(seconds *float64) string {
	if seconds == nil || *seconds <= 0 {
		return ""
	}
	total := int(math.Round(*seconds))
	return fmt.Sprintf("%d'%02d\"/km", total/60, total%60)
}

func durationText(seconds *float64) string {
	if seconds == nil || *seconds <= 0 {
		return ""
	}
	total := int(math.Round(*seconds))
	hours := total / 3600
	minutes := (total % 3600) / 60
	secs := total % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

func ratioPercent(v *float64) *float64 {
	if v == nil {
		return nil
	}
	value := math.Round(*v*10000) / 100
	return &value
}

func kmValue(meters *float64) *float64 {
	if meters == nil {
		return nil
	}
	value := math.Round((*meters/1000)*100) / 100
	return &value
}

func DebugString(v interface{}) string {
	body, _ := json.Marshal(v)
	return strings.TrimSpace(string(body))
}
