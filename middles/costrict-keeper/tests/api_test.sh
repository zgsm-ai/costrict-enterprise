#!/bin/bash

# 测试获取服务列表API
echo "测试获取服务列表:"
curl -X GET http://localhost:8080/api/services | jq

# 测试重启服务API (需要替换{service_name}为实际服务名)
echo -e "\n测试重启服务:"
curl -X POST http://localhost:8080/api/services/{service_name}/restart | jq

# 测试获取组件列表API  
echo -e "\n测试获取组件列表:"
curl -X GET http://localhost:8080/api/components | jq

# 测试升级组件API (需要替换{component_name}为实际组件名)
echo -e "\n测试升级组件:"
curl -X POST http://localhost:8080/api/components/{component_name}/upgrade | jq

# 测试获取服务端点API
echo -e "\n测试获取服务端点:"
curl -X GET http://localhost:8080/api/endpoints | jq