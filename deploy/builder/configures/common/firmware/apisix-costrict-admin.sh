#!/bin/sh

. ./configure.sh

curl -i http://$APISIX_ADDR/apisix/admin/upstreams -H "$AUTH" -H "$TYPE" -X PUT -d '{
    "id": "costrict-admin-frontend",
    "nodes": {
      "portal:80": 1
    },
    "type": "roundrobin"
  }'

curl -i http://$APISIX_ADDR/apisix/admin/upstreams -H "$AUTH" -H "$TYPE" -X PUT -d '{
    "id": "costrict-admin-backend",
    "nodes": {
      "costrict-admin-backend:8080": 1
    },
    "type": "roundrobin"
  }'

curl -i http://$APISIX_ADDR/apisix/admin/routes -H "$AUTH" -H "$TYPE" -X PUT -d '{
    "uri": "/costrict-admin/api/*",
    "id": "costrict-admin-backend",
    "name": "costrict-admin-backend",
    "upstream_id": "costrict-admin-backend",
    "priority": 100,
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

curl -i http://$APISIX_ADDR/apisix/admin/routes -H "$AUTH" -H "$TYPE" -X PUT -d '{
    "uri": "/costrict-admin/*",
    "id": "costrict-admin-frontend",
    "name": "costrict-admin-frontend",
    "upstream_id": "costrict-admin-frontend",
    "priority": 10,
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
