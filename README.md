# go-config

## 配置特点

>* **强环境相关**
>* **变更不频繁**
>* **快速响应变更推送**

### 强环境相关
    环境多是采用单独配置中心的主要原因，多环境切换采用传统方式过于痛苦，容易出错。
    环境包括：功能1测试服，功能2测试服，联调测试服，预发布服1，预发布服2，压测服，正式服。。。
    环境相关的一些例子：

>* 日志级别
>* 中间件配置
>* 某些环境相关开关
>* 一些池子容量，性能相关等
>* 其它因为环境不同需要改变的

### 变更不频繁
    一般配置项都不会频繁变动。
    
### 快速响应变更推送
    配置变动需要尽快通知到所有相关的服务。
    至少要满足特殊配置需满足快速响应变更推送。
    
## 配置项格式
```text
/data/config/{env}/{level}/{project}/{service}/{node}/{key}/{subkey}
```
    env 环境
    level 配置等级，
        common      公共配置
        area        机房配置
        private     node 私有配置
        例如公共配置，私有配置，当私有配置不存在时，再获取机房配置。
        可优化减少重复配置项
    project 项目，产品线
    service 服务
    node 服务唯一编号
    key 主键，由应用服配置
    subkey 子键，优化用，提高多配置项下的整体性能
    
## 整体考虑

### 高可用

    缓存+入库或者写磁盘记录配置项。避免配置中心宕机，依赖的服务不可用。
    
### 高性能

    **轮询机制** **快速响应变更推送** 需依靠轮询间隔来控制
    **长连接机制** 能满足 **快速响应变更推送**

### 方案对比

1. 配置服负责配置修改接口。其它服直连 consul 轮询配置项。
2. 配置服负责配置修改接口。配置服直连 consul 轮询配置项。
其它服通过消息队列订阅配置服相关配置项。
3. 配置服负责配置修改接口。配置服直连 consul 采用 Watch 机制监控配置项。
其它服通过消息队列订阅配置服相关配置项。
4. 实时性高的配置采用 consul Watch 机制监控。其它非实时采用轮询机制。

### 整体实现

>* 高可用方案 -> 缓存+入库或者写磁盘记录配置项
>* 高性能优化 -> 添加 subkey ,降低监听项目项，避免多项目下，性能降低。
>* 业务优化 -> 快速拷贝不同环境项目
>* 版本管理
>* 服务治理相关单独，例如路由策略，灰度发布，主备切换等(依赖其它服务治理相关服规则)。