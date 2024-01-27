# dzhwxbot-go
## 微信机器人

### v0.0.2
- 每天凌晨6点，自动拉去最新的群列表，好友列表，公众号列表
- 完成群聊向机器人订阅天气和查询天气，每天机器人定时推送天气预告

### v0.0.1
- 修改响应关键词规则，支持小智+空格，支持小智+换行，两种格式触发
- 增加@机器人触发回复
- 增加拍一拍回复
- 天气回复
- 完成对接到chatgpt问答




Project Makefile Commands: 
- `make cli`: Install or Update to the latest GoFrame CLI tool.
- `make dao`: Generate go files for `Entity/DAO/DO` according to the configuration file from `hack` folder.
- `make service`: Parse `logic` folder to generate interface go files into `service` folder.
- `make image TAG=xxx`: Run `docker build` to build image according `manifest/docker`.
- `make image.push TAG=xxx`: Run `docker build` and `docker push` to build and push image according `manifest/docker`.
- `make deploy TAG=xxx`: Run `kustomize build` to build and deploy deployment to kubernetes server group according `manifest/deploy`.