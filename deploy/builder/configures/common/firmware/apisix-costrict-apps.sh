#!/bin/sh

. ./configure.sh

curl -i http://$APISIX_ADDR/apisix/admin/upstreams -H "$AUTH" -H "$TYPE" -X PUT -d '{
    "id": "costrict-apps",
    "nodes": {
      "portal:80": 1
    },
    "type": "roundrobin"
  }'

curl -i http://$APISIX_ADDR/apisix/admin/routes -H "$AUTH" -H "$TYPE" -X PUT -d '{
    "uri": "/shenma/*",
    "id": "shenma-client",
    "name": "shenma-client",
    "upstream_id": "costrict-apps",
    "plugins": {
      "proxy-rewrite": {
        "regex_uri": ["^/shenma/api/v1(.*)", "/shenma-cli-tools$1"]
      },
      "limit-count": {
        "count": 100,
        "time_window": 60,
        "rejected_code": 429,
        "key_type": "var_combination",
        "key": "$remote_addr $http_x_forwarded_for"
      }
    }
  }'

curl -i http://$APISIX_ADDR/apisix/admin/routes -H "$AUTH" -H "$TYPE" -X PUT -d '{
    "uri": "/costrict/*",
    "id": "costrict-apps",
    "name": "costrict-apps",
    "upstream_id": "costrict-apps",
    "plugins": {
      "limit-count": {
        "count": 300,
        "time_window": 60,
        "rejected_code": 429,
        "key_type": "var_combination",
        "key": "$remote_addr $http_x_forwarded_for"
      }
    }
  }'
