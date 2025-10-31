# FitGo Project

## 目录结构

```
.
├── go.mod
├── go.sum
├── cmd/
│   ├── app/               # 后端服务入口
│   │   └── main.go
│   └── web/               # 前端项目目录
│       └── fitgo-web/     # Vue.js 前端项目
│           ├── public/    # 静态资源
│           ├── src/       # 源代码
│           │   ├── assets/     # 资源文件
│           │   ├── components/ # Vue 组件
│           │   ├── router/     # 路由配置
│           │   ├── App.vue     # 根组件
│           │   └── main.js     # 入口文件
│           └── package.json    # 前端依赖
├── configs/
│   └── config.json        # 配置文件
├── internal/
│   ├── handler/           # HTTP 处理器
│   │   ├── tcx.go
│   │   └── coros.go
│   └── service/           # 业务逻辑
│       ├── tcx/           # TCX 相关服务
│       └── coros/         # COROS 相关服务
├── pkg/
│   └── config/            # 配置处理
│       └── config.go
├── router/                # 路由定义
│   └── router.go
└── tests/                 # 测试文件
```

## 目录说明

- **go.mod**: Go 模块定义文件
- **main.go**: 程序入口文件（示例）
- **cmd/**: 应用程序入口点
  - `app/main.go`: Web 服务入口点
- **configs/**: 配置文件
  - `config.json`: 应用程序配置文件
- **internal/**: 私有应用程序和库代码
  - `handler/tcx.go`: TCX HTTP 请求处理器
  - `service/tcx/`: TCX 业务逻辑服务
    - `api.go`: TCX 接口定义和数据结构
    - `service.go`: TCX 服务实现
- **pkg/**: 可供外部使用的库代码
  - `config/config.go`: 配置处理包
- **router/**: 路由配置
- **scripts/**: 脚本文件
- **tests/**: 测试文件

## 功能概述

FitGo 是一个综合性的运动数据分析平台，提供 TCX 文件处理和 COROS 运动数据同步功能，并配有现代化的 Web 界面。

### 主要功能

#### 1. TCX 文件处理
- **TCXService**: 核心服务接口，定义了处理 TCX 文件的方法
- **UploadTCX**: 上传并处理 TCX 文件
- **GetTCXSummary**: 获取特定 TCX 文件摘要
- **ListTCXSummaries**: 列出所有 TCX 文件摘要

#### 2. COROS 运动数据同步
- 同步 COROS 运动记录
- 获取详细运动数据
- AI 运动数据分析

#### 3. Web 界面
- 响应式设计，适配各种设备
- 运动记录列表展示
- 详细数据分析报告
- 交互式数据可视化

### 数据结构

- **TCXSummary**: 表示 TCX 文件的摘要信息，包括运动时间、距离、卡路里等

## 快速开始

### 1. 后端服务

#### 配置

编辑 `configs/config.json` 配置文件：

```json
{
  "server": {
    "port": ":9092",
    "host": "localhost"
  },
  "app": {
    "name": "FitGo",
    "version": "1.0.0"
  },
  "cors": {
    "allowed_origins": ["http://localhost:9093"],
    "allowed_methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    "allowed_headers": ["Content-Type", "Authorization"]
  }
}
```

#### 启动后端服务

```bash
# 进入项目目录
cd /Library/MyFile/go/fitgo

# 启动后端服务
go run cmd/app/main.go
```

### 2. 前端开发

#### 环境要求
- Node.js 14+
- npm 或 yarn

#### 启动开发服务器

```bash
# 进入前端项目目录
cd cmd/web/fitgo-web

# 安装依赖
npm install

# 启动开发服务器
npm run serve
```

开发服务器将在 http://localhost:9093 运行，并自动代理 API 请求到后端服务。

## API 文档

### 运动记录

#### 获取运动记录列表

```
GET /coros/active?size=10&pageNumber=1
```

**参数:**
- `size`: 每页记录数
- `pageNumber`: 页码

#### 获取运动详情

```
GET /coros/sports/summary?labelId={labelId}&sportType={sportType}
```

#### 获取 AI 分析报告

```
GET /coros/ai/summary?labelId={labelId}&sportType={sportType}
```

## 开发指南

### 前端开发

前端项目使用 Vue 3 和 Naive UI 构建，主要特性：

- 使用 Composition API
- 响应式布局
- 现代化的 UI 组件
- 支持 Markdown 渲染

### 后端开发

后端使用 Go 语言开发，主要特性：

- 模块化设计
- 中间件支持（CORS、日志等）
- 清晰的目录结构
- 配置管理

## 部署

### 构建前端

```bash
cd cmd/web/fitgo-web
npm run build
```

构建后的文件将生成在 `dist` 目录下。

### 部署后端

```bash
# 构建可执行文件
go build -o fitgo cmd/app/main.go

# 运行
./fitgo
```

## 贡献指南

欢迎提交 Issue 和 Pull Request。在提交代码前，请确保：

1. 代码符合 Go 代码规范
2. 添加必要的测试
3. 更新相关文档

## 许可证

MIT

## 通用配置加载

项目提供了一个通用的配置加载包 `pkg/config`，包含以下功能：

- `LoadConfig(filepath string)`: 从指定路径加载配置
- `LoadConfigWithDefaults(primaryPath, fallbackPath string)`: 从主路径或备选路径加载配置
- `LoadDefaultConfig()`: 使用默认路径加载配置（推荐使用）

### 使用示例

```go
// 简单使用默认配置加载
cfg, err := config.LoadDefaultConfig()
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// 使用配置
port := cfg.Server.Port
```

## API 服务

项目包含一个基于 HTTP 的 API 服务，可以通过以下端点访问：

- `GET /` - API 根路径，返回服务信息
- `POST /upload` - 上传 TCX 文件
- `GET /summary/{id}` - 获取特定 TCX 文件摘要
- `GET /summaries` - 列出所有 TCX 文件摘要

### 启动 API 服务

```bash
cd cmd/app
go run main.go
```

服务将根据配置文件中的设置启动。

### 使用 API

#### 上传 TCX 文件

```bash
curl -X POST -F "file=@example.tcx" http://localhost:8080/upload
```

## 项目入口文件说明

项目包含两个不同的入口文件：

1. **根目录下的 main.go**: 这是一个简单的示例程序，展示了如何使用 TCX 服务的基本用法。

2. **cmd/app/main.go**: 这是完整的 Web 服务应用程序，提供 HTTP API 接口用于处理 TCX 文件。

根据你的需求，可以选择运行其中任何一个入口文件：
- 运行示例程序: `go run main.go`
- 运行 Web 服务: `cd cmd/app && go run main.go`

这是一个标准的 Go 项目布局，为构建结构良好的应用程序提供了良好基础。