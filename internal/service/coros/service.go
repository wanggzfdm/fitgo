package coros

import (
	"bytes"
	"coros-fit-mcp/pkg/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LoginResponse 定义登录响应结构体
type LoginResponse struct {
	ApiCode string    `json:"apiCode"`
	Msg     string    `json:"message"`
	Result  string    `json:"result"`
	Data    LoginData `json:"data"`
}

type LoginData struct {
	AccessToken string `json:"accessToken"`
	AccountQueryData
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
	Account     string `json:"account"`
	AccountType int    `json:"accountType"`
	Pwd         string `json:"pwd,omitempty"`
	P1          string `json:"p1,omitempty"`
	P2          string `json:"p2,omitempty"`
}

type corosService struct {
	token       string     // 缓存的 token
	tokenMutex  sync.Mutex // 用于保护 token 的并发访问
	tokenExpire time.Time  // token 过期时间
	loginData   *LoginData
	loginFlowMu sync.Mutex
}

var sharedCorosService = &corosService{}

const defaultRunningModeList = 100
const tokenTTL = 7 * 24 * time.Hour

type tokenCache struct {
	Account     string    `json:"account"`
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type corosAPIStatus struct {
	Result string `json:"result"`
	Msg    string `json:"message"`
}

type corosRequestFailure struct {
	Result string
	Msg    string
}

func (e *corosRequestFailure) Error() string {
	return fmt.Sprintf("高驰接口失败: result=%s message=%s", e.Result, e.Msg)
}

func (s *corosService) ActivityList(size, pageNumber, modeList int) (map[string]interface{}, error) {
	// 获取配置文件
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	if modeList == 0 {
		modeList = defaultRunningModeList
	}

	urlStr := fmt.Sprintf("%s/activity/query?size=%d&pageNumber=%d&modeList=%d",
		cfg.Coros.Address, size, pageNumber, modeList)

	bodyBytes, err := s.doAuthenticatedRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}
	if _, ok := result["data"].(map[string]interface{}); !ok {
		return nil, fmt.Errorf("高驰 activity/query 响应缺少 data 字段: %s", string(bodyBytes))
	}

	return result, nil
}

func NewCorosService() CorosService {
	return sharedCorosService
}

func (s *corosService) Login() (string, error) {
	return s.login(false)
}

func (s *corosService) login(forceRefresh bool) (string, error) {
	s.loginFlowMu.Lock()
	defer s.loginFlowMu.Unlock()

	if forceRefresh {
		s.invalidateTokenCache()
	}

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

	account := cfg.Coros.LoginAccount()
	if account == "" {
		return "", fmt.Errorf("缺少高驰账号，请配置 coros.account 或 coros.username")
	}

	if !forceRefresh {
		if cached, err := loadTokenCache(account); err == nil && cached.AccessToken != "" && time.Now().Before(cached.ExpiresAt) {
			s.tokenMutex.Lock()
			s.token = cached.AccessToken
			s.tokenExpire = cached.ExpiresAt
			s.tokenMutex.Unlock()
			return cached.AccessToken, nil
		}
	}

	accountType := cfg.Coros.LoginAccountType()
	loginUrl := fmt.Sprintf("%s/account/login",
		cfg.Coros.Address)

	loginForm, err := buildLoginForm(cfg.Coros)
	if err != nil {
		return "", err
	}
	loginForm.Account = account
	loginForm.AccountType = accountType

	// 将结构体序列化为 JSON
	jsonData, err := json.Marshal(loginForm)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %w", err)
	}
	logCOROSRequest("POST", loginUrl, jsonData)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 Content-Type 头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Host = req.URL.Host

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 处理响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}
	logCOROSResponse("POST", loginUrl, body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("登录失败: HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应JSON
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应码
	if loginResp.Result != "0000" {
		return "", fmt.Errorf("登录失败: %s", loginResp.Msg)
	}

	// 更新缓存，token 有效期为 7 天
	s.tokenMutex.Lock()
	s.token = loginResp.Data.AccessToken
	s.tokenExpire = time.Now().Add(tokenTTL)
	s.loginData = &loginResp.Data
	s.tokenMutex.Unlock()

	_ = saveTokenCache(tokenCache{
		Account:     account,
		AccessToken: loginResp.Data.AccessToken,
		ExpiresAt:   s.tokenExpire,
	})

	return loginResp.Data.AccessToken, nil
}

func (s *corosService) invalidateTokenCache() {
	s.tokenMutex.Lock()
	s.token = ""
	s.tokenExpire = time.Time{}
	s.loginData = nil
	s.tokenMutex.Unlock()

	path, err := tokenCachePath()
	if err == nil {
		_ = os.Remove(path)
	}
}

func tokenCachePath() (string, error) {
	if override := os.Getenv("COROS_TOKEN_CACHE_PATH"); override != "" {
		return override, nil
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("获取用户缓存目录失败: %w", err)
	}

	return filepath.Join(cacheDir, "coros-fit-mcp", "coros_token.json"), nil
}

func loadTokenCache(account string) (tokenCache, error) {
	path, err := tokenCachePath()
	if err != nil {
		return tokenCache{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return tokenCache{}, err
	}

	var cached tokenCache
	if err := json.Unmarshal(data, &cached); err != nil {
		return tokenCache{}, err
	}

	if cached.Account != account {
		return tokenCache{}, fmt.Errorf("token 缓存账号不匹配")
	}
	if cached.AccessToken == "" {
		return tokenCache{}, fmt.Errorf("token 缓存为空")
	}

	return cached, nil
}

func saveTokenCache(cached tokenCache) error {
	path, err := tokenCachePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func buildLoginForm(cfg config.CorosConfig) (loginform, error) {
	if cfg.P1 != "" || cfg.P2 != "" {
		if cfg.P1 == "" || cfg.P2 == "" {
			return loginform{}, fmt.Errorf("高驰登录配置不完整：p1 和 p2 需要同时提供")
		}
		return loginform{
			P1: cfg.P1,
			P2: cfg.P2,
		}, nil
	}

	if cfg.Password == "" {
		return loginform{}, fmt.Errorf("缺少高驰登录凭证，请配置 p1/p2 或 password")
	}

	return loginform{
		Pwd: cfg.Password,
	}, nil
}

func (s *corosService) AccountQuery() (*AccountQueryData, error) {
	s.tokenMutex.Lock()
	if s.loginData != nil {
		data := s.loginData.AccountQueryData
		s.tokenMutex.Unlock()
		return &data, nil
	}
	s.tokenMutex.Unlock()

	if _, err := s.login(true); err != nil {
		return nil, err
	}

	s.tokenMutex.Lock()
	defer s.tokenMutex.Unlock()
	if s.loginData == nil {
		return nil, fmt.Errorf("登录成功但未返回账号资料")
	}
	data := s.loginData.AccountQueryData
	return &data, nil
}

func (s *corosService) DashboardQuery() (*DashboardQueryData, error) {
	var response DashboardQueryResponse
	if err := s.doAuthenticatedJSONRequest("GET", "/dashboard/query", nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (s *corosService) DashboardDetailQuery() (*DashboardDetailQueryData, error) {
	var response DashboardDetailQueryResponse
	if err := s.doAuthenticatedJSONRequest("GET", "/dashboard/detail/query", nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (s *corosService) SportsSummary(labelId, sportType string) (*SportsSummaryResult, error) {
	// 获取配置文件
	cfg, cfgErr := config.LoadDefaultConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("加载配置失败: %v", cfgErr)
	}

	// 2. 构建请求URL
	urlStr := fmt.Sprintf(
		"%s/activity/detail/query?screenW=781&screenH=1440&labelId=%s&sportType=%s",
		cfg.Coros.Address,
		url.QueryEscape(labelId),
		url.QueryEscape(sportType),
	)

	body, err := s.doAuthenticatedRequest("POST", urlStr, nil)
	if err != nil {
		return nil, err
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

func (s *corosService) doAuthenticatedJSONRequest(method, path string, payload interface{}, out interface{}) error {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	urlStr := cfg.Coros.Address + path
	body, err := s.doAuthenticatedRequest(method, urlStr, payload)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	return nil
}

func (s *corosService) doAuthenticatedRequest(method, urlStr string, payload interface{}) ([]byte, error) {
	type attempt struct {
		forceRefresh bool
		label        string
	}

	attempts := []attempt{
		{forceRefresh: false, label: "cached-token"},
		{forceRefresh: true, label: "refresh-login"},
		{forceRefresh: true, label: "second-refresh-login"},
	}

	var lastResult string
	var lastMsg string
	for idx, attempt := range attempts {
		body, err := s.doAuthenticatedRequestWithToken(method, urlStr, payload, attempt.forceRefresh)
		if err != nil {
			return nil, err
		}

		ok, result, msg := parseCOROSStatus(body)
		if ok {
			if idx > 0 {
				log.Printf("[coros] recovered request method=%s url=%s attempt=%s", method, urlStr, attempt.label)
			}
			return body, nil
		}

		lastResult = result
		lastMsg = msg
		log.Printf("[coros] request failed method=%s url=%s attempt=%s result=%s message=%s", method, urlStr, attempt.label, result, msg)

		if !isTokenInvalid(result, msg) {
			return nil, &corosRequestFailure{Result: result, Msg: msg}
		}

		s.invalidateTokenCache()
	}

	return nil, fmt.Errorf("高驰接口失败: result=%s message=%s", lastResult, lastMsg)
}

func (s *corosService) doAuthenticatedRequestWithToken(method, urlStr string, payload interface{}, forceRefresh bool) ([]byte, error) {
	token, err := s.login(forceRefresh)
	if err != nil {
		return nil, err
	}
	return doCOROSJSONRequest(method, urlStr, payload, token)
}

func doCOROSJSONRequest(method, urlStr string, payload interface{}, token string) ([]byte, error) {
	var requestBody io.Reader
	var requestBytes []byte
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("JSON序列化失败: %w", err)
		}
		requestBytes = data
		requestBody = bytes.NewBuffer(data)
	}

	logCOROSRequest(method, urlStr, requestBytes)
	return doCOROSRequest(method, urlStr, requestBody, token)
}

func doCOROSRequest(method, urlStr string, body io.Reader, token string) ([]byte, error) {
	if body == nil {
		logCOROSRequest(method, urlStr, nil)
	}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("accesstoken", token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	logCOROSResponse(method, urlStr, responseBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("请求失败: HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

func logCOROSRequest(method, urlStr string, body []byte) {
	if len(body) == 0 {
		log.Printf("[coros] request method=%s url=%s body=<empty>", method, urlStr)
		return
	}
	log.Printf("[coros] request method=%s url=%s body=%s", method, urlStr, string(body))
}

func logCOROSResponse(method, urlStr string, body []byte) {
	if len(body) == 0 {
		log.Printf("[coros] response method=%s url=%s body=<empty>", method, urlStr)
		return
	}
	log.Printf("[coros] response method=%s url=%s body=%s", method, urlStr, string(body))
}

func parseCOROSStatus(body []byte) (bool, string, string) {
	var status corosAPIStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return true, "", ""
	}
	if status.Result == "" || status.Result == "0000" {
		return true, status.Result, status.Msg
	}
	return false, status.Result, status.Msg
}

func isTokenInvalid(result, msg string) bool {
	if result == "1019" {
		return true
	}
	return msg == "Access token is invalid"
}
