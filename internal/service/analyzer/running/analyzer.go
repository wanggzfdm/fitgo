package running

import (
	"context"
	"fitgo/internal/service/ai/client"
	aiservice "fitgo/internal/service/ai/service"
	"fitgo/internal/service/coros"
	"fitgo/pkg/config"
	"fmt"
)

// RunAnalyzer 分析运动数据并返回AI分析结果
func RunAnalyzer(userID, sportID string) (string, error) {
	// 1. 加载配置
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		return "", fmt.Errorf("加载配置失败: %v", err)
	}

	// 2. 创建 AI 服务
	aiService, err := aiservice.NewAIService(&cfg.AI)
	if err != nil {
		return "", fmt.Errorf("创建 AI 服务失败: %v", err)
	}

	// 3. 获取运动概要
	sportsSummary, err := coros.NewCorosService().SportsSummary(userID, sportID)
	if err != nil {
		return "", fmt.Errorf("获取运动概要失败: %v", err)
	}

	// 4. 准备提示词
	prompt := fmt.Sprintf(`
你是一个专业的运动数据分析助手。请分析以下高驰运动数据：

【要求】
**严格按照规则转换**。只显示最终格式化时间/距离，**不显示任何原始值和计算过程**。

【时间格式化规则】
所有时间（运动时间、暂停时间、配速等）字段转换为易读格式：
1. **基础转换**: 原始值 ÷ 100 = **总秒数 (S)**。
2. **小时/分钟/秒转换逻辑**:
   - 如果 S < 60秒: S 秒。
   - 如果 60秒 ≤ S < 3600秒 (60分钟): **分:秒** (M:SS)。
   - 如果 S ≥ 3600秒 (1小时): **时:分:秒** (H:MM:SS)。

**特别提醒：请确保秒数转换为分钟和小时的计算准确无误，例如：**
- **错误示例：** 389678 (3896.78 秒) **错误地** 转换为 6:29:28 (小时:分:秒)
- **正确示例：** 389678 (3896.78 秒) **应转换为 1:04:57** (时:分:秒)
- 349038 → 58:10 (小于1小时)
- 756000 → 2:06:00 (大于1小时)

【距离格式化规则】
距离字段转换为易读格式:
- 原始值 ÷ 1000 = 公里。
- 距离智能转换: 单位米，1001179 转换成 10.01公里。

【配速字段说明】
- adjustedPace、avgPace 单位是：秒/公里
   - 转换公式：秒/公里 → 分钟:秒/公里
   - 示例：373秒/公里 = 6分13秒/公里 (373÷60=6余13)

【分析报告模板】
请严格输出以下四个部分内容，并使用Markdown表格格式：
### **一、 🎯 运动表现总结 (Summary)**
[包含总距离、运动时间、平均配速、最快配速、总暂停时间，并给出 1-2 句核心洞察。]

### **二、 📊 关键生理与技术指标 (Metrics)**
[包含平均心率、平均步频、平均功率、训练负荷等，并进行简短分析。]

### **三、 📈 每公里配速稳定性分析 (Pace Consistency)**
[分析前、中、末段的配速和心率变化，总结节奏控制和疲劳出现的关键发现。]

### **四、 💡 针对性改进建议 (Actionable Advice)**
[给出至少 3 条优先级建议：提高速度耐力、强化跑步经济性、优化功率训练。]

运动数据：
每圈数据：%s
总结数据：%s`,
		sportsSummary.LapList, sportsSummary.Summary)
	// 5. 调用AI服务
	response, err := aiService.Chat(context.Background(), []client.ChatMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	})
	if err != nil {
		return "", fmt.Errorf("AI分析失败: %v", err)
	}

	return response, nil
}
