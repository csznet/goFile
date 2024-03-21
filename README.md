# goFile
easy file manager

我希望goFile是在运维时提供便利的工具，而不是大而全的文件管理器  

![image](https://github.com/csznet/goFile/assets/127601663/eec3f999-50c5-4119-86ff-c3eceb8b3b98)

<img width="1393" alt="image" src="https://user-images.githubusercontent.com/127601663/227174830-d5747bf9-6210-4fd4-b227-a154db494f11.png">

介绍
===

为了方便使用Caddy使用写的小东西  
简单的在线文件管理器，可以指定目录，指定端口，即用即开  
可以自定义前端HTML代码，只需要修改templates目录下的文件即可  
目前实现的功能： 

 - <del>后台远程下载</del>
 - 上传文件&拖放上传
 - 删除文件&文件夹
 - 新建文件&文件夹
 - 解压ZIP、gz压缩包
 - 在线编辑文件
 - 设备自适应明亮主题
 - 多语言支持（挖坑
 - 只读模式（也可以理解为阅读模式
 - <del>图片缩略图(鸡肋</del>

一键脚本
===

    bash <(curl -s https://raw.githubusercontent.com/csznet/goFile/main/goFile.sh)

一键脚本支持amd64、arm构架，Linux、MacOS系统  
Windows系统不会考虑（Windows就不需要去网页管理文件了吧

运行
===
如果是下载的二进制文件，则为

    ./goFile

如果使用的是一键脚本，则在需要开启goFile服务的文件夹中执行

    goFile


参数
===
    -path

目录，默认为./（一键脚本则为执行`goFile`命令的目录）

    -port

web端口，默认为8089

    -r

带入`-r`参数后表示为阅读模式，只能查看列表和下载文件，后面不需要带值  

带入`-t`参数代表打开缩略图（太鸡肋了，后面可能会删掉
