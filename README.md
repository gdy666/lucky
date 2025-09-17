# Lucky(万吉)
 
 本项目 CDN 加速及安全防护由 Tencent EdgeOne 赞助
 [亚洲最佳CDN、边缘和安全解决方案 - Tencent EdgeOne](https://edgeone.ai/zh?from=github)

 ![](https://edgeone.ai/media/34fe3a45-492d-4ea4-ae5d-ea1087ca7b4b.png)
 


 ## 注意：源码公布到1.4.10版本，后续暂无继续开源计划。

 ## 麻烦各位大佬发表lucky相关教程的时候不要加上“开源”神器，开源二字我不配，lucky后续也没开源打算。
        1.开源并不等于安全，闭源并不等于不安全。闭源软件开发也会受到安全人员的审查。无论是开源还是闭源软件，都有可能会受到各种安全人员的审查和研究。安全人员可以使用各种技术手段来检测软件的安全性和漏洞。
        2. 个人观点lucky这种应用类软件更多只是体力活，毫无技术含量，开源的优势在于透明度和社区参与，更多劳动力参与，但也可能导致功能过多、复杂度增加的问题。闭源软件的优势在于我想怎么写就怎么写,即使还未能从lucky中获利，lucky对我也有更深的特殊含义。
        3. 我对lucky的规划还有一大部分未实现，不想被人当免费劳动力使唤，不解释太多，就这样。

 
 ## 如果您是第一次使用Lucky，请务必先访问 https://lucky666.cn ，并仔细阅读相关的文档，以获得必要的信息和答案。在这些文档中，您可以了解到Lucky的基本功能和特性，掌握Lucky的使用方法，以及解决常见的问题和疑惑。
 

<!-- TOC -->
- [Lucky(万吉)](#)
  - [特性](#特性)
  - [一键安装](#一键安装)
  - [OpenwrtIPK包安装](#OpenwrtIPK包安装)
  - [使用](#使用)
  - [Docker中使用](#docker中使用)
  - [后台界面](#后台界面)

  - [开发编译](#开发编译)
  - [更新日志](#更新日志)
  - [使用注意与常见问题](#使用注意与常见问题)

<!-- /TOC -->


## 特性

Lucky最初是作为一个小工具，由开发者为自己的个人使用而开发，用于替代socat，在小米路由AX6000官方系统上实现公网IPv6转内网IPv4的功能。Lucky的设计始终致力于让更多的Linux嵌入式设备运行，以实现或集成个人用户常用功能，降低用户的硬件和软件操作学习成本，同时引导使用者注意网络安全。随着版本更新和网友反馈，Lucky不断迭代改进，拥有更多功能和更好的性能，成为用户值得信赖的工具。

Lucky 的核心程序完全采用 Golang 实现，具有高效、稳定、跨平台等优点。其后台前端则采用 Vue3.2 技术进行开发，具有良好的用户体验和响应速度。此外，Lucky 的管理后台采用前后端分离的架构，第三方开发者也可以自由使用OpenToken轻松调用Lucky的各种功能接口。



## 功能模块

目前已经实现/集成的主要功能模块有
  - 端口转发
  - 动态域名(DDNS)
  - Web服务
  - Stun内网穿透
  - 网络唤醒
  - 计划任务
  - ACME自动证书
  - 网络存储



### 端口转发
  1. 主要用于实现公网 IPv6 转内网 IPv4 的 TCP/UDP 端口转发。
  2. 支持界面化的管理转发规则，用户可以通过 web 后台轻松地进行规则的添加、删除、修改等操作。
  3. 单条转发规则支持设置多个转发端口，这样可以实现多个内网服务端口的转发。
  4. 提供了一键开关和定时开关功能，用户可以根据自己的需求设置转发规则的开启和关闭时间，还可以使用计划任务模块进行定时开关。
  5. 单条规则支持黑白名单安全模式切换，用户可以根据需要选择使用白名单模式或黑名单模式。
  6. 白名单模式可以让没有安全验证的内网服务端口稍微安全一点暴露到公网，提高服务可用性。
  7. 实时记录最新的访问日志，方便用户了解转发情况。
  8. 规则列表日志一目了然，用户可以方便地追踪转发异常，及时进行排查和处理。



### 动态域名(DDNS)
  1. 支持接入多个不同的 DNS 服务商。
  2. 支持全功能自定义回调（Callback），包括设置 BasicAuth，方便接入任意 DNS 服务商。
  3. Webhook 支持自定义 headers。
  4. 内置常用免费 DNS 服务商设置模板（每步、No-IP、Dynv6、Dynu），通过自定义回调进行快速接入，仅需修改相应用户密码或 token 即可一键填充。
  5. 支持 阿里云，百度云，华为云，京东云，腾讯云，火山引擎，帝恩爱斯-DNS.LA,Cloudflare，deSEC,DNSPod.CN，DNSPod.COM，Dynadot，Dynv6，Freemyip ,GoDaddy，Name.com，NameSilo,Porkbun，Vercel等服务商。


### Web服务
  1. 支持反向代理、重定向和 URL 跳转。
  2. 支持 HTTP 基本认证。
  3. 支持 IP 黑白名单模式。
  4. 支持 UserAgent 黑白名单。
  5. 规则日志清晰易懂，便于追踪异常。
  6. 支持一键开关规则和定时开关规则。


### Stun内网穿透
  1. 实现内网穿透，无需公网IPv4地址。
  2. 适合于国内运营商级NAT1宽带网络. 

### 网络唤醒
  1. 支持远程控制唤醒和关机操作
  2. 支持接入第三方物联网平台(点灯科技 巴法云),可通过各大平台的语音助手控制设备唤醒和关机.

### 计划任务
  1. 不依赖 Linux 系统的 Cron，支持 Windows 系统。
  2. 操作简便，可视化编辑。
  3. 可操作控制 Lucky 框架内的其他模块开关。

###  ACME自动证书
  1. 支持 ACME 自动证书的申请和续签。
  2. 支持 阿里云，百度云，华为云，京东云，腾讯云，火山引擎，帝恩爱斯-DNS.LA,Cloudflare，deSEC,DNSPod.CN，DNSPod.COM，Dynadot，Dynv6，Freemyip ,GoDaddy，Name.com，NameSilo,Porkbun，Vercel等服务商.


### 网络存储
  1. 网络存储模块是一个应用范围广泛的模块，它提供了将本地存储、WebDAV和阿里云盘挂载到Lucky内部的各个文件类服务功能。
  2. 通过网络存储模块，你可以将添加的存储挂载到Web服务的文件服务、WebDAV、FTP和FileBrowser模块，实现更加便捷的文件管理和访问。





## 一键安装

- [一键安装详看这里](https://github.com/gdy666/lucky-files)


## OpenwrtIPK包安装

- [Openwrt IPK包](https://github.com/gdy666/luci-app-lucky)


## 使用
    

- 默认后台管理地址 http://<运行设备IP>:16601
  默认登录账号: 666
  默认登录密码: 666

- 常规使用请用 -cd <配置文件夹路径> 指定配置文件夹的方式运行 
    ```bash
    #仅指定配置文件夹路径(如果配置文件夹不存在会自动创建),建议使用绝对路径
    lucky -cd luckyconf

    ```




## Docker中使用

- 不挂载主机目录, 删除容器同时会删除配置

  ```bash
  # host模式, 同时支持IPv4/IPv6, Liunx系统推荐
  docker run -d --name lucky --restart=always --net=host gdy666/lucky
  # 桥接模式, 只支持IPv4, Mac/Windows推荐,windows 不推荐使用docker版本
  docker run -d --name lucky --restart=always -p 16601:16601 gdy666/lucky
  ```

- 在浏览器中打开`http://主机IP:16601`，修改你的配置，成功
- [可选] 挂载主机目录, 删除容器后配置不会丢失。可替换 `/root/luckyconf` 为主机目录, 配置文件夹为lucky

  ```bash
  docker run -d --name lucky --restart=always --net=host -v /root/luckyconf:/goodluck gdy666/lucky
  ```


## 宝塔Docker安装

1.  安装宝塔面板 (9.2.0版本及以上)，前往 [宝塔面板](https://www.bt.cn/new/download.html) 官网，选择正式版的脚本下载安装
2.  安装后登录宝塔面板，在菜单栏中点击 Docker ，首次进入会提示安装 Docker 服务，点击立即安装，按提示完成安装
3.  安装完成后在应用商店中找到 lucky ，点击安装，配置基本选项 即可完成安装










#开发编译


    ```bash
    go build -v -tags "adminweb nomsgpack" -ldflags="-s -w"
    ```


# 更新日志

    2025-09-15 Lucky v2.19.4 
    1. FileBrowser 升级至 v2.43.0。
    2. Cloudflared 同步至最新官方源码。
    3. rclone 同步至最新官方源码。
    4. Web 服务测速优化
        HTTPS 下载测速请求尝试禁用 TCP 连接复用。
        因为多tcp连接可以充分利用多线程 TLS 加解密处理，在硬件性能支持的情况下，提升内网 HTTPS 测速下载接近带宽极限的能力。
    5. Cloudflare Tunnel 
        使用 Lucky 全局 DNS 设置解析 Edge 节点 IP，避免局域网默认 DNS 导致的解析污染问题。


2025-09-13 v2.19.3
    1. 修复前端问题
    2. Web服务：修复了在Web服务中的HTTPS规则，当默认规则设置为关闭连接时，如果子规则的前端为纯IP，未能正确识别LocalIP导致访问连接被关闭的问题。
    3. 在线测速优化：在线测速功能改为使用WebSocket发送ping pong方式计算ping时间，以获取更加贴近真实网络的ping值。在线测前端面板不再兼容原版Homebox，默认内置测速前端面板源改为https://cdn.66666.host/homebox/ ，用户可以自行构建面板源（https://github.com/gdy666/homebox）。

2025-09-12 v2.19.2
    1. Web 服务子规则类型 “测速后端” 更名为 “在线测速”。
    直接访问子规则即可完成测速，无需额外配置。
    与原版 Homebox 使用体验保持一致，如需测试内网速度，只需在前端地址中增加一条内网 IP 即可。
    2. 新增 背景模糊度设置，
    3. 优化若干前端细节
    4. 修复了使用第三方图片作为背景时可能通过 referer 头暴露后台域名和端口的风险。


2025-09-09 Lucky v2.19.1 
    1. IP 地址库
    1.1 集成最新 ip2region 查询客户端，支持 IPv6。
    ip2region 官方提供的 IPv6 数据库信息存在不准确以及文件体积过大的情况，不推荐使用。
    如须体验，不建议使用缓存模式，在 Linux 系统下，缓存整个 IPv6 数据文件可能导致内存占用过高触发 OOM。
    1.2 上传 IP 地址库文件时，新增上传进度显示
    2.DDNS/acme
    新增 DuckDNS 支持 （不建议使用，如遇问题无须反馈）
    3.Web服务
    修复 SNI 功能在处理非英文域名（如中文域名）时的识别问题
    4.移动端界面适配优化。
    本版本支持 Windows 7 与 Windows Server 2008。

2025-09-07 v2.19.0 beta1
    1.Web服务：
    修复：调整 SNI 默认超时时间，解决 Lucky 总览页面频繁断开 WebSocket 连接的问题。
    新增：反向代理新增对 Nginx 变量 $ssl_client_cert 的兼容支持。
    2.前端新增暗黑模式
    3.lucky设置新增支持设置自定义背景图片或背景颜色
    注意：从本版本开始，不再支持 Windows 7 与 Windows Server 2008。

   [更多日志请查看](https://lucky666.cn/docs/category/%E6%9B%B4%E6%96%B0%E6%97%A5%E5%BF%97)

















。



