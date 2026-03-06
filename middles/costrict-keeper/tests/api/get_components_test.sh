#!/bin/bash

echo "测试获取组件列表API"

# 测试正常获取组件列表
echo "1. 测试获取所有组件:"
curl -X GET http://localhost:8080/api/components -v | jq

# 测试带过滤条件
echo "2. 测试按类型过滤组件:"
curl -X GET "http://localhost:8080/api/components?type=database" -v | jq

# 测试分页参数
echo "3. 测试带分页参数:"
curl -X GET "http://localhost:8080/api/components?page=1&limit=5" -v | jq