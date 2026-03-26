# coros-fit-mcp

一个独立的 MCP 服务，当前聚焦两类能力：

- 从高驰账号读取运动摘要与个人训练指标
- 读取本地 `.fit` 文件，并转换成统一运动摘要

## 当前结构

```text
.
├── cmd/mcp/                          # MCP stdio 服务入口
├── configs/config.json               # 高驰账号配置
├── internal/service/coros/           # 高驰 API 客户端
├── internal/service/activitysummary/ # 摘要模型、格式化、COROS/FIT 转换
├── internal/service/personalmetrics/ # 个人指标提取与格式化
└── pkg/config/                       # 配置加载
```

## 配置

`configs/config.json` 只需要保留高驰配置。
仓库里另外提供了一个脱敏样例 `configs/config.example.json`：

```json
{
  "coros": {
    "account": "you@example.com",
    "accountType": 2,
    "p1": "$2b$10$exampleexampleexampleexampleexampleexampleexample",
    "p2": "$2b$10$exampleexampleexample",
    "address": "https://teamapi.coros.com"
  }
}
```

当前登录优先使用新版字段：

- `account`
- `accountType`
- `p1`
- `p2`

如果你手里还是旧凭证，也兼容：

- `username`
- `password`

## 运行

开发时直接启动 stdio MCP 服务：

```bash
go run ./cmd/mcp
```

如果要构建：

```bash
go build ./cmd/mcp
```

高驰登录 token 会优先走进程内缓存，并落到本机用户缓存目录，避免每次启动都重新登录。进程内 token 过期后，会先尝试读取本地缓存；本地缓存也过期了，才重新登录。

远程调试可以启动 HTTP 入口，同时支持 SSE 和 Streamable HTTP：

```bash
go run ./cmd/mcp-sse
```

默认监听 `:9093`，默认基址 `http://127.0.0.1:9093`。
可通过环境变量覆盖：

```bash
MCP_SSE_ADDR=:9090 MCP_BASE_URL=http://127.0.0.1:9090 go run ./cmd/mcp-sse
```

启动后可用的远程端点是：

- `GET /sse`
- `POST /message?sessionId=...`
- `POST /mcp`
- `GET /healthz`

其中：

- 旧版 SSE 客户端继续用 `http://host:port/sse`
- 支持 Streamable HTTP 的客户端用 `http://host:port/mcp`

## MCP Tools

### `get_runner_profile`

无参数。返回高驰个人基础训练资料，包括：

- 身高、体重、生日、性别、国家
- 最大心率、静息心率
- 乳酸阈心率、乳酸阈配速

### `get_training_zones`

无参数。返回高驰当前训练分区，包括：

- 心率区间
- 配速区间
- 配速格式化文本，例如 `5'07"/km`

### `get_training_dashboard`

无参数。返回高驰训练看板核心指标，包括：

- 跑步能力与分项能力
- 恢复状态
- 睡眠 HRV
- 个人纪录

### `get_training_load_status`

无参数。返回训练负荷状态，包括：

- 短期负荷 `ATI`
- 长期负荷 `CTI`
- 负荷比与百分比
- 疲劳状态
- 未来几天推荐训练负荷

### `get_recent_activities`

参数：

```json
{
  "limit": 5
}
```

`limit` 可选。返回最近运动列表，包括日期、距离、时长、平均配速、平均心率、平均功率、训练负荷等字段。

### `get_weekly_summary`

无参数。返回本周训练汇总，包括：

- 总距离
- 总时长
- 总训练负荷

### `get_training_trends`

参数：

```json
{
  "days": 7
}
```

`days` 可选。返回最近多日训练趋势，包括训练负荷、ATI、CTI、VO2 Max、跑步能力、阈值配速等变化。

### `get_latest_coros_activity_summary`

无参数。读取高驰账号最新一条活动，并返回统一摘要 JSON，结果里同时带结构化字段和一段 Markdown 摘要。

### `get_coros_daily_running_summaries`

参数：

```json
{
  "date": "2026-03-26"
}
```

`date` 可选，格式为 `YYYY-MM-DD`；未传时默认使用北京时间当天。返回当日跑步汇总和每次跑步摘要。

### `summarize_fit_file`

参数：

```json
{
  "path": "/absolute/or/relative/path/to/file.fit"
}
```

读取本地 `.fit` 文件并返回统一摘要 JSON。

## 返回格式

新的个人指标工具统一返回：

- `data`：结构化 JSON 数据
- `markdown`：简短 Markdown 文本，方便 MCP 客户端直接展示

出于安全考虑，不会返回以下敏感字段：

- `accessToken`
- `email`
- `userId`
- `nickname`

## MCP 客户端配置示例

### Claude Desktop

如果你希望 Claude Desktop 以 stdio 方式拉起本服务，可以在 Claude Desktop 的 MCP 配置里加入：

把下面的 `/path/to/coros-fit-mcp` 替换成你的本地仓库路径。

```json
{
  "mcpServers": {
    "coros-fit-mcp": {
      "command": "go",
      "args": [
        "run",
        "/path/to/coros-fit-mcp/cmd/mcp"
      ]
    }
  }
}
```

如果你已经构建过二进制，也可以改成：

```json
{
  "mcpServers": {
    "coros-fit-mcp": {
      "command": "/path/to/coros-fit-mcp/mcp"
    }
  }
}
```

常见调用示例：

- 获取跑者资料：`get_runner_profile`
- 获取训练区间：`get_training_zones`
- 获取训练看板：`get_training_dashboard`
- 获取训练负荷：`get_training_load_status`
- 获取最近活动：`get_recent_activities`，参数 `{"limit": 5}`
- 获取本周汇总：`get_weekly_summary`
- 获取训练趋势：`get_training_trends`，参数 `{"days": 7}`

### Cherry Studio

Cherry Studio 可以直接接 stdio 或远程 HTTP。

#### 方式一：stdio

可执行命令填写：

```bash
go run /path/to/coros-fit-mcp/cmd/mcp
```

如果使用已构建的二进制：

```bash
/path/to/coros-fit-mcp/mcp
```

#### 方式二：远程 HTTP / SSE

先启动服务：

```bash
go run ./cmd/mcp-sse
```

然后在 Cherry Studio 中按客户端支持方式填写：

- SSE 地址：`http://127.0.0.1:9093/sse`
- Streamable HTTP 地址：`http://127.0.0.1:9093/mcp`

常见调用示例：

```json
{}
```

适用于：

- `get_runner_profile`
- `get_training_zones`
- `get_training_dashboard`
- `get_training_load_status`
- `get_weekly_summary`

```json
{
  "limit": 5
}
```

适用于：

- `get_recent_activities`

```json
{
  "days": 7
}
```

适用于：

- `get_training_trends`

```json
{
  "date": "2026-03-26"
}
```

适用于：

- `get_coros_daily_running_summaries`

```json
{
  "path": "/absolute/or/relative/path/to/file.fit"
}
```

适用于：

- `summarize_fit_file`

## 验证

```bash
go test ./...
go build ./cmd/mcp
go build ./cmd/mcp-sse
```
