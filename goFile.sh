#!/bin/bash

# 检查是否以sudo权限运行
if [ "$EUID" -ne 0 ]; then
  echo -e "\033[31m请使用sudo权限运行此脚本\033[0m"
  exit 1
fi

# 检查是否安装了 gzip
if ! command -v gzip &> /dev/null; then
  echo "gzip is not installed. Installing gzip..."
  if [[ "$(uname -s)" == "Linux" ]]; then
    if [[ -f /etc/redhat-release ]]; then
      yum install -y gzip
    elif [[ -f /etc/debian_version ]]; then
      apt-get update
      apt-get install -y gzip
    else
      echo "Unsupported Linux distribution"
      exit 1
    fi
  elif [[ "$(uname -s)" == "Darwin" ]]; then
    brew install gzip
  else
    echo "Unsupported platform: $(uname -s)"
    exit 1
  fi
fi

# 根据操作系统和处理器架构选择下载的文件名
if [[ "$(uname -s)" == "Linux" ]]; then
  if [[ "$(uname -m)" == "x86_64" ]]; then
    file_name="goFile-linux-amd64.tar.gz"
  else
    file_name="goFile-linux-arm64.tar.gz"
  fi
elif [[ "$(uname -s)" == "Darwin" ]]; then
  if [[ "$(uname -m)" == "x86_64" ]]; then
    file_name="goFile-darwin-amd64.tar.gz"
  else
    file_name="goFile-darwin-arm64.tar.gz"
  fi
else
  echo "Unsupported platform: $(uname -s) $(uname -m)"
  exit 1
fi

# 让用户选择下载方式
echo "请选择下载方式："
echo "1) 从 GitHub 下载"
echo "2) 使用镜像地址下载"
echo "3) 自动判断"
read -p "请输入选项 (1/2/3): " choice

case $choice in
  1)
    echo "选择了 GitHub 下载"
    url="https://github.com/csznet/goFile/releases/latest/download/${file_name}"
    ;;
  2)
    echo "选择了镜像下载"
    url="https://ghfast.top/https://github.com/csznet/goFile/releases/latest/download/${file_name}"
    ;;
  3)
    echo "正在自动判断网络环境..."
    # 获取百度的平均延迟（ping 5次并取平均值）
    ping_result=$(ping -c 5 -q baidu.com | awk -F'/' 'END{print $5}')

    # 判断延迟是否在100以内
    if awk -v ping="$ping_result" 'BEGIN{exit !(ping < 100)}'; then
      echo "服务器位于中国国内，使用镜像下载"
      url="https://ghfast.top/https://github.com/csznet/goFile/releases/latest/download/${file_name}"
    else
      echo "服务器位于国外，使用 GitHub 下载"
      url="https://github.com/csznet/goFile/releases/latest/download/${file_name}"
    fi
    ;;
  *)
    echo -e "\033[31m无效选项，请重新运行脚本并选择 1/2/3\033[0m"
    exit 1
    ;;
esac

# 下载文件并解压
curl -L -O $url
tar xf $file_name

# 清理
rm $file_name

# 添加执行权限并移动到 bin 目录
chmod +x goFile
mv goFile /usr/local/bin/

# 提示用户
echo -e "\033[34m在想调用出 goFile 管理的目录下直接执行 goFile 命令即可\033[0m"
