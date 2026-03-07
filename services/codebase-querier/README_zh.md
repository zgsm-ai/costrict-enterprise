# codebase-querier

<div align="center">

[English](./README.md) | [简体中文](./README_zh.md)

连接客户端codebase-querier的接口代理

[![Go Report Card](https://goreportcard.com/badge/github.com/zgsm-ai/codebase-querier)](https://goreportcard.com/report/github.com/zgsm-ai/codebase-querier)
[![Go Reference](https://pkg.go.dev/badge/github.com/zgsm-ai/codebase-querier.svg)](https://pkg.go.dev/github.com/zgsm-ai/codebase-querier)
[![License](https://img.shields.io/github/license/zgsm-ai/codebase-querier)](LICENSE)

</div>

## 项目概述

codebase-querier 是诸葛神码 AI 编程助手的服务端模块之一，连接到客户端codebase-querier 提供代码调用链图关系检索。

## 环境要求

- Go 1.24.3 或更高版本
- Docker

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/zgsm-ai/codebase-querier.git
cd codebase-querier

# 安装依赖
go mod tidy
```

### 配置

```bash
vim etc/config.yaml
```


### 运行

```bash
# 构建项目
make build

```

## 许可证

本项目采用 [Apache 2.0 许可证](LICENSE)。