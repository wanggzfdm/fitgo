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

### `analyze_training_status`

参数：

```json
{
  "date": "2026-03-25"
}
```

`date` 可选。用于分析当前或指定日期的训练状态，并输出：

- 当前恢复、疲劳、负荷平衡、训练准备度
- 焦点训练属于恢复跑 / 有氧跑 / 阈值跑 / 间歇跑哪一类
- 明天训练建议
- 后续 3 天训练计划

不传 `date` 时默认分析当前状态和最新活动。

### `get_training_profile`

无参数。聚合返回适合分析长期训练能力的资料，包括：

- 跑者基础资料
- 训练分区
- 训练看板

如果部分接口不可用，结果里会保留已获取的数据，并在 `errors` 中说明失败项。

### `get_training_context`

参数：

```json
{
  "recent_limit": 5,
  "days": 7
}
```

聚合返回适合分析近期训练状态的数据，包括：

- 训练负荷状态
- 最近活动列表
- 本周训练汇总
- 多日训练趋势

`recent_limit` 和 `days` 都可选；如果部分接口不可用，结果里会保留已获取的数据，并在 `errors` 中说明失败项。

### `get_latest_coros_activity_summary`

无参数。读取高驰账号最新一条活动，并返回统一摘要 JSON，结果里同时带结构化字段和一段 Markdown 摘要。

### `get_coros_daily_running_summaries`

参数：

```json
{
  "date": "2026-03-26"
}
```

`date` 可选，格式为 `YYYY-MM-DD`；未传时默认使用北京时间当天。仅筛选户外跑步 `100` 和运动场跑步 `103`，返回当日跑步汇总和每次跑步摘要。

### `get_coros_daily_trail_running_summaries`

参数：

```json
{
  "date": "2026-03-26"
}
```

`date` 可选，格式为 `YYYY-MM-DD`；未传时默认使用北京时间当天。仅筛选越野跑 `102`，返回当日越野跑汇总和每次越野跑摘要。

示例输出：

```json
{
  "date": "2026-03-28",
  "timezone": "Asia/Shanghai",
  "activity_count": 1,
  "daily_summary": {
    "summary": {
      "source": "coros",
      "source_name": "coros_daily_trail_running_summaries",
      "name": "2026-03-28 越野跑汇总",
      "sport_type": "越野跑",
      "start_time": "2026-03-28T10:45:06+08:00",
      "end_time": "2026-03-28T15:33:44+08:00",
      "duration_seconds": 17318,
      "moving_seconds": 17317,
      "distance_meters": 28084.6,
      "ascent_meters": 2037,
      "descent_meters": 2013,
      "elevation_gain_per_km": 72.5,
      "vertical_ascent_per_hour": 423.5,
      "time_per_100m_ascent_seconds": 850.1,
      "moving_ratio": 1,
      "average_heart_rate": 160,
      "max_heart_rate": 178,
      "average_power": 170,
      "average_moving_pace_sec_per_km": 616.6,
      "training_load": 896,
      "highlights": [
        "完成 28.08 km，移动时间 4:48:37，累计爬升 2037 m。",
        "爬升效率 424 m/h。",
        "单位距离爬升 72 m/km。",
        "平均心率 160 bpm，最高 178 bpm。",
        "训练负荷 896。"
      ]
    },
    "markdown": "# 2026-03-28 越野跑汇总\n- 类型：越野跑\n- 开始时间：2026-03-28T10:45:06+08:00\n\n📍 越野概况\n- 距离：28.08 km\n- 总时间：4:48:38\n- 移动时间：4:48:37\n- 累计爬升：2037 m\n- 累计下降：2013 m\n- 单位距离爬升：72 m/km\n\n⛰️ 地形与效率\n- 爬升效率：424 m/h\n- 每爬升 100m 用时：850 s\n- 移动占比：100%\n- 平均移动配速：10:17 /km\n- 最快配速：5:26 /km\n\n❤️ 强度与负荷\n- 平均心率：160 bpm\n- 最大心率：178 bpm\n- 平均功率：170 W\n- 训练负荷：896\n- 消耗热量：3166 kcal\n\n👣 动作数据\n- 步频：136 spm\n- 步幅：0.74 m\n- 总步数：38172"
  },
  "activities": [
    {
      "summary": {
        "source": "coros",
        "source_name": "coros_latest_activity",
        "activity_id": "476387922021482797",
        "name": "福州市 越野跑",
        "sport_type": "越野跑",
        "start_time": "2026-03-28T10:45:06+08:00",
        "end_time": "2026-03-28T15:33:44+08:00",
        "distance_meters": 28084.6,
        "ascent_meters": 2037,
        "descent_meters": 2013,
        "elevation_gain_per_km": 72.5,
        "vertical_ascent_per_hour": 423.5,
        "average_heart_rate": 160,
        "max_heart_rate": 178,
        "average_moving_pace_sec_per_km": 616.6,
        "training_load": 896
      },
      "markdown": "# 福州市 越野跑\n- 类型：越野跑\n- 开始时间：2026-03-28T10:45:06+08:00\n\n📍 越野概况\n- 距离：28.08 km\n- 总时间：4:48:38\n- 移动时间：4:48:37\n- 累计爬升：2037 m\n- 累计下降：2013 m\n- 单位距离爬升：72 m/km\n\n⛰️ 地形与效率\n- 爬升效率：424 m/h\n- 每爬升 100m 用时：850 s\n- 移动占比：100%\n- 平均移动配速：10:17 /km\n- 最快配速：5:26 /km\n\n❤️ 强度与负荷\n- 平均心率：160 bpm\n- 最大心率：178 bpm\n- 平均功率：170 W\n- 训练负荷：896\n- 消耗热量：3166 kcal\n\n👣 动作数据\n- 步频：136 spm\n- 步幅：0.74 m\n- 总步数：38172"
    }
  ]
}
```

这类摘要会优先突出越野跑真正重要的信息：总时间、移动时间、累计爬升、爬升密度、爬升效率、移动占比，以及强度与负荷，而不是沿用公路跑那套以平均配速为中心的展示方式。

### `summarize_fit_file`

参数：

```json
{
  "path": "/absolute/or/relative/path/to/file.fit"
}
```

读取本地 `.fit` 文件并返回统一摘要 JSON。

## 返回格式

新的聚合分析工具统一返回：

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

- 分析训练状态：`analyze_training_status`，参数 `{"date": "2026-03-25"}`
- 获取训练档案：`get_training_profile`
- 获取近期训练状态：`get_training_context`，参数 `{"recent_limit": 5, "days": 7}`

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

- `analyze_training_status`

```json
{
  "date": "2026-03-25"
}
```

适用于：

- `analyze_training_status`

```json
{
  "recent_limit": 5,
  "days": 7
}
```

适用于：

- `get_training_context`

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
