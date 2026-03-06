#!/bin/bash

echo "测试升级组件API"

# 测试正常升级组件
echo "1. 测试升级组件(替换example-component为实际组件名):"
curl -X POST http://localhost:8080/api/components/example-component/upgrade -v | jq

# 测试不存在的组件
echo "2. 测试不存在的组件升级:"
curl -X POST http://localhost:8080/api/components/non-existent/upgrade -v | jq

# 测试带版本参数
echo "3. 测试指定版本升级:"
curl -X POST "http://localhost:8080/api/components/example-component/upgrade?version=1.2.0" -v | jq