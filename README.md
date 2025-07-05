# scripts-llm

这是大三小学习后端代码项目，用于使用 LLM 解析剧本，实现与剧本中的人物对话。

## API 文档
- [API 文档](https://kwtyo10mo4.apifox.cn/)

## 项目结构
```
├── .gitignore
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
├── api/
│   └── helloworld/
│       └── v1/
├── cmd/
│   └── scripts-llm/
│       ├── main.go
│       ├── wire.go
│       └── wire_gen.go
├── configs/
│   └── config.yaml
├── go.mod
├── go.sum
├── internal/
│   ├── biz/
│   │   ├── README.md
│   │   ├── biz.go
│   │   └── greeter.go
│   ├── conf/
│   │   ├── conf.pb.go
│   │   └── conf.proto
│   ├── data/
│   │   ├── README.md
│   │   ├── data.go
│   │   └── greeter.go
│   ├── server/
│   │   ├── grpc.go
│   │   ├── http.go
│   │   └── server.go
│   └── service/
│       ├── README.md
│       ├── greeter.go
│       └── service.go
├── openapi.yaml
└── third_party/
    ├── README.md
    ├── errors/
    │   └── errors.proto
    ├── google/
    │   ├── api/
    │   └── protobuf/
    ├── openapi/
    │   └── v3/
    └── validate/
        ├── README.md
        └── validate.proto
```

## 运行步骤
1. 安装依赖：`go mod tidy`
2. 运行项目：`go run cmd/scripts-llm/main.go`