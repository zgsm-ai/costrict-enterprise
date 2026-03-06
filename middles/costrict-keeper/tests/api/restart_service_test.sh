#!/bin/bash

echo "测试重启服务API"

# 测试正常重启服务
echo "1. 测试正常重启服务(替换example-service为实际服务名):"
curl -X POST http://localhost:8080/api/services/example-service/restart -v | jq

# 测试不存在的服务
echo "2. 测试不存在的服务:"
curl -X POST http://localhost:8080/api/services/non-existent-service/restart -v | jq

# 测试无效请求方法
echo "3. 测试GET方法请求(应该失败):"
curl -X GET http://localhost:8080/api/services/example-service/restart -v | jq