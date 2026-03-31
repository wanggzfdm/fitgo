package traininganalysis

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"coros-fit-mcp/internal/service/activitysummary"
	"coros-fit-mcp/internal/service/coros"
	"coros-fit-mcp/internal/service/personalmetrics"
)

type AnalysisResponse struct {
	Markdown string       `json:"markdown"`
	Data     AnalysisData `json:"data"`
}

type AnalysisData struct {
	FocusDate          string             `json:"focus_date,omitempty"`
	Status             TrainingStatus     `json:"status"`
	ActivityAssessment ActivityAssessment `json:"activity_assessment"`
	Recommendation     Recommendation     `json:"recommendation"`
	NextPlan           NextPlan           `json:"next_plan"`
	SupportingData     SupportingData     `json:"supporting_data"`
	Errors             map[string]string  `json:"errors,omitempty"`
}

type TrainingStatus struct {
	Recovery    string `json:"recovery,omitempty"`
	Fatigue     string `json:"fatigue,omitempty"`
	LoadBalance string `json:"load_balance,omitempty"`
	Readiness   string `json:"readiness,omitempty"`
}

type ActivityAssessment struct {
	FocusType string   `json:"focus_activity_type,omitempty"`
	Reason    []string `json:"reason,omitempty"`
}

type Recommendation struct {
	Tomorrow       string         `json:"tomorrow,omitempty"`
	Reason         []string       `json:"reason,omitempty"`
	SuggestedRange SuggestedRange `json:"suggested_range,omitempty"`
}

type SuggestedRange struct {
	HeartRate string `json:"heart_rate,omitempty"`
	Pace      string `json:"pace,omitempty"`
}

type NextPlan struct {
	Goal string    `json:"goal,omitempty"`
	Days []PlanDay `json:"days,omitempty"`
}

type PlanDay struct {
	DayOffset int            `json:"day_offset"`
	Type      string         `json:"type,omitempty"`
	Purpose   string         `json:"purpose,omitempty"`
	Intensity SuggestedRange `json:"intensity,omitempty"`
	Notes     string         `json:"notes,omitempty"`
}

type SupportingData struct {
	TrainingProfile *personalmetrics.TrainingProfileBundle `json:"training_profile,omitempty"`
	TrainingContext *personalmetrics.TrainingContextBundle `json:"training_context,omitempty"`
	LatestActivity  *activitysummary.SummaryResponse       `json:"latest_activity,omitempty"`
	FocusDaySummary *activitysummary.DailySummaryResponse  `json:"focus_day_summary,omitempty"`
}

func AnalyzeTrainingStatus(service coros.CorosService, date string) (*AnalysisResponse, error) {
	profile, profileErr := personalmetrics.GetTrainingProfileBundle(service)
	context, contextErr := personalmetrics.GetTrainingContextBundle(service, 5, 7)

	var latest *activitysummary.SummaryResponse
	var latestErr error
	if date == "" {
		summary, err := activitysummary.SummarizeLatestCorosActivity(service)
		if err != nil {
			latestErr = err
		} else {
			resp := activitysummary.ToResponse(*summary)
			latest = &resp
		}
	}

	var daily *activitysummary.DailySummaryResponse
	var dailyErr error
	if date != "" {
		daily, dailyErr = activitysummary.SummarizeCorosDailyActivities(service, date)
	}

	focusSummary, focusDate := chooseFocusSummary(date, latest, daily)
	if focusDate == "" && daily != nil {
		focusDate = daily.Date
	}

	if profile == nil && context == nil && latest == nil && daily == nil {
		return nil, fmt.Errorf("无法获取训练分析所需数据")
	}

	status := buildStatus(context)
	assessment := buildActivityAssessment(profile, focusSummary)
	recommendation := buildRecommendation(profile, context, assessment)
	plan := buildNextPlan(profile, status, recommendation)

	errs := map[string]string{}
	if profileErr != nil {
		errs["training_profile"] = profileErr.Error()
	}
	if contextErr != nil {
		errs["training_context"] = contextErr.Error()
	}
	if latestErr != nil {
		errs["latest_activity"] = latestErr.Error()
	}
	if dailyErr != nil {
		errs["focus_day_summary"] = dailyErr.Error()
	}
	if len(errs) == 0 {
		errs = nil
	}

	data := AnalysisData{
		FocusDate:          focusDate,
		Status:             status,
		ActivityAssessment: assessment,
		Recommendation:     recommendation,
		NextPlan:           plan,
		SupportingData: SupportingData{
			TrainingProfile: profile,
			TrainingContext: context,
			LatestActivity:  latest,
			FocusDaySummary: daily,
		},
		Errors: errs,
	}

	return &AnalysisResponse{
		Markdown: buildMarkdown(data),
		Data:     data,
	}, nil
}

func ResponseText(resp AnalysisResponse) (string, error) {
	body, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func chooseFocusSummary(date string, latest *activitysummary.SummaryResponse, daily *activitysummary.DailySummaryResponse) (*activitysummary.ActivitySummary, string) {
	if date != "" && daily != nil {
		return &daily.DailySummary.Summary, daily.Date
	}
	if latest != nil {
		return &latest.Summary, trimDate(latest.Summary.StartTime)
	}
	return nil, date
}

func buildStatus(context *personalmetrics.TrainingContextBundle) TrainingStatus {
	status := TrainingStatus{
		Recovery:    "unknown",
		Fatigue:     "unknown",
		LoadBalance: "unknown",
		Readiness:   "unknown",
	}
	if context == nil || context.TrainingLoadStatus == nil {
		return status
	}
	load := context.TrainingLoadStatus

	if load.FatigueNew != nil {
		switch {
		case *load.FatigueNew <= 5:
			status.Fatigue = "low"
		case *load.FatigueNew <= 15:
			status.Fatigue = "moderate"
		default:
			status.Fatigue = "high"
		}
	}

	if load.LoadRatio != nil {
		switch {
		case *load.LoadRatio < 0.8:
			status.LoadBalance = "underloaded"
		case *load.LoadRatio <= 1.15:
			status.LoadBalance = "balanced"
		default:
			status.LoadBalance = "overloaded"
		}
	}

	if load.FatigueState != nil {
		switch {
		case *load.FatigueState <= 2:
			status.Recovery = "good"
		case *load.FatigueState == 3:
			status.Recovery = "moderate"
		default:
			status.Recovery = "limited"
		}
	}

	switch {
	case status.Fatigue == "high" || status.LoadBalance == "overloaded":
		status.Readiness = "easy_only"
	case status.Recovery == "good" && status.Fatigue == "low" && status.LoadBalance == "balanced":
		status.Readiness = "moderate_to_good"
	default:
		status.Readiness = "moderate"
	}

	return status
}

func buildActivityAssessment(profile *personalmetrics.TrainingProfileBundle, summary *activitysummary.ActivitySummary) ActivityAssessment {
	assessment := ActivityAssessment{FocusType: "unknown"}
	if summary == nil {
		return assessment
	}

	zones := (*personalmetrics.TrainingZones)(nil)
	if profile != nil {
		zones = profile.TrainingZones
	}

	pace := value(summary.AveragePaceSecPerKM)
	hr := value(summary.AverageHeartRate)
	tl := value(summary.TrainingLoad)

	assessment.Reason = []string{}
	if zones != nil {
		recovery := zones.Guidance.Recovery
		aerobic := zones.Guidance.Aerobic
		threshold := zones.Guidance.Threshold
		interval := zones.Guidance.Interval

		switch {
		case tl > 0 && tl <= 35 && matchesRecovery(zones, pace, hr):
			assessment.FocusType = "recovery_run"
			assessment.Reason = append(assessment.Reason, "平均心率和配速都落在恢复跑区间附近")
		case matchesRecovery(zones, pace, hr):
			assessment.FocusType = "recovery_run"
			assessment.Reason = append(assessment.Reason, "平均心率和配速更接近恢复跑区间")
		case matchesAerobic(zones, pace, hr):
			assessment.FocusType = "aerobic_run"
			assessment.Reason = append(assessment.Reason, "平均心率和配速主要落在有氧跑区间")
		case matchesThreshold(zones, pace, hr):
			assessment.FocusType = "threshold_run"
			assessment.Reason = append(assessment.Reason, "平均心率或配速接近阈值训练区间")
		case matchesInterval(zones, pace, hr):
			assessment.FocusType = "interval_run"
			assessment.Reason = append(assessment.Reason, "配速明显快于阈值，更像高强度刺激")
		default:
			assessment.FocusType = "mixed_or_easy_run"
			assessment.Reason = append(assessment.Reason, "训练强度介于恢复和有氧之间")
		}

		if recovery.Pace != "" || aerobic.Pace != "" || threshold.Pace != "" || interval.Pace != "" {
			assessment.Reason = append(assessment.Reason, "分类依据使用了当前训练分区和阈值配速")
		}
	}

	if tl > 0 {
		assessment.Reason = append(assessment.Reason, fmt.Sprintf("本次训练负荷约 %.0f", tl))
	}

	return assessment
}

func buildRecommendation(profile *personalmetrics.TrainingProfileBundle, context *personalmetrics.TrainingContextBundle, assessment ActivityAssessment) Recommendation {
	rec := Recommendation{
		Tomorrow: "aerobic_run",
		Reason:   []string{},
	}

	if profile != nil && profile.TrainingZones != nil {
		rec.SuggestedRange = suggestedRangeFor(profile.TrainingZones, "aerobic_run")
	}

	status := buildStatus(context)
	switch {
	case status.Readiness == "easy_only":
		rec.Tomorrow = "recovery_run"
		if profile != nil && profile.TrainingZones != nil {
			rec.SuggestedRange = suggestedRangeFor(profile.TrainingZones, "recovery_run")
		}
		rec.Reason = append(rec.Reason, "当前疲劳或负荷偏高，优先恢复")
	case assessment.FocusType == "recovery_run" && status.Readiness == "moderate_to_good":
		rec.Tomorrow = "aerobic_run"
		rec.Reason = append(rec.Reason, "最近一次训练偏轻松，且恢复状态良好")
	case assessment.FocusType == "threshold_run" || assessment.FocusType == "interval_run":
		rec.Tomorrow = "recovery_run"
		if profile != nil && profile.TrainingZones != nil {
			rec.SuggestedRange = suggestedRangeFor(profile.TrainingZones, "recovery_run")
		}
		rec.Reason = append(rec.Reason, "最近已有质量刺激，下一天更适合恢复")
	default:
		rec.Reason = append(rec.Reason, "近期负荷处于可继续训练的区间")
	}

	if context != nil && context.TrainingLoadStatus != nil && context.TrainingLoadStatus.LoadRatioPercent != nil {
		rec.Reason = append(rec.Reason, fmt.Sprintf("当前负荷比约 %.0f%%", *context.TrainingLoadStatus.LoadRatioPercent))
	}

	return rec
}

func buildNextPlan(profile *personalmetrics.TrainingProfileBundle, status TrainingStatus, rec Recommendation) NextPlan {
	plan := NextPlan{
		Goal: "近期以恢复吸收和稳态推进为主",
		Days: make([]PlanDay, 0, 3),
	}

	plan.Days = append(plan.Days, PlanDay{
		DayOffset: 1,
		Type:      rec.Tomorrow,
		Purpose:   purposeFor(rec.Tomorrow),
		Intensity: rec.SuggestedRange,
	})

	day2Type := "recovery_run_or_rest"
	day2Notes := "观察腿部反馈和主观疲劳，优先吸收前一天训练刺激。"
	if status.Readiness == "easy_only" {
		day2Type = "recovery_run"
	}
	plan.Days = append(plan.Days, PlanDay{
		DayOffset: 2,
		Type:      day2Type,
		Purpose:   "恢复与吸收",
		Notes:     day2Notes,
		Intensity: suggestedRangeForBundle(profile, "recovery_run"),
	})

	day3Type := "threshold_run"
	day3Purpose := "若恢复顺利，安排一堂轻量质量课。"
	if status.Readiness == "easy_only" || status.Fatigue == "high" {
		day3Type = "aerobic_run"
		day3Purpose = "继续稳态推进，避免过早叠加强度。"
	}
	plan.Days = append(plan.Days, PlanDay{
		DayOffset: 3,
		Type:      day3Type,
		Purpose:   day3Purpose,
		Intensity: suggestedRangeForBundle(profile, day3Type),
	})

	return plan
}

func buildMarkdown(data AnalysisData) string {
	lines := []string{"# 训练状态分析"}
	if data.FocusDate != "" {
		lines = append(lines, "", fmt.Sprintf("- 分析日期：%s", data.FocusDate))
	}

	lines = append(lines, "", "## 当前状态")
	lines = append(lines, fmt.Sprintf("- 恢复：%s", data.Status.Recovery))
	lines = append(lines, fmt.Sprintf("- 疲劳：%s", data.Status.Fatigue))
	lines = append(lines, fmt.Sprintf("- 负荷平衡：%s", data.Status.LoadBalance))
	lines = append(lines, fmt.Sprintf("- 训练准备度：%s", data.Status.Readiness))

	lines = append(lines, "", "## 焦点训练判断")
	lines = append(lines, fmt.Sprintf("- 当前焦点训练类型：%s", data.ActivityAssessment.FocusType))
	for _, reason := range data.ActivityAssessment.Reason {
		lines = append(lines, fmt.Sprintf("- %s", reason))
	}

	lines = append(lines, "", "## 明日建议")
	lines = append(lines, fmt.Sprintf("- 推荐：%s", data.Recommendation.Tomorrow))
	if data.Recommendation.SuggestedRange.HeartRate != "" {
		lines = append(lines, fmt.Sprintf("- 心率：%s", data.Recommendation.SuggestedRange.HeartRate))
	}
	if data.Recommendation.SuggestedRange.Pace != "" {
		lines = append(lines, fmt.Sprintf("- 配速：%s", data.Recommendation.SuggestedRange.Pace))
	}
	for _, reason := range data.Recommendation.Reason {
		lines = append(lines, fmt.Sprintf("- %s", reason))
	}

	lines = append(lines, "", "## 后续 3 天计划")
	lines = append(lines, fmt.Sprintf("- 目标：%s", data.NextPlan.Goal))
	for _, day := range data.NextPlan.Days {
		line := fmt.Sprintf("- D+%d：%s，%s", day.DayOffset, day.Type, day.Purpose)
		if day.Intensity.HeartRate != "" || day.Intensity.Pace != "" {
			parts := make([]string, 0, 2)
			if day.Intensity.HeartRate != "" {
				parts = append(parts, day.Intensity.HeartRate)
			}
			if day.Intensity.Pace != "" {
				parts = append(parts, day.Intensity.Pace)
			}
			line += fmt.Sprintf("（%s）", strings.Join(parts, "；"))
		}
		lines = append(lines, line)
		if day.Notes != "" {
			lines = append(lines, fmt.Sprintf("- 备注：%s", day.Notes))
		}
	}

	if len(data.Errors) > 0 {
		keys := make([]string, 0, len(data.Errors))
		for k := range data.Errors {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		lines = append(lines, "", "## 数据异常")
		for _, key := range keys {
			lines = append(lines, fmt.Sprintf("- %s: %s", key, data.Errors[key]))
		}
	}

	return strings.Join(lines, "\n")
}

func suggestedRangeForBundle(profile *personalmetrics.TrainingProfileBundle, kind string) SuggestedRange {
	if profile == nil || profile.TrainingZones == nil {
		return SuggestedRange{}
	}
	return suggestedRangeFor(profile.TrainingZones, kind)
}

func suggestedRangeFor(zones *personalmetrics.TrainingZones, kind string) SuggestedRange {
	if zones == nil {
		return SuggestedRange{}
	}
	switch kind {
	case "recovery_run", "recovery_run_or_rest":
		return SuggestedRange{HeartRate: zones.Guidance.Recovery.HeartRate, Pace: zones.Guidance.Recovery.Pace}
	case "aerobic_run":
		return SuggestedRange{HeartRate: zones.Guidance.Aerobic.HeartRate, Pace: zones.Guidance.Aerobic.Pace}
	case "threshold_run":
		return SuggestedRange{HeartRate: zones.Guidance.Threshold.HeartRate, Pace: zones.Guidance.Threshold.Pace}
	case "interval_run":
		return SuggestedRange{HeartRate: zones.Guidance.Interval.HeartRate, Pace: zones.Guidance.Interval.Pace}
	default:
		return SuggestedRange{}
	}
}

func purposeFor(kind string) string {
	switch kind {
	case "recovery_run":
		return "通过轻松跑促进恢复。"
	case "aerobic_run":
		return "保持稳态有氧刺激。"
	case "threshold_run":
		return "提升阈值能力和节奏耐力。"
	case "interval_run":
		return "进行高强度神经和速度刺激。"
	default:
		return "维持训练节奏。"
	}
}

func matchesRecovery(zones *personalmetrics.TrainingZones, pace, hr float64) bool {
	return withinHR(zones.HRZonesLTHR, hr, 0, 1) || withinPace(zones.PaceZonesLTSP, pace, 0, 1)
}

func matchesAerobic(zones *personalmetrics.TrainingZones, pace, hr float64) bool {
	return withinHR(zones.HRZonesLTHR, hr, 1, 2) || withinPace(zones.PaceZonesLTSP, pace, 1, 3)
}

func matchesThreshold(zones *personalmetrics.TrainingZones, pace, hr float64) bool {
	return withinHR(zones.HRZonesLTHR, hr, 2, 3) || withinPace(zones.PaceZonesLTSP, pace, 3, 4)
}

func matchesInterval(zones *personalmetrics.TrainingZones, pace, hr float64) bool {
	return withinHR(zones.HRZonesLTHR, hr, 3, 4) || withinPace(zones.PaceZonesLTSP, pace, 4, 5)
}

func withinHR(zones []personalmetrics.HeartZone, hr float64, startIdx, endIdx int) bool {
	if hr <= 0 || startIdx >= len(zones) {
		return false
	}
	if endIdx >= len(zones) {
		endIdx = len(zones) - 1
	}
	low := ptrValue(zones[startIdx].HR)
	high := ptrValue(zones[endIdx].HR)
	if low == 0 || high == 0 {
		return false
	}
	return hr >= low && hr <= high
}

func withinPace(zones []personalmetrics.PaceZone, pace float64, slowerIdx, fasterIdx int) bool {
	if pace <= 0 || slowerIdx >= len(zones) {
		return false
	}
	if fasterIdx >= len(zones) {
		fasterIdx = len(zones) - 1
	}
	slower := ptrValue(zones[slowerIdx].Pace)
	faster := ptrValue(zones[fasterIdx].Pace)
	if slower == 0 || faster == 0 {
		return false
	}
	return pace <= slower && pace >= faster
}

func value(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func ptrValue(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func trimDate(value string) string {
	if len(value) >= 10 {
		return value[:10]
	}
	return value
}
