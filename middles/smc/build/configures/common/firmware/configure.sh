#!/bin/sh

source costrict.env

#---------------------------------------------------------
# apisix设置，无需修改
#---------------------------------------------------------

APISIX_ADDR="127.0.0.1:${PORT_APISIX_API}"
AUTH="X-API-KEY: ${APIKEY_APISIX_ADMIN}"
TYPE="Content-Type: application/json"

#-------------------------------------------------------------------------------
#   以下设置请根据部署环境信息进行修改
#-------------------------------------------------------------------------------
# VSCODE扩展连接诸葛神码后端时使用的入口URL地址
# 一般会利用DNS及应用发布设备将该地址映射到 http://${COSTRICT_HOST}:${PORT_APISIX_ENTRY}
SERVER_IP=$(hostname -I | awk '{ print $1 }')
declare -r SERVER_IP
# 判断 COSTRICT_HOST 是否已定义，若未定义则设一个默认值
if [ -z "${COSTRICT_HOST:-}" ]; then
    COSTRICT_HOST="${SERVER_IP}"
fi

# 判断 COSTRICT_BASEURL 是否已定义，若未定义则设一个默认值
if [ -z "${COSTRICT_BASEURL:-}" ]; then
    COSTRICT_BASEURL="http://${COSTRICT_HOST}:${PORT_APISIX_ENTRY}"
fi
#---------------------------------------------------------
# 大模型相关设置，请根据实际部署情况设置
#---------------------------------------------------------
# 模型服务器的IP，需要根据实际情况设置
MODEL_SERVER_ADDR="127.0.0.1:${PORT_ONEAPI}"
CHAT_DEFAULT_MODEL="GLM-4.5-FP8"

# 代码补全模型的BASEURL,MODEL,APIKEY
COMPLETION_BASEURL="http://${MODEL_SERVER_ADDR}/v1/completions"
COMPLETION_MODEL="DeepSeek-Coder-V2-Lite"
COMPLETION_APIKEY=""

# 向量嵌入模型的BASEURL,MODEL和APIKEY
EMBEDDER_BASEURL="http://${MODEL_SERVER_ADDR}/v1/embeddings"
EMBEDDER_MODEL="embedding"
EMBEDDER_APIKEY=""

# RAG排序模型的BASEURL,MODEL和APIKEY
RERANKER_BASEURL="http://${MODEL_SERVER_ADDR}/v1/rerank"
RERANKER_MODEL="rerank"
RERANKER_APIKEY=""
