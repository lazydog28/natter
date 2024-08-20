# Natter
> 软件实现了一个简陋的STUN内网穿透服务，能够在NAT1网络下将内网主机端口映射到公网端口上，实现内网穿透。
> 
> 因为我的需求只是将内网的网站暴露到公网，使用的TCP协议，所以软件只支持TCP协议，不支持UDP协议。


> [NAT网络类型检测](https://github.com/HMBSbige/NatTypeTester)
> 
> NAT1：Full Cone NAT（全锥形NAT）；
> 
> NAT2：Address Restricted Cone NAT（受限锥型NAT）；
> 
> NAT3：Port Restricted Cone NAT（端口受限锥型NAT）；
> 
> NAT4：Symmetric NAT（对称型NAT）；


## 缺点
- 仅支持TCP协议
- 仅支持NAT1网络

## 优点
- 轻量化，软件只有几兆大小
- 软件简单，只有一个可执行文件
- 功能纯粹，没有什么多余的功能

## 使用说明

下载地址：[Natter Releases](https://github.com/lazydog28/natter/releases)
下载对应平台的软件，在终端中运行即可。
```shell
# 查看帮助
.\natter_windows_amd64.exe -h

# 运行服务 比如你有一个网站运行在本地的5244端口，你想要将这个网站暴露到公网上，那么你可以运行下面的命令
.\natter_windows_amd64.exe -f 127.0.0.1:5244
```

![ad4df11ccc245df617632ed8de6cb1e1.png](https://imagesbed28.caiyun.fun/ad4df11ccc245df617632ed8de6cb1e1.png)

暴露的端口是随机的，你现在可以通过访问输出的公网访问地址来访问你的内网服务。
