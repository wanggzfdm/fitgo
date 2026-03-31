package activitysummary

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

const beijingTimezone = "Asia/Shanghai"

func ptr[T any](v T) *T {
	return &v
}

func isFinitePtr(v *float64) bool {
	return v != nil && !math.IsNaN(*v) && !math.IsInf(*v, 0)
}

func roundPtr(v *float64, places int) *float64 {
	if !isFinitePtr(v) {
		return nil
	}
	factor := math.Pow10(places)
	out := math.Round((*v)*factor) / factor
	return &out
}

func secondsToHMS(seconds *float64) string {
	if !isFinitePtr(seconds) {
		return "-"
	}
	total := int(math.Round(*seconds))
	hours := total / 3600
	minutes := (total % 3600) / 60
	secs := total % 60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

func metersToKM(meters *float64) string {
	if !isFinitePtr(meters) {
		return "-"
	}
	return fmt.Sprintf("%.2f km", *meters/1000)
}

func speedToKPH(speedMPS *float64) string {
	if !isFinitePtr(speedMPS) {
		return "-"
	}
	return fmt.Sprintf("%.2f km/h", *speedMPS*3.6)
}

func paceToText(secondsPerKM *float64) string {
	if !isFinitePtr(secondsPerKM) || *secondsPerKM <= 0 {
		return "-"
	}
	total := int(math.Round(*secondsPerKM))
	return fmt.Sprintf("%d:%02d /km", total/60, total%60)
}

func numberText(v *float64, format string) string {
	if !isFinitePtr(v) {
		return "-"
	}
	return fmt.Sprintf(format, *v)
}

func percentText(v *float64) string {
	if !isFinitePtr(v) {
		return "-"
	}
	return fmt.Sprintf("%.0f%%", *v*100)
}

func isTrailSummary(summary ActivitySummary) bool {
	return summary.SportType == "越野跑"
}

func buildRoadHighlights(summary ActivitySummary) []string {
	highlights := make([]string, 0, 4)

	if isFinitePtr(summary.DistanceMeters) && isFinitePtr(summary.MovingSeconds) && *summary.DistanceMeters >= 1000 {
		highlights = append(highlights, fmt.Sprintf("完成 %s，用时 %s。", metersToKM(summary.DistanceMeters), secondsToHMS(summary.MovingSeconds)))
	}
	if isFinitePtr(summary.AveragePaceSecPerKM) {
		highlights = append(highlights, fmt.Sprintf("平均配速 %s。", paceToText(summary.AveragePaceSecPerKM)))
	} else if isFinitePtr(summary.AverageSpeedMPS) {
		highlights = append(highlights, fmt.Sprintf("平均速度 %s。", speedToKPH(summary.AverageSpeedMPS)))
	}
	if isFinitePtr(summary.AverageHeartRate) {
		line := fmt.Sprintf("平均心率 %.0f bpm", *summary.AverageHeartRate)
		if isFinitePtr(summary.MaxHeartRate) {
			line += fmt.Sprintf("，最高 %.0f bpm", *summary.MaxHeartRate)
		}
		highlights = append(highlights, line+"。")
	}
	if isFinitePtr(summary.TrainingLoad) {
		highlights = append(highlights, fmt.Sprintf("训练负荷 %.0f。", *summary.TrainingLoad))
	}

	return highlights
}

func buildTrailHighlights(summary ActivitySummary) []string {
	highlights := make([]string, 0, 5)

	if isFinitePtr(summary.DistanceMeters) && isFinitePtr(summary.MovingSeconds) {
		line := fmt.Sprintf("完成 %s，移动时间 %s", metersToKM(summary.DistanceMeters), secondsToHMS(summary.MovingSeconds))
		if isFinitePtr(summary.AscentMeters) {
			line += fmt.Sprintf("，累计爬升 %.0f m", *summary.AscentMeters)
		}
		highlights = append(highlights, line+"。")
	}
	if isFinitePtr(summary.VerticalAscentPerHour) {
		highlights = append(highlights, fmt.Sprintf("爬升效率 %.0f m/h。", *summary.VerticalAscentPerHour))
	}
	if isFinitePtr(summary.ElevationGainPerKM) {
		highlights = append(highlights, fmt.Sprintf("单位距离爬升 %.0f m/km。", *summary.ElevationGainPerKM))
	}
	if isFinitePtr(summary.AverageHeartRate) {
		line := fmt.Sprintf("平均心率 %.0f bpm", *summary.AverageHeartRate)
		if isFinitePtr(summary.MaxHeartRate) {
			line += fmt.Sprintf("，最高 %.0f bpm", *summary.MaxHeartRate)
		}
		highlights = append(highlights, line+"。")
	}
	if isFinitePtr(summary.TrainingLoad) {
		highlights = append(highlights, fmt.Sprintf("训练负荷 %.0f。", *summary.TrainingLoad))
	}

	return highlights
}

func buildHighlights(summary ActivitySummary) []string {
	if isTrailSummary(summary) {
		return buildTrailHighlights(summary)
	}
	return buildRoadHighlights(summary)
}

func ToResponse(summary ActivitySummary) SummaryResponse {
	summary.DistanceMeters = roundPtr(summary.DistanceMeters, 2)
	summary.Calories = roundPtr(summary.Calories, 1)
	summary.AscentMeters = roundPtr(summary.AscentMeters, 1)
	summary.DescentMeters = roundPtr(summary.DescentMeters, 1)
	summary.ElevationGainPerKM = roundPtr(summary.ElevationGainPerKM, 1)
	summary.VerticalAscentPerHour = roundPtr(summary.VerticalAscentPerHour, 1)
	summary.TimePer100MAscentSeconds = roundPtr(summary.TimePer100MAscentSeconds, 1)
	summary.MovingRatio = roundPtr(summary.MovingRatio, 3)
	summary.AverageHeartRate = roundPtr(summary.AverageHeartRate, 1)
	summary.MaxHeartRate = roundPtr(summary.MaxHeartRate, 1)
	summary.AverageCadence = roundPtr(summary.AverageCadence, 1)
	summary.MaxCadence = roundPtr(summary.MaxCadence, 1)
	summary.AveragePower = roundPtr(summary.AveragePower, 1)
	summary.MaxPower = roundPtr(summary.MaxPower, 1)
	summary.AverageSpeedMPS = roundPtr(summary.AverageSpeedMPS, 3)
	summary.MaxSpeedMPS = roundPtr(summary.MaxSpeedMPS, 3)
	summary.AveragePaceSecPerKM = roundPtr(summary.AveragePaceSecPerKM, 1)
	summary.AverageMovingPaceSecPerKM = roundPtr(summary.AverageMovingPaceSecPerKM, 1)
	summary.BestPaceSecPerKM = roundPtr(summary.BestPaceSecPerKM, 1)
	summary.TrainingLoad = roundPtr(summary.TrainingLoad, 1)
	summary.TrainingEffect = roundPtr(summary.TrainingEffect, 1)
	summary.AverageTemperature = roundPtr(summary.AverageTemperature, 1)
	summary.AverageStrideLengthMeters = roundPtr(summary.AverageStrideLengthMeters, 2)
	summary.StepCount = roundPtr(summary.StepCount, 0)
	summary.Highlights = buildHighlights(summary)

	var lines []string
	title := summary.Name
	if title == "" {
		title = "运动摘要"
	}
	lines = append(lines, fmt.Sprintf("# %s", title))
	if summary.SportType != "" {
		lines = append(lines, fmt.Sprintf("- 类型：%s", summary.SportType))
	}
	if summary.StartTime != "" {
		lines = append(lines, fmt.Sprintf("- 开始时间：%s", summary.StartTime))
	}

	if isTrailSummary(summary) {
		lines = append(lines, "", "📍 越野概况")
		lines = append(lines, fmt.Sprintf("- 距离：%s", metersToKM(summary.DistanceMeters)))
		lines = append(lines, fmt.Sprintf("- 总时间：%s", secondsToHMS(summary.DurationSeconds)))
		lines = append(lines, fmt.Sprintf("- 移动时间：%s", secondsToHMS(summary.MovingSeconds)))
		lines = append(lines, fmt.Sprintf("- 累计爬升：%s", numberText(summary.AscentMeters, "%.0f m")))
		lines = append(lines, fmt.Sprintf("- 累计下降：%s", numberText(summary.DescentMeters, "%.0f m")))
		lines = append(lines, fmt.Sprintf("- 单位距离爬升：%s", numberText(summary.ElevationGainPerKM, "%.0f m/km")))

		lines = append(lines, "", "⛰️ 地形与效率")
		lines = append(lines, fmt.Sprintf("- 爬升效率：%s", numberText(summary.VerticalAscentPerHour, "%.0f m/h")))
		lines = append(lines, fmt.Sprintf("- 每爬升 100m 用时：%s", numberText(summary.TimePer100MAscentSeconds, "%.0f s")))
		lines = append(lines, fmt.Sprintf("- 移动占比：%s", percentText(summary.MovingRatio)))
		lines = append(lines, fmt.Sprintf("- 平均移动配速：%s", paceToText(summary.AverageMovingPaceSecPerKM)))
		lines = append(lines, fmt.Sprintf("- 最快配速：%s", paceToText(summary.BestPaceSecPerKM)))

		lines = append(lines, "", "❤️ 强度与负荷")
		lines = append(lines, fmt.Sprintf("- 平均心率：%s", numberText(summary.AverageHeartRate, "%.0f bpm")))
		lines = append(lines, fmt.Sprintf("- 最大心率：%s", numberText(summary.MaxHeartRate, "%.0f bpm")))
		lines = append(lines, fmt.Sprintf("- 平均功率：%s", numberText(summary.AveragePower, "%.0f W")))
		lines = append(lines, fmt.Sprintf("- 训练负荷：%s", numberText(summary.TrainingLoad, "%.0f")))
		lines = append(lines, fmt.Sprintf("- 消耗热量：%s", numberText(summary.Calories, "%.0f kcal")))

		lines = append(lines, "", "👣 动作数据")
		lines = append(lines, fmt.Sprintf("- 步频：%s", numberText(summary.AverageCadence, "%.0f spm")))
		lines = append(lines, fmt.Sprintf("- 步幅：%s", numberText(summary.AverageStrideLengthMeters, "%.2f m")))
		lines = append(lines, fmt.Sprintf("- 总步数：%s", numberText(summary.StepCount, "%.0f")))
	} else {
		lines = append(lines, "", "📍 基本信息")
		lines = append(lines, fmt.Sprintf("- 距离：%s", metersToKM(summary.DistanceMeters)))
		lines = append(lines, fmt.Sprintf("- 用时：%s", secondsToHMS(summary.DurationSeconds)))
		lines = append(lines, fmt.Sprintf("- 平均配速：%s", paceToText(summary.AveragePaceSecPerKM)))
		lines = append(lines, fmt.Sprintf("- 平均移动配速：%s", paceToText(summary.AverageMovingPaceSecPerKM)))
		lines = append(lines, fmt.Sprintf("- 最快配速：%s", paceToText(summary.BestPaceSecPerKM)))

		lines = append(lines, "", "❤️ 身体数据")
		lines = append(lines, fmt.Sprintf("- 平均心率：%s", numberText(summary.AverageHeartRate, "%.0f bpm")))
		lines = append(lines, fmt.Sprintf("- 最大心率：%s", numberText(summary.MaxHeartRate, "%.0f bpm")))
		lines = append(lines, fmt.Sprintf("- 步频：%s", numberText(summary.AverageCadence, "%.0f spm")))
		lines = append(lines, fmt.Sprintf("- 步幅：%s", numberText(summary.AverageStrideLengthMeters, "%.2f m")))

		lines = append(lines, "", "🔥 消耗与负荷")
		lines = append(lines, fmt.Sprintf("- 消耗热量：%s", numberText(summary.Calories, "%.0f kcal")))
		lines = append(lines, fmt.Sprintf("- 训练负荷：%s", numberText(summary.TrainingLoad, "%.0f")))
		lines = append(lines, fmt.Sprintf("- 总步数：%s", numberText(summary.StepCount, "%.0f")))
	}

	return SummaryResponse{
		Summary:  summary,
		Markdown: strings.Join(lines, "\n"),
	}
}

func ResponseText(summary ActivitySummary) (string, error) {
	response := ToResponse(summary)
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func DailyResponseText(response DailySummaryResponse) (string, error) {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
