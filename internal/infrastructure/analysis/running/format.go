package running

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"
)

type AIAnalysis struct {
	Training        TrainingSection        `json:"training"`
	Summary         SummarySection         `json:"summary"`
	Metrics         MetricsSection         `json:"metrics"`
	PaceConsistency PaceConsistencySection `json:"pace_consistency"`
	Advice          []AdviceItem           `json:"advice"`
	NextPlan        NextPlanSection        `json:"next_plan"`
}

type TrainingSection struct {
	Type     *string `json:"type"`
	Date     *string `json:"date"`
	Location *string `json:"location"`
}

type SummarySection struct {
	TotalDistanceM *float64 `json:"total_distance_m"`
	TotalTimeCS    *float64 `json:"total_time_cs"`
	TotalPauseTime *float64 `json:"total_pause_time_cs"`
	AvgPaceSPKM    *float64 `json:"avg_pace_s_per_km"`
	BestPaceSPKM   *float64 `json:"best_pace_s_per_km"`
	Insight        string   `json:"insight"`
	KeyConclusion  string   `json:"key_conclusion"`
}

type MetricsSection struct {
	AvgHeartRateBPM *float64 `json:"avg_heart_rate_bpm"`
	AvgCadenceSPM   *float64 `json:"avg_cadence_spm"`
	AvgPowerW       *float64 `json:"avg_power_w"`
	TrainingLoad    *float64 `json:"training_load"`
	Note            string   `json:"note"`
}

type PaceConsistencySection struct {
	Segments []PaceSegment `json:"segments"`
	Summary  string        `json:"summary"`
}

type PaceSegment struct {
	Segment         string   `json:"segment"`
	DistanceM       *float64 `json:"distance_m"`
	AvgPaceSPKM     *float64 `json:"avg_pace_s_per_km"`
	AvgHeartRateBPM *float64 `json:"avg_heart_rate_bpm"`
	Note            string   `json:"note"`
}

type AdviceItem struct {
	Priority int    `json:"priority"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type NextPlanSection struct {
	Goal      string `json:"goal"`
	PaceRange string `json:"pace_range"`
	Focus     string `json:"focus"`
}

func parseAIAnalysis(raw string) (*AIAnalysis, error) {
	payload, err := extractJSON(raw)
	if err != nil {
		return nil, err
	}

	var analysis AIAnalysis
	if err := json.Unmarshal([]byte(payload), &analysis); err != nil {
		return nil, fmt.Errorf("解析AI JSON失败: %w", err)
	}
	return &analysis, nil
}

func extractJSON(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("AI 返回为空")
	}

	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimLeft(trimmed, " \n\r\t")
		if idx := strings.Index(trimmed, "\n"); idx >= 0 {
			trimmed = trimmed[idx+1:]
		}
		if end := strings.LastIndex(trimmed, "```"); end >= 0 {
			trimmed = trimmed[:end]
		}
		trimmed = strings.TrimSpace(trimmed)
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start == -1 || end == -1 || end < start {
		return "", fmt.Errorf("未找到JSON对象")
	}

	return strings.TrimSpace(trimmed[start : end+1]), nil
}

type LapRow struct {
	Index           int
	DistanceM       *float64
	DurationCS      *float64
	AvgPaceSPKM     *float64
	AvgSpeed        *float64
	AvgHeartRateBPM *float64
	AvgCadenceSPM   *float64
	AvgPowerW       *float64
	Note            string
}

func renderAnalysisHTML(analysis *AIAnalysis, laps []map[string]interface{}) string {
	var builder strings.Builder
	builder.WriteString("<!doctype html>\n")
	builder.WriteString("<html lang=\"zh-CN\">\n<head>\n")
	builder.WriteString("  <meta charset=\"utf-8\" />\n")
	builder.WriteString("  <meta name=\"viewport\" content=\"width=device-width,initial-scale=1\" />\n")
	builder.WriteString("  <title>跑步训练分析报告</title>\n")
	builder.WriteString("  <style>\n")
	builder.WriteString("    :root{--bg:#ffffff;--fg:#111827;--muted:#6b7280;--line:#e5e7eb;--card:#f9fafb;--tag:#eef2ff;--tagfg:#3730a3;--radius:14px;}\n")
	builder.WriteString("    *{box-sizing:border-box}\n")
	builder.WriteString("    body{margin:0;padding:24px;font-family:ui-sans-serif,system-ui,-apple-system,\"Segoe UI\",Roboto,\"PingFang SC\",\"Hiragino Sans GB\",\"Microsoft YaHei\",Arial,\"Noto Sans\",\"Helvetica Neue\",sans-serif;color:var(--fg);background:var(--bg);line-height:1.55;}\n")
	builder.WriteString("    .container{max-width:980px;margin:0 auto}\n")
	builder.WriteString("    header{display:flex;flex-direction:column;gap:10px;padding:18px;border:1px solid var(--line);border-radius:var(--radius);background:linear-gradient(180deg,#fff,#fbfbff);}\n")
	builder.WriteString("    h1{margin:0;font-size:22px}\n")
	builder.WriteString("    .meta{display:flex;flex-wrap:wrap;gap:10px;color:var(--muted);font-size:13px;}\n")
	builder.WriteString("    .pill{display:inline-flex;align-items:center;gap:6px;padding:6px 10px;border-radius:999px;background:var(--tag);color:var(--tagfg);border:1px solid #e0e7ff;font-weight:600;}\n")
	builder.WriteString("    section{margin-top:16px}\n")
	builder.WriteString("    .section-title{display:flex;align-items:center;gap:8px;margin:0 0 10px 0;font-size:16px;}\n")
	builder.WriteString("    .card{border:1px solid var(--line);border-radius:var(--radius);overflow:hidden;background:#fff;}\n")
	builder.WriteString("    .table-wrap{width:100%;overflow-x:auto;-webkit-overflow-scrolling:touch;}\n")
	builder.WriteString("    table{width:100%;border-collapse:separate;border-spacing:0;table-layout:fixed;}\n")
	builder.WriteString("    thead th{font-size:13px;color:var(--muted);text-align:left;background:var(--card);border-bottom:1px solid var(--line);padding:10px 12px;}\n")
	builder.WriteString("    tbody td{padding:12px;border-bottom:1px solid var(--line);vertical-align:top;word-break:break-word;}\n")
	builder.WriteString("    tbody tr:last-child td{border-bottom:none}\n")
	builder.WriteString("    .kv td{font-size:18px;font-weight:700}\n")
	builder.WriteString("    .kv small{display:block;font-size:12px;font-weight:600;color:var(--muted);margin-top:3px}\n")
	builder.WriteString("    .note{padding:12px;border-top:1px solid var(--line);background:#fff;}\n")
	builder.WriteString("    .note-title{font-weight:800;margin:0 0 6px 0}\n")
	builder.WriteString("    .note p{margin:0;color:#374151}\n")
	builder.WriteString("    .advice{display:grid;gap:10px;padding:12px;}\n")
	builder.WriteString("    .advice .summary{padding:12px;border:1px dashed #d1d5db;border-radius:12px;background:#fcfcfd;}\n")
	builder.WriteString("    .advice .summary b{display:block;margin-bottom:6px}\n")
	builder.WriteString("    .advice .summary p{margin:0;color:#374151;}\n")
	builder.WriteString("    .advice-list{display:grid;gap:10px;}\n")
	builder.WriteString("    .advice-item{border:1px solid var(--line);border-radius:12px;padding:12px;background:#fff;}\n")
	builder.WriteString("    .advice-item .head{display:flex;gap:8px;align-items:center;margin-bottom:6px;}\n")
	builder.WriteString("    .advice-item .priority{font-weight:700;color:#111827;background:#eef2ff;border:1px solid #e0e7ff;border-radius:999px;padding:2px 8px;font-size:12px;}\n")
	builder.WriteString("    .advice-item .title{font-weight:700;}\n")
	builder.WriteString("    .advice-item .detail{margin:0;color:#374151;}\n")
	builder.WriteString("    .muted{color:var(--muted)}\n")
	builder.WriteString("    @media (max-width: 680px){\n")
	builder.WriteString("      body{padding:14px;}\n")
	builder.WriteString("      header{padding:14px;}\n")
	builder.WriteString("      h1{font-size:18px;}\n")
	builder.WriteString("      .section-title{font-size:15px;}\n")
	builder.WriteString("      thead th, tbody td{padding:8px 10px;font-size:13px;}\n")
	builder.WriteString("      .kv td{font-size:15px;}\n")
	builder.WriteString("      .pill{font-size:12px;padding:4px 8px;}\n")
	builder.WriteString("      .advice{padding:10px;}\n")
	builder.WriteString("      .advice .summary{padding:10px;}\n")
	builder.WriteString("      .advice-item{padding:10px;}\n")
	builder.WriteString("    }\n")
	builder.WriteString("  </style>\n")
	builder.WriteString("</head>\n<body>\n")
	builder.WriteString("  <div class=\"container\">\n")
	builder.WriteString("    <header>\n")
	builder.WriteString("      <h1>🏃 跑步训练分析报告</h1>\n")
	builder.WriteString("      <div class=\"meta\">\n")
	builder.WriteString(fmt.Sprintf("        <span class=\"pill\">训练类型：%s</span>\n", escape(formatOptionalText(analysis.Training.Type))))
	builder.WriteString(fmt.Sprintf("        <span class=\"pill\">日期：%s</span>\n", escape(formatOptionalText(analysis.Training.Date))))
	builder.WriteString(fmt.Sprintf("        <span class=\"pill\">地点：%s</span>\n", escape(formatOptionalText(analysis.Training.Location))))
	builder.WriteString("      </div>\n")
	builder.WriteString("    </header>\n")

	builder.WriteString("    <section>\n")
	builder.WriteString("      <h2 class=\"section-title\">🧭 一、运动概况（Overview）</h2>\n")
	builder.WriteString("      <div class=\"card\">\n")
	builder.WriteString("        <div class=\"table-wrap\">\n")
	builder.WriteString("        <table aria-label=\"overview\">\n")
	builder.WriteString("          <thead><tr><th>总距离</th><th>运动时间</th><th>平均配速</th><th>最快配速</th><th>总暂停时间</th></tr></thead>\n")
	builder.WriteString("          <tbody class=\"kv\"><tr>")
	builder.WriteString(fmt.Sprintf("<td>%s<small>Distance</small></td>", escape(formatDistance(analysis.Summary.TotalDistanceM))))
	builder.WriteString(fmt.Sprintf("<td>%s<small>Time</small></td>", escape(formatTime(analysis.Summary.TotalTimeCS))))
	builder.WriteString(fmt.Sprintf("<td>%s<small>Avg Pace</small></td>", escape(formatPace(summaryAvgPace(analysis.Summary)))))
	builder.WriteString(fmt.Sprintf("<td>%s<small>Best Pace</small></td>", escape(formatPace(analysis.Summary.BestPaceSPKM))))
	builder.WriteString(fmt.Sprintf("<td>%s<small>Paused</small></td>", escape(formatTime(analysis.Summary.TotalPauseTime))))
	builder.WriteString("</tr></tbody>\n")
	builder.WriteString("        </table>\n")
	builder.WriteString("        </div>\n")
	builder.WriteString("        <div class=\"note\"><div class=\"note-title\">🧠 核心洞察</div><p>")
	builder.WriteString(escape(formatText(analysis.Summary.Insight)))
	builder.WriteString("</p></div>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("    </section>\n")

	builder.WriteString("    <section>\n")
	builder.WriteString("      <h2 class=\"section-title\">📊 二、全程关键指标（Metrics）</h2>\n")
	builder.WriteString("      <div class=\"card\">\n")
	builder.WriteString("        <div class=\"table-wrap\">\n")
	builder.WriteString("        <table aria-label=\"metrics\">\n")
	builder.WriteString("          <thead><tr><th>平均心率</th><th>平均步频</th><th>平均功率</th><th>训练负荷</th></tr></thead>\n")
	builder.WriteString("          <tbody class=\"kv\"><tr>")
	builder.WriteString(fmt.Sprintf("<td>%s bpm<small>Avg HR</small></td>", escape(formatNumber(analysis.Metrics.AvgHeartRateBPM))))
	builder.WriteString(fmt.Sprintf("<td>%s spm<small>Cadence</small></td>", escape(formatNumber(analysis.Metrics.AvgCadenceSPM))))
	builder.WriteString(fmt.Sprintf("<td>%s W<small>Power</small></td>", escape(formatNumber(analysis.Metrics.AvgPowerW))))
	builder.WriteString(fmt.Sprintf("<td>%s<small>Load</small></td>", escape(formatNumber(analysis.Metrics.TrainingLoad))))
	builder.WriteString("</tr></tbody>\n")
	builder.WriteString("        </table>\n")
	builder.WriteString("        </div>\n")
	builder.WriteString("        <div class=\"note\"><div class=\"note-title\">📈 指标解读</div><p>")
	builder.WriteString(escape(formatText(analysis.Metrics.Note)))
	builder.WriteString("</p></div>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("    </section>\n")

	builder.WriteString("    <section>\n")
	builder.WriteString("      <h2 class=\"section-title\">📈 三、配速稳定性分析（Pace Consistency）</h2>\n")
	builder.WriteString("      <div class=\"card\">\n")
	builder.WriteString("        <div class=\"note\">\n")
	builder.WriteString("          <div class=\"note-title\">结论</div>\n")
	builder.WriteString("          <p>")
	builder.WriteString(escape(formatText(analysis.PaceConsistency.Summary)))
	builder.WriteString("</p>\n")
	builder.WriteString("        </div>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("    </section>\n")

	builder.WriteString("    <section>\n")
	builder.WriteString("      <h2 class=\"section-title\">🏃 四、全程每圈数据（Lap Details）</h2>\n")
	builder.WriteString("      <div class=\"card\">\n")
	builder.WriteString("        <div class=\"table-wrap\">\n")
	builder.WriteString("        <table aria-label=\"lap-details\">\n")
	builder.WriteString("          <thead><tr><th style=\"width:70px\">圈次</th><th style=\"width:120px\">距离</th><th style=\"width:140px\">配速</th><th style=\"width:140px\">平均心率</th><th style=\"width:140px\">平均步频</th><th style=\"width:140px\">平均功率</th></tr></thead>\n")
	builder.WriteString("          <tbody>\n")
	lapRows := buildLapRows(laps)
	if len(lapRows) == 0 {
		builder.WriteString("            <tr><td>-</td><td>-</td><td>-</td><td>-</td><td>-</td><td>-</td></tr>\n")
	} else {
		for _, row := range lapRows {
			builder.WriteString("            <tr>")
			builder.WriteString(fmt.Sprintf("<td>%d</td>", row.Index))
			builder.WriteString(fmt.Sprintf("<td>%s</td>", escape(formatDistance(row.DistanceM))))
			builder.WriteString(fmt.Sprintf("<td>%s</td>", escape(formatPace(selectPace(row.AvgPaceSPKM, row.DistanceM, row.DurationCS)))))
			builder.WriteString(fmt.Sprintf("<td>%s bpm</td>", escape(formatNumber(row.AvgHeartRateBPM))))
			builder.WriteString(fmt.Sprintf("<td>%s spm</td>", escape(formatNumber(row.AvgCadenceSPM))))
			builder.WriteString(fmt.Sprintf("<td>%s W</td>", escape(formatNumber(row.AvgPowerW))))
			builder.WriteString("</tr>\n")
		}
	}
	builder.WriteString("          </tbody>\n")
	builder.WriteString("        </table>\n")
	builder.WriteString("        </div>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("    </section>\n")

	builder.WriteString("    <section>\n")
	builder.WriteString("      <h2 class=\"section-title\">💡 五、训练总结与改进建议（Actionable Advice）</h2>\n")
	builder.WriteString("      <div class=\"card\">\n")
	builder.WriteString("        <div class=\"advice\">\n")
	builder.WriteString("          <div class=\"summary\"><b>🎯 关键结论</b><p>")
	builder.WriteString(escape(formatText(summaryKeyConclusion(analysis.Summary))))
	builder.WriteString("</p></div>\n")
	builder.WriteString("          <div class=\"advice-list\">\n")
	if len(analysis.Advice) == 0 {
		builder.WriteString("            <div class=\"advice-item\"><div class=\"head\"><span class=\"priority\">-</span><span class=\"title\">暂无建议</span></div><p class=\"detail\">-</p></div>\n")
	} else {
		for _, item := range analysis.Advice {
			title := strings.TrimSpace(item.Title)
			detail := strings.TrimSpace(item.Detail)
			if title == "" && detail == "" {
				continue
			}
			if title == "" {
				title = "建议"
			}
			priority := "-"
			if item.Priority > 0 {
				priority = fmt.Sprintf("P%d", item.Priority)
			}
			builder.WriteString("            <div class=\"advice-item\">")
			builder.WriteString("              <div class=\"head\">")
			builder.WriteString(fmt.Sprintf("<span class=\"priority\">%s</span>", escape(priority)))
			builder.WriteString(fmt.Sprintf("<span class=\"title\">%s</span>", escape(title)))
			builder.WriteString("</div>")
			builder.WriteString(fmt.Sprintf("<p class=\"detail\">%s</p>", escape(formatText(detail))))
			builder.WriteString("</div>\n")
		}
	}
	builder.WriteString("          </div>\n")
	builder.WriteString("        </div>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("    </section>\n")

	builder.WriteString("    <section>\n")
	builder.WriteString("      <h2 class=\"section-title\">📌 六、下次训练建议（Next Plan）</h2>\n")
	builder.WriteString("      <div class=\"card\">\n")
	builder.WriteString("        <div class=\"table-wrap\">\n")
	builder.WriteString("        <table aria-label=\"next-plan\">\n")
	builder.WriteString("          <thead><tr><th style=\"width:160px\">训练目标</th><th>建议内容</th></tr></thead>\n")
	builder.WriteString("          <tbody>\n")
	builder.WriteString(fmt.Sprintf("            <tr><td><b>训练目标</b></td><td>%s</td></tr>\n", escape(formatText(analysis.NextPlan.Goal))))
	builder.WriteString(fmt.Sprintf("            <tr><td><b>配速区间</b></td><td>%s</td></tr>\n", escape(formatPaceRange(analysis.NextPlan.PaceRange))))
	builder.WriteString(fmt.Sprintf("            <tr><td><b>重点关注</b></td><td>%s</td></tr>\n", escape(formatText(analysis.NextPlan.Focus))))
	builder.WriteString("          </tbody>\n")
	builder.WriteString("        </table>\n")
	builder.WriteString("        </div>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("    </section>\n")

	builder.WriteString("  </div>\n</body>\n</html>")
	return builder.String()
}

func renderAnalysisMarkdown(analysis *AIAnalysis, laps []map[string]interface{}) string {
	var builder strings.Builder

	builder.WriteString("# 🏃 跑步训练分析报告\n\n")
	builder.WriteString(fmt.Sprintf("> 训练类型：%s  \n", formatOptionalText(analysis.Training.Type)))
	builder.WriteString(fmt.Sprintf("> 训练日期：%s  \n", formatOptionalText(analysis.Training.Date)))
	builder.WriteString(fmt.Sprintf("> 训练地点：%s  \n\n", formatOptionalText(analysis.Training.Location)))
	builder.WriteString("---\n\n")

	builder.WriteString("## 🧭 一、运动概况（Overview）\n\n")
	builder.WriteString("| 总距离 | 运动时间 | 平均配速 | 最快配速 | 总暂停时间 |\n")
	builder.WriteString("| --- | --- | --- | --- | --- |\n")
	builder.WriteString(fmt.Sprintf(
		"| **%s** | **%s** | **%s** | **%s** | **%s** |\n\n",
		formatDistance(analysis.Summary.TotalDistanceM),
		formatTime(analysis.Summary.TotalTimeCS),
		formatPace(summaryAvgPace(analysis.Summary)),
		formatPace(analysis.Summary.BestPaceSPKM),
		formatTime(analysis.Summary.TotalPauseTime),
	))
	builder.WriteString("| 🧠 核心洞察 |\n")
	builder.WriteString("| --- |\n")
	builder.WriteString(fmt.Sprintf("| %s |\n\n", formatText(analysis.Summary.Insight)))

	builder.WriteString("## 📊 二、全程关键指标（Metrics）\n\n")
	builder.WriteString("| 平均心率 | 平均步频 | 平均功率 | 训练负荷 |\n")
	builder.WriteString("| --- | --- | --- | --- |\n")
	builder.WriteString(fmt.Sprintf(
		"| **%s bpm** | **%s spm** | **%s W** | **%s** |\n\n",
		formatNumber(analysis.Metrics.AvgHeartRateBPM),
		formatNumber(analysis.Metrics.AvgCadenceSPM),
		formatNumber(analysis.Metrics.AvgPowerW),
		formatNumber(analysis.Metrics.TrainingLoad),
	))
	builder.WriteString("| 📈 指标解读 |\n")
	builder.WriteString("| --- |\n")
	builder.WriteString(fmt.Sprintf("| %s |\n\n", formatText(analysis.Metrics.Note)))

	builder.WriteString("## 📈 三、配速稳定性分析（Pace Consistency）\n\n")
	builder.WriteString("| 阶段 | 距离 | 平均配速 | 平均心率 | 说明 |\n")
	builder.WriteString("| --- | --- | --- | --- | --- |\n")
	if len(analysis.PaceConsistency.Segments) == 0 {
		builder.WriteString("| - | - | - | - | - |\n")
	} else {
		for _, segment := range analysis.PaceConsistency.Segments {
			builder.WriteString(fmt.Sprintf(
				"| %s | %s | %s | %s bpm | %s |\n",
				formatSegment(segment.Segment),
				formatDistance(segment.DistanceM),
				formatPace(segment.AvgPaceSPKM),
				formatNumber(segment.AvgHeartRateBPM),
				formatText(segment.Note),
			))
		}
	}
	builder.WriteString(fmt.Sprintf("| 小结 | - | - | - | %s |\n\n", formatText(analysis.PaceConsistency.Summary)))

	builder.WriteString("## 🏃 四、全程每圈数据（Lap Details）\n\n")
	builder.WriteString("| 圈次 | 距离 | 配速 | 平均心率 | 平均步频 | 平均功率 |\n")
	builder.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	lapRows := buildLapRows(laps)
	if len(lapRows) == 0 {
		builder.WriteString("| - | - | - | - | - | - |\n\n")
	} else {
		for _, row := range lapRows {
			builder.WriteString(fmt.Sprintf(
				"| %d | %s | %s | %s bpm | %s spm | %s W |\n",
				row.Index,
				formatDistance(row.DistanceM),
				formatPace(selectPace(row.AvgPaceSPKM, row.DistanceM, row.DurationCS)),
				formatNumber(row.AvgHeartRateBPM),
				formatNumber(row.AvgCadenceSPM),
				formatNumber(row.AvgPowerW),
			))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("## 💡 五、训练总结与改进建议（Actionable Advice）\n\n")
	builder.WriteString("| 🎯 关键结论 |\n")
	builder.WriteString("| --- |\n")
	builder.WriteString(fmt.Sprintf("| %s |\n\n", formatText(summaryKeyConclusion(analysis.Summary))))
	if len(analysis.Advice) == 0 {
		builder.WriteString("- **暂无建议**\n")
	} else {
		builder.WriteString("| 优先级 | 改进方向 | 具体建议 |\n")
		builder.WriteString("| --- | --- | --- |\n")
		for _, item := range analysis.Advice {
			title := strings.TrimSpace(item.Title)
			detail := strings.TrimSpace(item.Detail)
			if title == "" && detail == "" {
				continue
			}
			if title == "" {
				title = "建议"
			}
			priority := "-"
			if item.Priority > 0 {
				priority = fmt.Sprintf("P%d", item.Priority)
			}
			builder.WriteString(fmt.Sprintf("| %s | %s | %s |\n", priority, title, formatText(detail)))
		}
	}

	builder.WriteString("\n## 📌 六、下次训练建议（Next Plan）\n\n")
	builder.WriteString("| 训练目标 | 建议内容 |\n")
	builder.WriteString("| --- | --- |\n")
	builder.WriteString(fmt.Sprintf("| 训练目标 | %s |\n", formatText(analysis.NextPlan.Goal)))
	builder.WriteString(fmt.Sprintf("| 配速区间 | %s |\n", formatPaceRange(analysis.NextPlan.PaceRange)))
	builder.WriteString(fmt.Sprintf("| 重点关注 | %s |\n", formatText(analysis.NextPlan.Focus)))

	return builder.String()
}

func formatDistance(value *float64) string {
	if value == nil {
		return "-"
	}
	km := normalizeDistanceKM(*value)
	return fmt.Sprintf("%.2f 公里", km)
}

func formatTime(value *float64) string {
	if value == nil {
		return "-"
	}
	seconds := int(math.Round(*value / 100.0))
	if seconds < 0 {
		seconds = 0
	}
	return formatSeconds(seconds)
}

func formatPace(value *float64) string {
	if value == nil {
		return "-"
	}
	seconds := int(math.Round(*value))
	if seconds <= 0 {
		return "-"
	}
	return fmt.Sprintf("%s /公里", formatSeconds(seconds))
}

func formatSeconds(totalSeconds int) string {
	if totalSeconds < 60 {
		return fmt.Sprintf("%d秒", totalSeconds)
	}
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%d:%02d", minutes, seconds)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

func formatNumber(value *float64) string {
	if value == nil {
		return "-"
	}
	if math.Abs(*value-math.Round(*value)) < 0.0001 {
		return strconv.FormatInt(int64(math.Round(*value)), 10)
	}
	return strconv.FormatFloat(*value, 'f', 2, 64)
}

func formatText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func formatOptionalText(value *string) string {
	if value == nil {
		return "-"
	}
	return formatText(*value)
}

func formatPaceRange(value string) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "-"
	}
	if strings.Contains(raw, ":") {
		return raw
	}
	nums := extractInts(raw)
	if len(nums) == 0 {
		return raw
	}
	parts := make([]string, 0, len(nums))
	for _, num := range nums {
		if num <= 0 {
			continue
		}
		parts = append(parts, formatSeconds(num))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " - ") + " /公里"
}

func extractInts(value string) []int {
	var nums []int
	current := ""
	for _, r := range value {
		if r >= '0' && r <= '9' {
			current += string(r)
			continue
		}
		if current == "" {
			continue
		}
		num, err := strconv.Atoi(current)
		if err == nil {
			nums = append(nums, num)
		}
		current = ""
	}
	if current != "" {
		num, err := strconv.Atoi(current)
		if err == nil {
			nums = append(nums, num)
		}
	}
	return nums
}

func escape(value string) string {
	return html.EscapeString(value)
}

func formatSegment(segment string) string {
	switch strings.ToLower(strings.TrimSpace(segment)) {
	case "start", "begin", "first":
		return "前段"
	case "middle", "mid":
		return "中段"
	case "end", "last":
		return "末段"
	default:
		if segment == "" {
			return "-"
		}
		return segment
	}
}

func normalizeDistanceKM(raw float64) float64 {
	if raw <= 0 {
		return 0
	}
	if raw >= 100000 {
		return raw / 100000.0
	}
	return raw / 1000.0
}

func selectPace(avgPace, distanceM, durationCS *float64) *float64 {
	if avgPace != nil && *avgPace > 0 {
		return avgPace
	}
	if distanceM == nil || durationCS == nil {
		return nil
	}
	km := normalizeDistanceKM(*distanceM)
	if km <= 0 {
		return nil
	}
	seconds := *durationCS / 100.0
	if seconds <= 0 {
		return nil
	}
	value := seconds / km
	return &value
}

func buildLapRows(laps []map[string]interface{}) []LapRow {
	if len(laps) == 0 {
		return nil
	}
	rows := make([]LapRow, 0, len(laps))
	for i, lap := range laps {
		index := i + 1
		if val, ok := getInt(lap, "lapIndex", "index", "lap", "lapNo", "lapNum", "seq", "no"); ok && val > 0 {
			index = val
		}
		rows = append(rows, LapRow{
			Index:           index,
			DistanceM:       getFloat(lap, "distance", "totalDistance", "dist", "distanceM"),
			DurationCS:      getFloat(lap, "duration", "totalTime", "movingTime", "time", "durationTime", "timeTotal"),
			AvgPaceSPKM:     getFloat(lap, "avgPace", "avgPaceSec", "pace", "avgPaceSeconds", "adjustedPace"),
			AvgSpeed:        getFloat(lap, "avgSpeed", "speed", "avgSpeedMps", "avgSpeedKph"),
			AvgHeartRateBPM: getFloat(lap, "avgHeartRate", "averageHeartRate", "avgHr", "hrAvg"),
			AvgCadenceSPM:   getFloat(lap, "avgCadence", "averageCadence", "avgStepFrequency", "cadenceAvg"),
			AvgPowerW:       getFloat(lap, "avgPower", "averagePower", "powerAvg"),
			Note:            "",
		})
	}
	return rows
}

func summaryAvgPace(summary SummarySection) *float64 {
	if summary.AvgPaceSPKM != nil && *summary.AvgPaceSPKM > 0 {
		return summary.AvgPaceSPKM
	}
	return selectPace(nil, summary.TotalDistanceM, summary.TotalTimeCS)
}

func summaryKeyConclusion(summary SummarySection) string {
	key := strings.TrimSpace(summary.KeyConclusion)
	if key != "" {
		return key
	}
	return summary.Insight
}

func getFloat(data map[string]interface{}, keys ...string) *float64 {
	for _, key := range keys {
		value, ok := data[key]
		if !ok {
			continue
		}
		if num, ok := toFloat64(value); ok {
			return &num
		}
	}
	return nil
}

func getInt(data map[string]interface{}, keys ...string) (int, bool) {
	for _, key := range keys {
		value, ok := data[key]
		if !ok {
			continue
		}
		if num, ok := toFloat64(value); ok {
			return int(math.Round(num)), true
		}
	}
	return 0, false
}

func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case json.Number:
		num, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return num, true
	case string:
		val := strings.TrimSpace(v)
		if val == "" {
			return 0, false
		}
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, false
		}
		return num, true
	default:
		return 0, false
	}
}
