package activitysummary

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"

	fitlib "github.com/tormoder/fit"
)

func SummarizeFITPath(path string) (*ActivitySummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 FIT 文件失败: %w", err)
	}
	summary, err := SummarizeFITReader(bytes.NewReader(data), filepath.Base(path))
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func SummarizeFITReader(reader io.Reader, sourceName string) (*ActivitySummary, error) {
	file, err := fitlib.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("解析 FIT 文件失败: %w", err)
	}

	switch file.Type() {
	case fitlib.FileTypeActivity:
		activity, err := file.Activity()
		if err != nil {
			return nil, fmt.Errorf("读取 FIT activity 失败: %w", err)
		}
		summary := summarizeFITActivity(file, activity, sourceName)
		return &summary, nil
	case fitlib.FileTypeActivitySummary:
		activity, err := file.ActivitySummary()
		if err != nil {
			return nil, fmt.Errorf("读取 FIT activity summary 失败: %w", err)
		}
		summary := summarizeFITActivitySummary(file, activity, sourceName)
		return &summary, nil
	default:
		return nil, fmt.Errorf("暂不支持的 FIT 文件类型: %s", file.Type())
	}
}

func summarizeFITActivity(file *fitlib.File, activity *fitlib.ActivityFile, sourceName string) ActivitySummary {
	var session *fitlib.SessionMsg
	if len(activity.Sessions) > 0 {
		session = activity.Sessions[0]
	}

	laps := make([]LapSummary, 0, len(activity.Laps))
	for i, lap := range activity.Laps {
		lapSummary := LapSummary{
			Index:               i + 1,
			StartTime:           normalizeFITTime(lap.StartTime),
			DurationSeconds:     finiteFloat(lap.GetTotalElapsedTimeScaled()),
			MovingSeconds:       finiteFloat(lap.GetTotalTimerTimeScaled()),
			DistanceMeters:      finiteFloat(lap.GetTotalDistanceScaled()),
			AverageHeartRate:    finiteUint8(lap.AvgHeartRate, 0xFF),
			MaxHeartRate:        finiteUint8(lap.MaxHeartRate, 0xFF),
			AverageCadence:      finiteUint8(lap.AvgCadence, 0xFF),
			MaxCadence:          finiteUint8(lap.MaxCadence, 0xFF),
			AveragePower:        finiteUint16(lap.AvgPower, 0xFFFF),
			MaxPower:            finiteUint16(lap.MaxPower, 0xFFFF),
			AverageSpeedMPS:     finiteFloat(lap.GetAvgSpeedScaled()),
			MaxSpeedMPS:         finiteFloat(lap.GetMaxSpeedScaled()),
			AveragePaceSecPerKM: paceFromSpeed(lap.GetAvgSpeedScaled()),
		}
		laps = append(laps, lapSummary)
	}

	summary := ActivitySummary{
		Source:             "fit",
		SourceName:         sourceName,
		ActivityID:         normalizeFITTime(file.FileId.TimeCreated),
		Name:               sourceName,
		SportType:          fitSportType(session),
		StartTime:          normalizeFITTime(sessionStartTime(session)),
		EndTime:            normalizeFITTime(sessionEndTime(session)),
		DurationSeconds:    sessionFloat(session, (*fitlib.SessionMsg).GetTotalElapsedTimeScaled),
		MovingSeconds:      sessionFloat(session, (*fitlib.SessionMsg).GetTotalTimerTimeScaled),
		DistanceMeters:     sessionFloat(session, (*fitlib.SessionMsg).GetTotalDistanceScaled),
		Calories:           sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.TotalCalories }, 0xFFFF),
		AscentMeters:       sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.TotalAscent }, 0xFFFF),
		DescentMeters:      sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.TotalDescent }, 0xFFFF),
		AverageHeartRate:   sessionUint8(session, func(s *fitlib.SessionMsg) uint8 { return s.AvgHeartRate }, 0xFF),
		MaxHeartRate:       sessionUint8(session, func(s *fitlib.SessionMsg) uint8 { return s.MaxHeartRate }, 0xFF),
		AverageCadence:     sessionUint8(session, func(s *fitlib.SessionMsg) uint8 { return s.AvgCadence }, 0xFF),
		MaxCadence:         sessionUint8(session, func(s *fitlib.SessionMsg) uint8 { return s.MaxCadence }, 0xFF),
		AveragePower:       sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.AvgPower }, 0xFFFF),
		MaxPower:           sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.MaxPower }, 0xFFFF),
		AverageSpeedMPS:    sessionFloat(session, (*fitlib.SessionMsg).GetAvgSpeedScaled),
		MaxSpeedMPS:        sessionFloat(session, (*fitlib.SessionMsg).GetMaxSpeedScaled),
		TrainingLoad:       sessionFloat(session, (*fitlib.SessionMsg).GetTrainingStressScoreScaled),
		TrainingEffect:     sessionFloat(session, (*fitlib.SessionMsg).GetTotalTrainingEffectScaled),
		AverageTemperature: sessionInt8(session, func(s *fitlib.SessionMsg) int8 { return s.AvgTemperature }, 0x7F),
		RecordCount:        len(activity.Records),
		LapCount:           len(laps),
		Laps:               laps,
		Device:             fitDeviceName(file),
	}

	summary.AveragePaceSecPerKM = paceFromPtr(summary.AverageSpeedMPS)
	summary.BestPaceSecPerKM = paceFromPtr(summary.MaxSpeedMPS)
	if isFinitePtr(summary.DurationSeconds) && isFinitePtr(summary.MovingSeconds) && *summary.DurationSeconds >= *summary.MovingSeconds {
		summary.PauseSeconds = ptr(*summary.DurationSeconds - *summary.MovingSeconds)
	}
	return summary
}

func summarizeFITActivitySummary(file *fitlib.File, activity *fitlib.ActivitySummaryFile, sourceName string) ActivitySummary {
	var session *fitlib.SessionMsg
	if len(activity.Sessions) > 0 {
		session = activity.Sessions[0]
	}

	laps := make([]LapSummary, 0, len(activity.Laps))
	for i, lap := range activity.Laps {
		laps = append(laps, LapSummary{
			Index:               i + 1,
			StartTime:           normalizeFITTime(lap.StartTime),
			DurationSeconds:     finiteFloat(lap.GetTotalElapsedTimeScaled()),
			MovingSeconds:       finiteFloat(lap.GetTotalTimerTimeScaled()),
			DistanceMeters:      finiteFloat(lap.GetTotalDistanceScaled()),
			AverageHeartRate:    finiteUint8(lap.AvgHeartRate, 0xFF),
			MaxHeartRate:        finiteUint8(lap.MaxHeartRate, 0xFF),
			AverageCadence:      finiteUint8(lap.AvgCadence, 0xFF),
			MaxCadence:          finiteUint8(lap.MaxCadence, 0xFF),
			AveragePower:        finiteUint16(lap.AvgPower, 0xFFFF),
			MaxPower:            finiteUint16(lap.MaxPower, 0xFFFF),
			AverageSpeedMPS:     finiteFloat(lap.GetAvgSpeedScaled()),
			MaxSpeedMPS:         finiteFloat(lap.GetMaxSpeedScaled()),
			AveragePaceSecPerKM: paceFromSpeed(lap.GetAvgSpeedScaled()),
		})
	}

	summary := ActivitySummary{
		Source:             "fit",
		SourceName:         sourceName,
		ActivityID:         normalizeFITTime(file.FileId.TimeCreated),
		Name:               sourceName,
		SportType:          fitSportType(session),
		StartTime:          normalizeFITTime(sessionStartTime(session)),
		EndTime:            normalizeFITTime(sessionEndTime(session)),
		DurationSeconds:    sessionFloat(session, (*fitlib.SessionMsg).GetTotalElapsedTimeScaled),
		MovingSeconds:      sessionFloat(session, (*fitlib.SessionMsg).GetTotalTimerTimeScaled),
		DistanceMeters:     sessionFloat(session, (*fitlib.SessionMsg).GetTotalDistanceScaled),
		Calories:           sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.TotalCalories }, 0xFFFF),
		AscentMeters:       sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.TotalAscent }, 0xFFFF),
		DescentMeters:      sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.TotalDescent }, 0xFFFF),
		AverageHeartRate:   sessionUint8(session, func(s *fitlib.SessionMsg) uint8 { return s.AvgHeartRate }, 0xFF),
		MaxHeartRate:       sessionUint8(session, func(s *fitlib.SessionMsg) uint8 { return s.MaxHeartRate }, 0xFF),
		AverageCadence:     sessionUint8(session, func(s *fitlib.SessionMsg) uint8 { return s.AvgCadence }, 0xFF),
		MaxCadence:         sessionUint8(session, func(s *fitlib.SessionMsg) uint8 { return s.MaxCadence }, 0xFF),
		AveragePower:       sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.AvgPower }, 0xFFFF),
		MaxPower:           sessionUint16(session, func(s *fitlib.SessionMsg) uint16 { return s.MaxPower }, 0xFFFF),
		AverageSpeedMPS:    sessionFloat(session, (*fitlib.SessionMsg).GetAvgSpeedScaled),
		MaxSpeedMPS:        sessionFloat(session, (*fitlib.SessionMsg).GetMaxSpeedScaled),
		TrainingLoad:       sessionFloat(session, (*fitlib.SessionMsg).GetTrainingStressScoreScaled),
		TrainingEffect:     sessionFloat(session, (*fitlib.SessionMsg).GetTotalTrainingEffectScaled),
		AverageTemperature: sessionInt8(session, func(s *fitlib.SessionMsg) int8 { return s.AvgTemperature }, 0x7F),
		LapCount:           len(laps),
		Laps:               laps,
		Device:             fitDeviceName(file),
	}

	summary.AveragePaceSecPerKM = paceFromPtr(summary.AverageSpeedMPS)
	summary.BestPaceSecPerKM = paceFromPtr(summary.MaxSpeedMPS)
	if isFinitePtr(summary.DurationSeconds) && isFinitePtr(summary.MovingSeconds) && *summary.DurationSeconds >= *summary.MovingSeconds {
		summary.PauseSeconds = ptr(*summary.DurationSeconds - *summary.MovingSeconds)
	}
	return summary
}

func finiteFloat(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

func finiteUint8(v uint8, invalid uint8) *float64 {
	if v == invalid {
		return nil
	}
	f := float64(v)
	return &f
}

func finiteUint16(v uint16, invalid uint16) *float64 {
	if v == invalid {
		return nil
	}
	f := float64(v)
	return &f
}

func finiteInt8(v int8, invalid int8) *float64 {
	if v == invalid {
		return nil
	}
	f := float64(v)
	return &f
}

func sessionFloat(session *fitlib.SessionMsg, getter func(*fitlib.SessionMsg) float64) *float64 {
	if session == nil {
		return nil
	}
	return finiteFloat(getter(session))
}

func sessionUint8(session *fitlib.SessionMsg, getter func(*fitlib.SessionMsg) uint8, invalid uint8) *float64 {
	if session == nil {
		return nil
	}
	return finiteUint8(getter(session), invalid)
}

func sessionUint16(session *fitlib.SessionMsg, getter func(*fitlib.SessionMsg) uint16, invalid uint16) *float64 {
	if session == nil {
		return nil
	}
	return finiteUint16(getter(session), invalid)
}

func sessionInt8(session *fitlib.SessionMsg, getter func(*fitlib.SessionMsg) int8, invalid int8) *float64 {
	if session == nil {
		return nil
	}
	return finiteInt8(getter(session), invalid)
}

func paceFromSpeed(speed float64) *float64 {
	if speed <= 0 || math.IsNaN(speed) || math.IsInf(speed, 0) {
		return nil
	}
	pace := 1000 / speed
	return &pace
}

func paceFromPtr(speed *float64) *float64 {
	if !isFinitePtr(speed) {
		return nil
	}
	return paceFromSpeed(*speed)
}

func fitSportType(session *fitlib.SessionMsg) string {
	if session == nil {
		return ""
	}
	if session.Sport.String() != "Invalid" {
		sport := session.Sport.String()
		if session.SubSport.String() != "Invalid" && session.SubSport.String() != "Generic" {
			return fmt.Sprintf("%s/%s", sport, session.SubSport.String())
		}
		return sport
	}
	if session.SubSport.String() != "Invalid" && session.SubSport.String() != "Generic" {
		return session.SubSport.String()
	}
	return ""
}

func sessionStartTime(session *fitlib.SessionMsg) time.Time {
	if session == nil {
		return time.Time{}
	}
	return session.StartTime
}

func sessionEndTime(session *fitlib.SessionMsg) time.Time {
	if session == nil {
		return time.Time{}
	}
	return session.Timestamp
}

func normalizeFITTime(value time.Time) string {
	if value.IsZero() || value.Year() <= 1990 {
		return ""
	}
	return value.Format(time.RFC3339)
}

func fitDeviceName(file *fitlib.File) string {
	if file.FileId.ProductName != "" {
		return file.FileId.ProductName
	}
	return fmt.Sprint(file.FileId.GetProduct())
}
