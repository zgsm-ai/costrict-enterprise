#!/bin/bash

echo "测试获取服务列表API"

# 基本测试
echo "1. 测试正常获取服务列表:"
curl -X GET http://localhost:8080/api/services -v | jq

# 测试带参数
echo "2. 测试带过滤条件的服务列表:"
curl -X GET "http://localhost:8080/api/services?status=running" -v | jq

# 测试错误情况
echo "3. 测试错误路径:"
curl -X GET http://localhost:8080/api/service -v | jq