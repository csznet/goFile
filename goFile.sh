#!/bin/bash

# 获取最新版本的tag
LATEST_TAG=$(curl --silent "https://api.github.com/repos/csznet/goFile/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# 下载最新版本的压缩包
curl -L -O "https://github.com/csznet/goFile/releases/download/${LATEST_TAG}/goFile-${LATEST_TAG}-linux-amd64.tar.gz"

# 解压缩压缩包
tar -zxvf "goFile-${LATEST_TAG}-linux-amd64.tar.gz"

# 执行goFile
./goFile