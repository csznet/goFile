# goFile
easy file manager

<img width="1029" alt="image" src="https://user-images.githubusercontent.com/127601663/225728027-fdfe5172-1220-4619-8635-60bb4a085c89.png">

介绍
===

为了方便使用Caddy使用写的小东西
简单的网页文件管理器，可以指定目录，指定端口，即用即开
目前实现的功能：
 - 后台远程下载
 - 上传文件
 - 删除文件
 - 解压ZIP文件

下载
===
    wget https://github.com/csznet/goFile/releases/download/v1.0.4/goFile-v1.0.4-linux-amd64.tar.gz
解压
===
    tar -xvzf goFile-v1.0.4-linux-amd64.tar.gz
运行
===
    ./goFile

参数
===
    -path
目录，默认为./

    -port
web端口，默认为8089