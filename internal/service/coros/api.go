package coros

// CorosSummary 表示高驰运动数据摘要
type CorosSummary struct {
	ID        string  `json:"id"`
	Date      string  `json:"date"`
	Duration  int     `json:"duration"`
	Distance  float64 `json:"distance"`
	Calories  float64 `json:"calories"`
	SportType string  `json:"sport_type"`
}

// 定义获取高驰API数据的服务接口
type CorosService interface {
	Login() (string, error)
	ListCorosSummaries() ([]*CorosSummary, error)
	SportsSummary(labelId, sportType string) (*SportsSummaryResult, error)
	ActivityList(size, pageNumber, modeList int) (map[string]interface{}, error)
}
