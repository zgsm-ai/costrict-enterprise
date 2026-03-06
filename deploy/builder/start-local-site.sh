#!/bin/sh

#
#   在本地启动一个nginx，构建一个可供下载包的站点
#   设置cloud地址为http://localhost即可通过该站点更新软件
#
service docker start
cp packages.json packages/
cd site && docker compose up -d

