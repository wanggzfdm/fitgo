package coros

// 定义获取高驰API数据的服务接口
type CorosService interface {
	Login() (string, error)
	SportsSummary(labelId, sportType string) (*SportsSummaryResult, error)
	ActivityList(size, pageNumber, modeList int) (map[string]interface{}, error)
	AccountQuery() (*AccountQueryData, error)
	DashboardQuery() (*DashboardQueryData, error)
	DashboardDetailQuery() (*DashboardDetailQueryData, error)
}
