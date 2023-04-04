#!/bin/bash

# 获取最新版本 tag 名称
latest_tag=$(curl --silent "https://api.github.com/repos/csznet/goFile/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# 根据操作系统和处理器架构选择下载的文件名
if [[ "$(uname -s)" == "Linux" ]]; then
  if [[ "$(uname -m)" == "x86_64" ]]; then
    file_name="goFile-${latest_tag}-linux-amd64.tar.gz"
  else
    file_name="goFile-${latest_tag}-linux-arm64.tar.gz"
  fi
elif [[ "$(uname -s)" == "Darwin" ]]; then
  if [[ "$(uname -m)" == "x86_64" ]]; then
    file_name="goFile-${latest_tag}-darwin-amd64.tar.gz"
  else
    file_name="goFile-${latest_tag}-darwin-arm64.tar.gz"
  fi
else
  echo "Unsupported platform: $(uname -s) $(uname -m)"
  exit 1
fi

# 下载文件并解压
url="https://github.com/csznet/goFile/releases/download/${latest_tag}/${file_name}"
curl -L -O $url
tar xf $file_name

# 清理
rm $file_name

# 添加执行权限并移动到 bin 目录
chmod +x goFile
mv goFile /usr/local/bin/

# 提示用户
echo -e "\033[34m在想调用出 goFile 管理的目录下直接执行 goFile 命令即可\033[0m"
