package activitysummary

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"coros-fit-mcp/internal/service/coros"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

const (
	runningModeList      = "100,103"
	allRunningModes      = "100,101,102,103"
	trailRunningModeList = "102"
)

type corosDateWindow struct {
	targetDate string
	location   *time.Location
}

func SummarizeLatestCorosActivity(service coros.CorosService) (*ActivitySummary, error) {
	items, err := loadLatestActivitiesByModes(service, splitModeList(allRunningModes))
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("没有可用的高驰活动记录")
	}

	latest := items[0]
	labelID := valueAsString(findAnyValue(latest, "labelId", "id"))
	sportType := valueAsString(findAnyValue(latest, "sportType", "mode"))
	if labelID == "" || sportType == "" {
		return nil, fmt.Errorf("最新活动缺少 labelId 或 sportType")
	}

	detail, err := service.SportsSummary(labelID, sportType)
	if err != nil {
		return nil, fmt.Errorf("获取高驰活动详情失败: %w", err)
	}

	summary := summarizeCorosDetail(latest, detail)
	return &summary, nil
}

func SummarizeCorosDailyActivities(service coros.CorosService, date string) (*DailySummaryResponse, error) {
	return summarizeCorosDailyActivitiesByModes(service, date, runningModeList, "跑步", "coros_daily_running_summaries")
}

func SummarizeCorosDailyTrailRunningActivities(service coros.CorosService, date string) (*DailySummaryResponse, error) {
	return summarizeCorosDailyActivitiesByModes(service, date, trailRunningModeList, "越野跑", "coros_daily_trail_running_summaries")
}

func summarizeCorosDailyActivitiesByModes(service coros.CorosService, date, modeList, sportName, sourceName string) (*DailySummaryResponse, error) {
	window, err := resolveCorosDateWindow(date)
	if err != nil {
		return nil, err
	}

	activities, err := loadCorosActivitiesByDate(service, window, modeList)
	if err != nil {
		return nil, err
	}

	summaries := make([]SummaryResponse, 0, len(activities))
	activitySummaries := make([]ActivitySummary, 0, len(activities))
	for _, activity := range activities {
		summary, err := summarizeSingleCorosActivity(service, activity)
		if err != nil {
			return nil, err
		}
		activitySummaries = append(activitySummaries, *summary)
		summaries = append(summaries, ToResponse(*summary))
	}

	dailySummary := buildDailyAggregate(window.targetDate, activitySummaries, sportName, sourceName)

	return &DailySummaryResponse{
		Date:          window.targetDate,
		Timezone:      beijingTimezone,
		ActivityCount: len(summaries),
		DailySummary:  ToResponse(dailySummary),
		Activities:    summaries,
	}, nil
}

func summarizeSingleCorosActivity(service coros.CorosService, activity map[string]interface{}) (*ActivitySummary, error) {
	labelID := valueAsString(findAnyValue(activity, "labelId", "id"))
	sportType := valueAsString(findAnyValue(activity, "sportType", "mode"))
	if labelID == "" || sportType == "" {
		return nil, fmt.Errorf("活动缺少 labelId 或 sportType")
	}

	detail, err := service.SportsSummary(labelID, sportType)
	if err != nil {
		return nil, fmt.Errorf("获取高驰活动详情失败: %w", err)
	}
	if detail == nil {
		return nil, fmt.Errorf("高驰活动详情为空: %s", labelID)
	}

	summary := summarizeCorosDetail(activity, detail)
	return &summary, nil
}

func loadCorosActivitiesByDate(service coros.CorosService, window corosDateWindow, modeList string) ([]map[string]interface{}, error) {
	var activities []map[string]interface{}
	seen := make(map[string]struct{})

	for _, mode := range splitModeList(modeList) {
		for page := 1; ; page++ {
			list, err := service.ActivityListByModeList(20, page, mode)
			if err != nil {
				return nil, fmt.Errorf("获取高驰活动列表失败: %w", err)
			}

			data, ok := list["data"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("高驰活动列表缺少 data 字段")
			}

			items, ok := data["dataList"].([]interface{})
			if !ok || len(items) == 0 {
				break
			}

			foundOlder := false
			for _, item := range items {
				activity, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				activityDate := corosActivityDate(activity, window.location)
				if activityDate == "" {
					continue
				}

				switch {
				case activityDate == window.targetDate:
					labelID := valueAsString(findAnyValue(activity, "labelId", "id"))
					if labelID != "" {
						if _, exists := seen[labelID]; exists {
							continue
						}
						seen[labelID] = struct{}{}
					}
					activities = append(activities, activity)
				case activityDate < window.targetDate:
					foundOlder = true
				}
			}

			if foundOlder {
				break
			}
		}
	}

	sort.Slice(activities, func(i, j int) bool {
		return corosActivityStart(activities[i], window.location).Before(corosActivityStart(activities[j], window.location))
	})

	return activities, nil
}

func loadLatestActivitiesByModes(service coros.CorosService, modes []string) ([]map[string]interface{}, error) {
	activities := make([]map[string]interface{}, 0, len(modes))
	seen := make(map[string]struct{})

	for _, mode := range modes {
		list, err := service.ActivityListByModeList(20, 1, mode)
		if err != nil {
			return nil, fmt.Errorf("获取高驰活动列表失败: %w", err)
		}

		data, ok := list["data"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("高驰活动列表缺少 data 字段")
		}

		items, ok := data["dataList"].([]interface{})
		if !ok || len(items) == 0 {
			continue
		}

		for _, item := range items {
			activity, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			labelID := valueAsString(findAnyValue(activity, "labelId", "id"))
			if labelID != "" {
				if _, exists := seen[labelID]; exists {
					continue
				}
				seen[labelID] = struct{}{}
			}
			activities = append(activities, activity)
		}
	}

	sort.Slice(activities, func(i, j int) bool {
		return corosActivityStart(activities[i], nil).After(corosActivityStart(activities[j], nil))
	})
	return activities, nil
}

func splitModeList(modeList string) []string {
	parts := strings.Split(modeList, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func summarizeCorosDetail(latest map[string]interface{}, detail *coros.SportsSummaryResult) ActivitySummary {
	summaryData := detail.Summary

	moving := firstFinite(
		corosDurationSeconds(readFloat(latest, "workoutTime", "sportTime")),
		corosDurationSeconds(readFloat(summaryData, "sportTime", "timerTime", "movingTime", "exerciseTime", "totalTimerTime")),
	)
	duration := firstFinite(
		corosDurationSeconds(readFloat(latest, "totalTime", "workoutTime")),
		corosDurationSeconds(readFloat(summaryData, "duration", "totalTime", "elapsedTime", "totalElapsedTime")),
	)
	if duration == nil {
		duration = moving
	}
	pause := corosDurationSeconds(readFloat(summaryData, "pauseTime", "restTime"))
	if pause == nil && isFinitePtr(duration) && isFinitePtr(moving) && *duration >= *moving {
		pause = ptr(*duration - *moving)
	}

	distance := firstFinite(
		corosMeters(readFloat(latest, "distance", "total")),
		corosMeters(readFloat(summaryData, "distance", "totalDistance")),
	)
	avgPace := firstFinite(
		corosPaceSeconds(readFloat(latest, "adjustedPace")),
		corosPaceSeconds(readFloat(summaryData, "avgPace", "averagePace")),
	)
	avgMovingPace := firstFinite(
		corosPaceSeconds(readFloat(latest, "avgSpeed")),
		avgPace,
	)
	bestPace := firstFinite(
		corosPaceSeconds(readFloat(latest, "bestKm", "best", "best500m", "bestLen")),
		corosPaceSeconds(readFloat(summaryData, "bestPace", "fastestPace", "minPace")),
	)
	avgSpeed := paceToSpeed(avgMovingPace)
	maxSpeed := paceToSpeed(bestPace)
	if avgPace == nil && isFinitePtr(distance) && isFinitePtr(moving) && *distance > 0 && *moving > 0 {
		avgPace = ptr(*moving / (*distance / 1000))
	}

	laps := make([]LapSummary, 0, len(detail.LapList))
	for i, lap := range detail.LapList {
		lapMoving := corosDurationSeconds(readFloat(lap, "sportTime", "timerTime", "movingTime", "exerciseTime", "totalTimerTime"))
		if lapMoving == nil {
			lapMoving = corosDurationSeconds(readFloat(lap, "duration", "totalTime", "elapsedTime", "totalElapsedTime"))
		}
		lapDistance := corosMeters(readFloat(lap, "distance", "totalDistance"))
		lapAvgPace := corosPaceSeconds(readFloat(lap, "avgPace", "averagePace"))
		if lapAvgPace == nil && isFinitePtr(lapDistance) && isFinitePtr(lapMoving) && *lapDistance > 0 && *lapMoving > 0 {
			lapAvgPace = ptr(*lapMoving / (*lapDistance / 1000))
		}
		lapMaxPace := corosPaceSeconds(readFloat(lap, "maxSpeed", "bestPace", "maximumSpeed"))

		laps = append(laps, LapSummary{
			Index:               i + 1,
			StartTime:           normalizeTimeString(valueAsString(findAnyValue(lap, "startTime", "beginTime"))),
			DurationSeconds:     lapMoving,
			MovingSeconds:       lapMoving,
			DistanceMeters:      lapDistance,
			AverageHeartRate:    readFloat(lap, "avgHr", "averageHeartRate", "avgHeartRate"),
			MaxHeartRate:        readFloat(lap, "maxHr", "maximumHeartRate", "maxHeartRate"),
			AverageCadence:      readFloat(lap, "avgCadence", "averageCadence"),
			MaxCadence:          readFloat(lap, "maxCadence", "maximumCadence"),
			AveragePower:        readFloat(lap, "avgPower", "averagePower"),
			MaxPower:            readFloat(lap, "maxPower", "maximumPower"),
			AverageSpeedMPS:     paceToSpeed(lapAvgPace),
			MaxSpeedMPS:         paceToSpeed(lapMaxPace),
			AveragePaceSecPerKM: lapAvgPace,
		})
	}

	stepCount := readFloat(latest, "step")
	strideLength := strideLengthMeters(distance, stepCount)

	summary := ActivitySummary{
		Source:                    "coros",
		SourceName:                "coros_latest_activity",
		ActivityID:                valueAsString(findAnyValue(latest, "labelId", "id")),
		Name:                      valueAsString(findAnyValue(latest, "name", "title")),
		SportType:                 humanSportType(firstNonEmptyString(valueAsString(findAnyValue(latest, "sportType")), valueAsString(findAnyValue(summaryData, "sportType", "sportName", "modeName")))),
		StartTime:                 firstNonEmptyString(normalizeCOROSTime(findAnyValue(latest, "startTime")), normalizeCOROSTime(findAnyValue(latest, "date")), normalizeCOROSTime(findAnyValue(summaryData, "startTime", "beginTime"))),
		EndTime:                   firstNonEmptyString(normalizeCOROSTime(findAnyValue(latest, "endTime")), normalizeCOROSTime(findAnyValue(summaryData, "endTime", "finishTime"))),
		DurationSeconds:           duration,
		MovingSeconds:             moving,
		PauseSeconds:              pause,
		DistanceMeters:            distance,
		Calories:                  normalizeCalories(firstFinite(readFloat(latest, "calorie"), readFloat(summaryData, "calorie", "calories"))),
		AscentMeters:              firstFinite(corosMeters(readFloat(latest, "ascent", "totalAscent")), corosMeters(readFloat(summaryData, "totalClimb", "totalAscent", "ascent", "altitudeGain"))),
		DescentMeters:             firstFinite(corosMeters(readFloat(latest, "descent", "totalDescent")), corosMeters(readFloat(summaryData, "totalDescent", "descent"))),
		AverageHeartRate:          firstFinite(readFloat(latest, "avgHr"), readFloat(summaryData, "avgHr", "averageHeartRate", "avgHeartRate")),
		MaxHeartRate:              firstFinite(readFloat(summaryData, "maxHr", "maximumHeartRate", "maxHeartRate"), readFloat(latest, "maxHr")),
		AverageCadence:            firstFinite(readFloat(latest, "avgCadence", "cadence"), readFloat(summaryData, "avgCadence", "averageCadence")),
		MaxCadence:                firstFinite(readFloat(summaryData, "maxCadence", "maximumCadence"), readFloat(latest, "maxCadence")),
		AveragePower:              firstFinite(readFloat(latest, "avgPower"), readFloat(summaryData, "avgPower", "averagePower")),
		MaxPower:                  firstFinite(readFloat(summaryData, "maxPower", "maximumPower"), readFloat(latest, "maxPower")),
		AverageSpeedMPS:           avgSpeed,
		MaxSpeedMPS:               maxSpeed,
		AveragePaceSecPerKM:       avgPace,
		AverageMovingPaceSecPerKM: avgMovingPace,
		BestPaceSecPerKM:          bestPace,
		TrainingLoad:              firstFinite(readFloat(latest, "trainingLoad"), readFloat(summaryData, "trainingLoad", "load")),
		TrainingEffect:            readFloat(summaryData, "trainingEffect", "aerobicTrainingEffect", "totalTrainingEffect"),
		AverageTemperature:        readFloat(summaryData, "avgTemperature", "averageTemperature"),
		AverageStrideLengthMeters: strideLength,
		StepCount:                 stepCount,
		LapCount:                  len(laps),
		Device:                    valueAsString(findAnyValue(latest, "device")),
		Laps:                      laps,
		RawLatestActivity:         latest,
	}
	enrichTerrainMetrics(&summary)

	if summary.Name == "" {
		summary.Name = "最新高驰运动"
	}

	return summary
}

func buildDailyAggregate(date string, activities []ActivitySummary, sportName, sourceName string) ActivitySummary {
	daily := ActivitySummary{
		Source:     "coros",
		SourceName: sourceName,
		Name:       fmt.Sprintf("%s %s汇总", date, sportName),
		SportType:  sportName,
	}

	if len(activities) == 0 {
		return daily
	}

	var (
		totalDuration, totalMoving, totalDistance, totalCalories, totalAscent, totalDescent, totalLoad, totalSteps float64
		weightedHeart, weightedCadence, weightedPower, weightedTemp                                                float64
		earliestStart, latestEnd                                                                                   string
		bestPace                                                                                                   *float64
		maxHeart, maxCadence, maxPower                                                                             *float64
	)

	for i, activity := range activities {
		totalDuration += deref(activity.DurationSeconds)
		totalMoving += deref(activity.MovingSeconds)
		totalDistance += deref(activity.DistanceMeters)
		totalCalories += deref(activity.Calories)
		totalAscent += deref(activity.AscentMeters)
		totalDescent += deref(activity.DescentMeters)
		totalLoad += deref(activity.TrainingLoad)
		totalSteps += deref(activity.StepCount)

		weight := firstPositive(activity.MovingSeconds, activity.DurationSeconds)
		if isFinitePtr(weight) && *weight > 0 {
			weightedHeart += deref(activity.AverageHeartRate) * *weight
			weightedCadence += deref(activity.AverageCadence) * *weight
			weightedPower += deref(activity.AveragePower) * *weight
			weightedTemp += deref(activity.AverageTemperature) * *weight
		}

		maxHeart = maxFinite(maxHeart, activity.MaxHeartRate)
		maxCadence = maxFinite(maxCadence, activity.MaxCadence)
		maxPower = maxFinite(maxPower, activity.MaxPower)
		bestPace = minFinite(bestPace, activity.BestPaceSecPerKM)

		if i == 0 || (activity.StartTime != "" && activity.StartTime < earliestStart) {
			if activity.StartTime != "" {
				earliestStart = activity.StartTime
			}
		}
		if i == 0 || (activity.EndTime != "" && activity.EndTime > latestEnd) {
			if activity.EndTime != "" {
				latestEnd = activity.EndTime
			}
		}
	}

	var weightedDuration *float64
	if totalMoving > 0 {
		weightedDuration = ptr(totalMoving)
	} else if totalDuration > 0 {
		weightedDuration = ptr(totalDuration)
	}

	daily.StartTime = earliestStart
	daily.EndTime = latestEnd
	daily.DurationSeconds = finitePtr(totalDuration)
	daily.MovingSeconds = finitePtr(totalMoving)
	if totalDuration >= totalMoving && totalMoving > 0 {
		daily.PauseSeconds = finitePtr(totalDuration - totalMoving)
	}
	daily.DistanceMeters = finitePtr(totalDistance)
	daily.Calories = finitePtr(totalCalories)
	daily.AscentMeters = finitePtr(totalAscent)
	daily.DescentMeters = finitePtr(totalDescent)
	daily.TrainingLoad = finitePtr(totalLoad)
	daily.StepCount = finitePtr(totalSteps)
	daily.MaxHeartRate = maxHeart
	daily.MaxCadence = maxCadence
	daily.MaxPower = maxPower
	daily.BestPaceSecPerKM = bestPace
	daily.LapCount = len(activities)

	if isFinitePtr(weightedDuration) && *weightedDuration > 0 {
		daily.AverageHeartRate = finitePtr(weightedHeart / *weightedDuration)
		daily.AverageCadence = finitePtr(weightedCadence / *weightedDuration)
		daily.AveragePower = finitePtr(weightedPower / *weightedDuration)
		daily.AverageTemperature = finitePtr(weightedTemp / *weightedDuration)
	}

	if totalDistance > 0 && totalDuration > 0 {
		daily.AveragePaceSecPerKM = finitePtr(totalDuration / (totalDistance / 1000))
	}
	if totalDistance > 0 && totalMoving > 0 {
		daily.AverageMovingPaceSecPerKM = finitePtr(totalMoving / (totalDistance / 1000))
		daily.AverageSpeedMPS = paceToSpeed(daily.AverageMovingPaceSecPerKM)
	}
	if bestPace != nil {
		daily.MaxSpeedMPS = paceToSpeed(bestPace)
	}
	daily.AverageStrideLengthMeters = strideLengthMeters(daily.DistanceMeters, daily.StepCount)
	enrichTerrainMetrics(&daily)

	return daily
}

func enrichTerrainMetrics(summary *ActivitySummary) {
	if summary == nil {
		return
	}

	if isFinitePtr(summary.AscentMeters) && isFinitePtr(summary.DistanceMeters) && *summary.DistanceMeters > 0 {
		summary.ElevationGainPerKM = ptr(*summary.AscentMeters / (*summary.DistanceMeters / 1000))
	}
	if isFinitePtr(summary.AscentMeters) && isFinitePtr(summary.MovingSeconds) && *summary.MovingSeconds > 0 {
		summary.VerticalAscentPerHour = ptr(*summary.AscentMeters / (*summary.MovingSeconds / 3600))
		summary.TimePer100MAscentSeconds = ptr(*summary.MovingSeconds / (*summary.AscentMeters / 100))
	}
	if isFinitePtr(summary.MovingSeconds) && isFinitePtr(summary.DurationSeconds) && *summary.DurationSeconds > 0 {
		summary.MovingRatio = ptr(*summary.MovingSeconds / *summary.DurationSeconds)
	}
}

func corosTime(value *float64) *float64 {
	if !isFinitePtr(value) {
		return nil
	}
	out := *value
	// COROS activity detail commonly uses centiseconds for durations.
	if out >= 100 || math.Mod(out, 100) == 0 {
		out = out / 100
	}
	return &out
}

func corosDurationSeconds(value *float64) *float64 {
	if !isFinitePtr(value) {
		return nil
	}
	out := *value
	if out > 100000 {
		out = out / 100
	}
	return &out
}

func corosMeters(value *float64) *float64 {
	if !isFinitePtr(value) {
		return nil
	}
	out := *value
	if out > 100000 {
		out = out / 100
	}
	return &out
}

func corosDistance(value *float64) *float64 {
	if !isFinitePtr(value) {
		return nil
	}
	out := *value
	// COROS activity detail commonly uses centimeters for distance/elevation.
	if out > 1000 {
		out = out / 100
	}
	return &out
}

func corosSpeed(value *float64) *float64 {
	if !isFinitePtr(value) {
		return nil
	}
	out := *value
	// Some COROS fields use mm/s or cm/s; normalize to m/s with a simple heuristic.
	if out > 100 {
		out = out / 1000
	} else if out > 20 {
		out = out / 100
	}
	return &out
}

func corosPace(value *float64) *float64 {
	if !isFinitePtr(value) {
		return nil
	}
	out := *value
	// Pace in COROS payloads is usually seconds/km. If it comes as centiseconds, normalize.
	if out > 2000 {
		out = out / 100
	}
	return &out
}

func corosPaceSeconds(value *float64) *float64 {
	if !isFinitePtr(value) {
		return nil
	}
	out := *value
	if out > 100000 {
		out = out / 100
	}
	return &out
}

func normalizeCalories(value *float64) *float64 {
	if !isFinitePtr(value) {
		return nil
	}
	out := *value
	if out > 10000 {
		out = out / 1000
	}
	return &out
}

func paceToSpeed(pace *float64) *float64 {
	if !isFinitePtr(pace) || *pace <= 0 {
		return nil
	}
	out := 1000 / *pace
	return &out
}

func finitePtr(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v == 0 {
		return nil
	}
	return &v
}

func firstFinite(values ...*float64) *float64 {
	for _, value := range values {
		if isFinitePtr(value) {
			return value
		}
	}
	return nil
}

func firstPositive(values ...*float64) *float64 {
	for _, value := range values {
		if isFinitePtr(value) && *value > 0 {
			return value
		}
	}
	return nil
}

func maxFinite(current, candidate *float64) *float64 {
	if !isFinitePtr(candidate) {
		return current
	}
	if !isFinitePtr(current) || *candidate > *current {
		value := *candidate
		return &value
	}
	return current
}

func minFinite(current, candidate *float64) *float64 {
	if !isFinitePtr(candidate) {
		return current
	}
	if !isFinitePtr(current) || *candidate < *current {
		value := *candidate
		return &value
	}
	return current
}

func deref(value *float64) float64 {
	if !isFinitePtr(value) {
		return 0
	}
	return *value
}

func strideLengthMeters(distanceMeters, stepCount *float64) *float64 {
	if !isFinitePtr(distanceMeters) || !isFinitePtr(stepCount) || *stepCount <= 0 {
		return nil
	}
	out := *distanceMeters / *stepCount
	return &out
}

func humanSportType(value string) string {
	switch strings.TrimSpace(value) {
	case "100":
		return "跑步"
	default:
		return value
	}
}

func normalizeCOROSTime(value interface{}, aliases ...string) string {
	switch v := value.(type) {
	case nil:
		return ""
	case float64:
		return normalizeCOROSNumericTime(int64(v))
	case int:
		return normalizeCOROSNumericTime(int64(v))
	case int64:
		return normalizeCOROSNumericTime(v)
	case string:
		return normalizeTimeString(v)
	default:
		return normalizeTimeString(fmt.Sprint(v))
	}
}

func normalizeCOROSNumericTime(v int64) string {
	switch {
	case v > 1_000_000_000:
		return time.Unix(v, 0).Format(time.RFC3339)
	case v > 20_000_000:
		s := fmt.Sprintf("%08d", v)
		if t, err := time.Parse("20060102", s); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return fmt.Sprint(v)
}

func resolveCorosDateWindow(date string) (corosDateWindow, error) {
	location, err := time.LoadLocation(beijingTimezone)
	if err != nil {
		return corosDateWindow{}, fmt.Errorf("加载时区失败: %w", err)
	}

	if strings.TrimSpace(date) == "" {
		date = time.Now().In(location).Format("2006-01-02")
	}

	if _, err := time.ParseInLocation("2006-01-02", date, location); err != nil {
		return corosDateWindow{}, fmt.Errorf("date 参数格式错误，必须是 YYYY-MM-DD")
	}

	return corosDateWindow{
		targetDate: date,
		location:   location,
	}, nil
}

func corosActivityDate(activity map[string]interface{}, location *time.Location) string {
	start := corosActivityStart(activity, location)
	if !start.IsZero() {
		return start.In(location).Format("2006-01-02")
	}
	return ""
}

func corosActivityStart(activity map[string]interface{}, location *time.Location) time.Time {
	if value := findAnyValue(activity, "startTime"); value != nil {
		if parsed := parseCOROSTime(value, location); !parsed.IsZero() {
			return parsed
		}
	}
	if value := findAnyValue(activity, "date"); value != nil {
		if parsed := parseCOROSTime(value, location); !parsed.IsZero() {
			return parsed
		}
	}
	return time.Time{}
}

func parseCOROSTime(value interface{}, location *time.Location) time.Time {
	switch v := value.(type) {
	case float64:
		return parseCOROSUnixOrDate(int64(v), location)
	case int64:
		return parseCOROSUnixOrDate(v, location)
	case int:
		return parseCOROSUnixOrDate(int64(v), location)
	case string:
		if t, err := time.ParseInLocation("2006-01-02", v, location); err == nil {
			return t
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.In(location)
		}
	}
	return time.Time{}
}

func parseCOROSUnixOrDate(v int64, location *time.Location) time.Time {
	switch {
	case v > 1_000_000_000:
		return time.Unix(v, 0).In(location)
	case v > 20_000_000:
		s := fmt.Sprintf("%08d", v)
		if t, err := time.ParseInLocation("20060102", s, location); err == nil {
			return t
		}
	}
	return time.Time{}
}

func normalizeKey(key string) string {
	return nonAlphaNum.ReplaceAllString(strings.ToLower(key), "")
}

func findAnyValue(data map[string]interface{}, aliases ...string) interface{} {
	if len(data) == 0 {
		return nil
	}

	normalized := make(map[string]interface{}, len(data))
	for key, value := range data {
		normalized[normalizeKey(key)] = value
	}

	for _, alias := range aliases {
		if value, ok := normalized[normalizeKey(alias)]; ok {
			return value
		}
	}

	return nil
}

func readFloat(data map[string]interface{}, aliases ...string) *float64 {
	value := findAnyValue(data, aliases...)
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil
		}
		return &v
	case float32:
		f := float64(v)
		return &f
	case int:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case int32:
		f := float64(v)
		return &f
	case uint:
		f := float64(v)
		return &f
	case uint64:
		f := float64(v)
		return &f
	case jsonNumber:
		f, err := strconv.ParseFloat(string(v), 64)
		if err == nil {
			return &f
		}
	case string:
		if v == "" {
			return nil
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err == nil {
			return &f
		}
	}

	return nil
}

type jsonNumber string

func valueAsString(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func normalizeTimeString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format(time.RFC3339)
		}
	}

	return value
}
