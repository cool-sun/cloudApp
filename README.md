# 本项目开发者决定去创业了，短期内可能不会更新了，有感兴趣想接着开发的可以通过底部的QQ联系我！
# 本项目开发者决定去创业了，短期内可能不会更新了，有感兴趣想接着开发的可以通过底部的QQ联系我！
# 本项目开发者决定去创业了，短期内可能不会更新了，有感兴趣想接着开发的可以通过底部的QQ联系我！

# cloudApp 基于 K8S 的云应用平台

> cloud-app是基于k8s的简单易用的云应用平台，借助它可以快速安装部署各种应用，或者一键安装helm chart包，大幅提升k8s应用部署的效率。注意cloud-app的定位是应用的安装部署等操作，不是k8s的管理界面。

### 这里是后端代码，前端代码见这 [前端代码](https://gitee.com/coolsun972/cloudApp-front)

### 预览地址

[cloud-app](http://4kfox.com:8090/index)  用户名:`cloud-app` 密码`123456`

*偶尔访问不通可能是因为服务器在重启￣□￣｜｜*

### 功能特性

- 官方应用:

> 1.在 应用商店>官方应用 中，点击创建按钮，选择版本(版本号全部来自dockerhub)，选择CPU大小,内存大小,磁盘大小,存储类型。输入必要的PASSWORD等环境变量，便可快速创建应用。

> 2.在 应用管理>官方应用 中，可以对刚才创建的应用进行重启，创建数据快照，从快照恢复数据(不怕删库跑路了)，升级版本等一系列操作。

- Helm官方仓库chart包安装:

> 1.应用商店>Helm仓库 中，对接到Helm官方仓库，可以在Helm详情页点击 自动安装 按钮 一键安装chart包，安装前还可以可视化编辑Helm安装使用的values

> 2.应用商店>Helm仓库中的所有chart包都能安装，这可能是安装Helm包最简单的方法了！helm命令行工具可以卸载了😸

- 感觉我做的界面不好看？或者想要把它作为第三方接入？

> 参见接口文档自行开发。[文档地址](http://4kfox.com:8090/swagger/index.html)

### 技术栈

- 前端：Vue全家桶，AntD for Vue UI库。

  ~~Helm仓库部分原来使用是React全家桶开发，现已弃用。改成通过代理在Helm官方网站注入脚本的方式。~~
- 后端：Gin框架,Xorm操作数据库,k8s crd 和 operator开发

### 本地开发

1. > 克隆后端代码 git clone https://github.com/cool-sun/cloudApp.git
2. > 设置环境变量参数
    - `mysql_user` : 数据库用户名(推荐使用root账户,程序启动时会自动创建使用的`cloud-app`库，并切换连接到该库)
    - `mysql_password` : 数据库密码
    - `mysql_host` : 数据库主机
    - `mysql_port` : 数据库端口(不设置的话则使用`3306`端口)
    - `kube_config`: 连接到k8s的config文件地址
    - `admin_password`:`cloud-app`账户的密码(程序启动时会创建用户名为`cloud-app`的账户,密码为该值,不设置的话则为`123456`默认值)

   (上述操作做完后可以在本地开发运行了，不懂前端的忽略下一条)
3. > 克隆前端代码 git clone https://github.com/cool-sun/cloudApp-front.git

   > cd cloudApp-front && npm i && npm run server


### TODO

- [ ] 监控k8s事件，并通过websocket发给前端，以便用户更好的了解应用安装情况
- [ ] 首页显示当前用户的概况，包括当前app数量，helm release数量，和CPU,内存,磁盘当前占用量等
- [ ] 应用管理>官方应用 网络设置 添加黑白名单设置
- [ ] 应用管理>官方应用 添加应用详情(包括该app的详细信息，还有用e-chart展示历史的CPU和内存占用情况)
- [ ] 持续优化用户体验
- [ ] 。。。


### 效果图预览

![111.png](http://ww1.sinaimg.cn/large/0077OfRbly1gt4vwmwu0gj32hi1i01bf.jpg)
![222.png](http://ww1.sinaimg.cn/large/0077OfRbly1gt4vwmtci8j32he1hs15k.jpg)
![333.png](http://ww1.sinaimg.cn/large/0077OfRbly1gt4vwmvlp4j32la1hw18v.jpg)
![444.png](http://ww1.sinaimg.cn/large/0077OfRbly1gt4vwmx1quj32ke1hw7qz.jpg)
![555.png](http://ww1.sinaimg.cn/large/0077OfRbly1gt4vwmwydjj32ke1huaua.jpg)
![api.png](http://ww1.sinaimg.cn/large/0077OfRbly1gt4w1fknx7j31ur336e81.jpg)

### 最后如果你在使用`cloud-app`中遇到了问题，可以加群反馈。或者你也可以加群跟大家一起探讨云原生技术相关的问题
![qq.jpg](http://ww1.sinaimg.cn/large/0077OfRbly1gt5sol9hpyj30u00z0acd.jpg )

