# Lucky(大吉)
 
<!-- TOC -->
- [Lucky(大吉)](#)
  - [特性](#特性)
  - [一键安装](#一键安装)
  - [使用](#使用)
  - [Docker中使用](#docker中使用)
  - [后台界面](#后台界面)

  - [开发编译](#开发编译)
  - [更新日志](#更新日志)
  - [使用注意与常见问题](#使用注意与常见问题)
 
<!-- /TOC -->


## 特性

- 这是一个自用的,目前主要运行在自己的主路由(小米ax6000)里面的程序.

    - 后端golang,前端vue3
    - 支持Windows、Linux系统，支持x86、ARM、MIPS、MIPSLE等架构

- 目前已经实现的功能有
    - 1.替代socat,主要用于公网IPv6 tcp/udp转 内网ipv4
        - 支持界面化(web后台)管理转发规则,单条转发规则支持设置多个转发端口,一键开关指定转发规则
        - 单条规则支持黑白名单安全模式切换,白名单模式可以让没有安全验证的内网服务端口稍微安全一丢丢暴露到公网
        - Web后台支持查看最新100条日志
        - 另有精简版不带后台,支持命令行快捷设置转发规则,有利于空间有限的嵌入式设备运行.(不再提供编译版本,如有需求可以自己编译)
    - 2.动态域名服务
        - 参考和部分代码来自 https://github.com/jeessy2/ddns-go
        - 在ddns-go的基础上主要改进/增加的功能有
            - 1.同时支持接入多个不同的DNS服务商
            - 2.支持http/https/socks5代理设置
            - 3.自定义(Callback)和Webhook支持自定义headers
            - 4.支持BasicAuth
            - 5.DDNS任务列表即可了解全部信息(包含错误信息),无需单独查看日志.
            - 6.调用DNS服务商接口更新域名信息前可以先通过DNS解析域名比较IP,减少对服务商接口调用.
            - 其它细节功能自己慢慢发现...
            - 没有文档,后台各处的提示信息已经足够多.
            - 支持的DNS服务商和DDNS-GO一样,有Alidns(阿里云),百度云,Cloudflare,Dnspod(腾讯云),华为云.自定义(Callback)内置有每步,No-IP,Dynv6,Dynu模版,一键填充,仅需修改相应用户密码或者token即可快速接入.
    - 3.http/https反向代理
        - 特点
            - SSL证书管理简单
            - 设置简单
            - 支持HttpBasic认证  
            - 支持IP黑白名单
            - 支持UserAgent黑白名单
            - 日志记录最近访问情况
            - 一键开关子规则
            - 前端域名与后端地址 支持一对一,一对多(均衡负载),多对多(下一级反向代理)

- 将要实现的功能
    - 有建议可联系作者.



## 一键安装

- [一键安装详看这里](https://github.com/gdy666/lucky-files)


## 使用


- [百度网盘下载地址](https://pan.baidu.com/s/1NfumD9XjYU3OTeVmbu6vOQ?pwd=6666)
    百度网盘版本可能会更新比较频繁,
    

- 默认后台管理地址 http://<运行设备IP>:16601
  默认登录账号: 666
  默认登录密码: 666

- 常规使用请用 -c <配置文件路径> 指定配置文件的路由方式运行 , -p <后台端口> 可以指定后台管理端口
    ```bash
    #仅指定配置文件路径(如果配置文件不存在会自动创建),建议使用绝对路径
    lucky -c 666.conf
    #同时指定后台端口 8899
    lucky -c 666.conf -p 8899
    ```

- 命令行直接运行转发规则,注意后台无法编辑修改命令行启动的转发规则,主要用在不带后台的精简版
    ```bash
    #指定后台端口8899 
    lucky -p 8899 <转发规则1> <转发规则2> <转发规则3>...<<转发规则N>
    ```


## Docker中使用

- 不挂载主机目录, 删除容器同时会删除配置

  ```bash
  # host模式, 同时支持IPv4/IPv6, Liunx系统推荐
  docker run -d --name lucky --restart=always --net=host gdy666/lucky
  # 桥接模式, 只支持IPv4, Mac/Windows推荐,windows另外会有专用版本支持ipv6,待开发
  docker run -d --name lucky --restart=always -p 16601:16601 gdy666/lucky
  ```

- 在浏览器中打开`http://主机IP:16601`，修改你的配置，成功
- [可选] 挂载主机目录, 删除容器后配置不会丢失。可替换 `/root/luckyconf` 为主机目录, 配置文件为lucky.conf

  ```bash
  docker run -d --name lucky --restart=always --net=host -v /root/luckyconf:/goodluck gdy666/lucky
  ```












## 后台界面
![规则设置](./previews/relayruleset.png)
![规则列表](./previews/relayrules.png)
![](./previews/whitelistset.png)
![](./previews/whitelist.png)
#### 动态域名服务

![](./previews/ddnslist.png)


![](./previews/iphistroy.png)

![](./previews/webhookhistroy.png)

![](./previews/domainsync.png)

#### Http反向代理
![](./previews/reverseproxy.png)





#开发编译


    ```bash
    go build -v -tags "adminweb nomsgpack" -ldflags="-s -w"
    ```


#更新日志
2022-10-08
发布 1.4.9 版本
1.反向代理新增https支持。
2.加入SSL证书管理模块
3.后台接口支持HTTPS
4.端口转发模块重构优化，移除命令行配置功能,注意使用此版本后原来的端口转发配置会全部消失。
5.修复已知BUG.
6.源码不再和二进制版本同时发布.




## 使用注意与常见问题

 - 不同于防火墙端口转发规则,不要设置没有用上的端口,会增加内存的使用.

 - 小米路由 ipv4 类型的80和443端口被占用,但只设置监听tcp6(ipv6)的80/443端口转发规则完全没问题.

 - 如果需要使用白名单模式,请根据自身需求打开外网访问后台管理页面开关.

 - 转发规则启用异常,端口转发没有生效时请登录后台查看日志.

 - 开启外网访问可以直接修改配置文件中的"AllowInternetaccess": false, 将false改为true

