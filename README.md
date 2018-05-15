# 简介
《GOS》是一款基于Go语言的分布式游戏服务器框架，高可用、动态伸缩、在线扩容，可应用于市面上绝大多数游戏类型：SLG、RPG、ARPG、MMO、MOBA。

## 结构图
![Architecture](Architecture.png)

## 单点消除
![SinglePoints](SinglePoints.png)
  
## 结构详解
  * AuthService
    
    验证服务，提供账户注册、玩家登陆，以及为玩家获取连接服务(AgentCell)信息的功能。
  * AgentMgr
  
    连接服务管理器，连接服务负载均衡管理，根据当前负载情况为玩家分配的连接服务。向集群管理器申请开辟和释放连接服务。
  * GameMgr
  
    游戏服务管理器，游戏服务负载均衡管理，根据当前负载情况为玩家分配游戏服务。向集群管理器申请开辟和释放游戏服务。
  * Agent
  
    连接服务，转发玩家信息至游戏服务，转发游戏服务信息至玩家，处理游戏内广播消息。
  * Game
  
    游戏服务，游戏场景管理，处理玩家请求，游戏逻辑的主要发生地；加载、持久化场景、玩家信息至MySQL集群。
  * Scene
  
    游戏场景，每个游戏服务内包含一个或多个游戏场景，场景可以是MMO的每个场景地图，也可以是SLG的世界地图，还可以是MMO的大厅服务，甚至可以是一个游戏服。场景的大小粒度，可以根据游戏的实际情况而定，游戏场景概念是《GOS》进行负载分布的核心设计。
  * Hot Data
  
    热数据管理，由于游戏内玩家数据会频繁变更，所以游戏场景和玩家启动后会将其相关的数据加载到内存中，并由Hot Data进行管理，并定时的回写到MySQL集群。
  * MySQL Cluster
  
    数据库集群，主要用于保存玩家数据和游戏场景数据；由于当下云服务已经非常成熟和完善，这里建议直接使用云服务的RDS服务，在后台点点点就能创建一个MySQL集群，读写性能和数据安全性都有保证。
  * Redis Cluster
  
    Redis集群，主要用于集群配置信息保存，玩家Session缓存；Redis集群可以自己根据Redis官网搭建，同时也推荐大家使用云服务，简单快捷稳定，费用不高。
  

## 基础工具集
  * 分布式服务健康监测
  * GenServer：类似于Erlang的gen_server，封装了Groutine的基本启动、查找、消息同步/异步发送
  * 协议生成器：根据YAML生成客户端与服务器的通信协议
  * 路由管理：根据YAML文件生成路由协议，玩家请求自动分发至相应Controller进行处理
  * 热数据管理：按需加载玩家数据，并定时持久化至MySQL
  * MySQL管理：基于Rails的ActiveRecord进行migration管理，并生成Go的ORM文件

## Quick Start
```bash
git checkout https://github.com/mafei198/gos.git
make dep_install
make setup
make build
make start_all
```

## TODO
  * 定时任务 (done)
  * 配置数据生成工具 (done)
  * 排行榜服务 (done)
  * 聊天服务
  * 邮件服务
  * 玩家日志
  * 世界数据
  * 云服务器管理
  * 运行时REPL交互环境
  * 推送消息：苹果、谷歌
  * 支付验证：苹果、谷歌
  
  
##
以上是一个简要的框架设计介绍，算是个开始；《GOS》也在紧锣密鼓的逐渐完善中（当然是利用业余时间）。将思路记录在这里，一方面可以帮助自己缕清思路，另一方面监督自己持续完善，当然如果还能帮助到其他服务器研发人员就是Bonus了！欢饮各位批评指正，一同讨论~

## License
GOS is under The MIT License (MIT)

Copyright (c) 2018-2028
Savin Max <mafei.198@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

