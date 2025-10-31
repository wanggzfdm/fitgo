package coros

import (
	"bytes"
	"encoding/json"
	"fitgo/pkg/config"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// LoginResponse 定义登录响应结构体
type LoginResponse struct {
	ApiCode string `json:"apiCode"`
	Msg     string `json:"message"`
	Result  string `json:"result"`
	Data    struct {
		AccessToken string `json:"accessToken"`
	} `json:"data"`
}

type Lap struct {
	// 根据实际API返回的字段来定义
	// 这里使用map[string]interface{}来接收任意JSON结构
	Data map[string]interface{} `json:"-"`
}

// 用于返回给调用方的数据结构
type SportsSummaryResult struct {
	LapList []map[string]interface{} `json:"lapList"`
	Summary map[string]interface{}   `json:"summary"`
}

type loginform struct {
	Account     int    `json:"account"`
	AccountType int8   `json:"accountType"`
	Pwd         string `json:"pwd"`
}

type corosService struct {
	token       string     // 缓存的 token
	tokenMutex  sync.Mutex // 用于保护 token 的并发访问
	tokenExpire time.Time  // token 过期时间
}

func (s *corosService) ActivityList(size, pageNumber, modeList int) (map[string]interface{}, error) {
	token, loginErr := s.Login()
	if loginErr != nil {
		return nil, loginErr
	}

	// 获取配置文件
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	urlStr := fmt.Sprintf("%s/activity/query?size=%d&pageNumber=%d&modeList=",
		cfg.Coros.Address, size, pageNumber)

	// 3. 创建新的请求
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 4. 添加认证头
	req.Header.Set("accesstoken", token)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15")

	// 5. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体为字符串
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析为 map[string]interface{}
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}

	return result, nil
}

func (s *corosService) ListCorosSummaries() ([]*CorosSummary, error) {
	//TODO implement me
	panic("implement me")
}

func NewCorosService() CorosService {
	return &corosService{}
}

func (s *corosService) Login() (string, error) {
	// 检查是否有未过期的 token
	s.tokenMutex.Lock()
	if s.token != "" && time.Now().Before(s.tokenExpire) {
		token := s.token
		s.tokenMutex.Unlock()
		return token, nil
	}
	s.tokenMutex.Unlock()

	// 获取配置文件
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		return "", fmt.Errorf("加载配置失败: %v", err)
	}

	username := cfg.Coros.Username
	password := cfg.Coros.Password
	loginUrl := fmt.Sprintf("%s/account/login",
		cfg.Coros.Address)

	loginForm := loginform{Account: username, AccountType: 2, Pwd: password}

	// 将结构体序列化为 JSON
	jsonData, err := json.Marshal(loginForm)
	if err != nil {
		fmt.Printf("JSON序列化失败: %v\n", err)

	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)

	}

	// 设置 Content-Type 头
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)

	}
	defer resp.Body.Close()

	// 处理响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)

	}

	// 解析响应JSON
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
	}

	// 检查响应码
	if loginResp.Result != "0000" {
		return "", fmt.Errorf("登录失败: %s", loginResp.Msg)
	}

	// 更新缓存，token 有效期为 7 天
	s.tokenMutex.Lock()
	s.token = loginResp.Data.AccessToken
	s.tokenExpire = time.Now().Add(7 * 24 * time.Hour)
	s.tokenMutex.Unlock()

	return loginResp.Data.AccessToken, nil
}

func (s *corosService) SportsSummary(labelId, sportType string) (*SportsSummaryResult, error) {
	// 1. 首先获取access token
	token, err := s.Login()
	if err != nil {
		return nil, fmt.Errorf("登录失败: %v", err)
	}

	// 获取配置文件
	cfg, cfgErr := config.LoadDefaultConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	// 2. 构建请求URL
	urlStr := fmt.Sprintf(
		"%s/activity/detail/query?screenW=781&screenH=1440&labelId=%s&sportType=%s",
		cfg.Coros.Address,
		url.QueryEscape(labelId),
		url.QueryEscape(sportType),
	)

	// 3. 创建新的请求
	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 4. 添加认证头
	req.Header.Set("accesstoken", token)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15")

	// 5. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 处理响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应JSON
	var respData struct {
		Result string `json:"result"`
		Msg    string `json:"message"`
		Data   struct {
			LapList []map[string]interface{} `json:"lapList"`
			Summary map[string]interface{}   `json:"summary"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应码
	if respData.Result != "0000" {
		return nil, fmt.Errorf("获取失败: %s", respData.Msg)
	}

	// 筛选出type为2的lapList项，并获取对应的lapItemList
	filteredLapList := make([]map[string]interface{}, 0)
	for _, lap := range respData.Data.LapList {
		if lapType, ok := lap["type"].(float64); ok && int(lapType) == 2 {
			// 获取lapItemList，如果存在的话
			if lapItemList, ok := lap["lapItemList"].([]interface{}); ok {
				// 将[]interface{}转换为[]map[string]interface{}
				for _, item := range lapItemList {
					if lapItem, ok := item.(map[string]interface{}); ok {
						// 添加type字段到每个lapItem中，方便后续处理
						lapItem["lapType"] = lapType
						filteredLapList = append(filteredLapList, lapItem)
					}
				}
			}
		}
	}

	// 返回处理后的数据
	return &SportsSummaryResult{
		LapList: filteredLapList,
		Summary: respData.Data.Summary,
	}, nil
}
