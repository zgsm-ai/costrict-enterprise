#!/bin/sh

. ./configure.sh

curl -i http://$APISIX_ADDR/apisix/admin/upstreams -H "$AUTH" -H "$TYPE" -X PUT  -d '{
    "id": "client-manager",
    "nodes": {
      "client-manager.costrict.svc.cluster.local:8080": 1
    },
    "type": "roundrobin"
  }'

curl -i  http://$APISIX_ADDR/apisix/admin/routes -H "$AUTH" -H "$TYPE" -X PUT -d '{
    "id": "client-manager",
    "name": "client-manager",
    "uris": ["/client-manager/api/v1/*"],
    "upstream_id": "client-manager",
    "plugins": {
      "limit-count": {
        "allow_degradation": false,
        "count": 10000,
        "key_type": "var_combination",
        "key": "$remote_addr $http_x_forwarded_for",
        "policy": "local",
        "rejected_code": 429,
        "show_limit_quota_header": true,
        "time_window": 86400
      },
      "limit-req": {
        "allow_degradation": false,
        "burst": 30,
        "key": "$remote_addr $http_x_forwarded_for",
        "key_type": "var_combination",
        "nodelay": false,
        "policy": "local",
        "rate": 30,
        "rejected_code": 429
      },
      "openid-connect": {
        "client_id": "'"$OIDC_CLIENT_ID"'",
        "client_secret": "'"$OIDC_CLIENT_SECRET"'",
        "discovery": "'"$OIDC_DISCOVERY_ADDR"'",
        "introspection_endpoint": "'"$OIDC_INTROSPECTION_ENDPOINT"'",
        "introspection_endpoint_auth_method": "client_secret_basic",
        "introspection_interval": 60,
        "bearer_only": true,
        "scope": "openid profile email"
      }
    }
  }'

