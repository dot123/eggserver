# eggserver百人大逃杀游戏服务器

## 客户端：https://github.com/dot123/eggroyale

## 概要
- mysql数据落地，orm映射
- 使用redis实现数据共享
- 使用gin框架开发api
- 使用luban导出配置表数据
- 使用Aws Lambda实现高可用

### 服务端使用简要介绍
- ./install dependencies.cmd 安装工具包
- ./build_swag.cmd 生成api接口doc
- ./configs_gen.cmd 导出配置表文件
- ./robot/run.cmd 压测工具
- ./tools/error-gen/build.cmd 生成错误码文件到./internal/constant/code.go
- ./data/conf/ mysql redis等配置信息
- ./build.sh 打包成bin.zip

## 客户端截图
### 主界面
![主界面](./screenshot/01.png)

### 宠物界面
![宠物界面](./screenshot/02.png)

### 任务界面
![任务界面](./screenshot/03.png)

### 商店界面
![商店界面](./screenshot/04.png)

### 签到界面
![签到界面](./screenshot/05.png)

### 战令界面
![战令界面](./screenshot/06.png)

### 匹配界面
![宠物界面](./screenshot/07.png)

### 战斗界面
![战斗界面](./screenshot/08.png)
